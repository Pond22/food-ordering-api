package service

import (
	"encoding/json"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"log"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

type PrinterConnection struct {
	Conn     *websocket.Conn
	StoreID  string
	Printers map[string]bool
}

type PrinterInfo struct {
	IP         string `json:"ip"`
	Port       int    `json:"port"`
	Name       string `json:"name"`
	Department string `json:"department"`
	Status     string `json:"status"`
	LastSeen   string `json:"last_seen"`
}

type ClientMessage struct {
	Type     string        `json:"type"`
	StoreID  string        `json:"store_id,omitempty"`
	Printers []PrinterInfo `json:"printers,omitempty"`
}

type PrintStatusMessage struct {
	Type   string `json:"type"`
	JobID  uint   `json:"job_id"`
	Status string `json:"status"`
}

var (
	activeConnections = make(map[*websocket.Conn]*PrinterConnection)
	connMutex         = &sync.RWMutex{}
)

func HandlePrinterWebSocket(c *fiber.Ctx) error {
	if websocket.IsWebSocketUpgrade(c) {
		c.Locals("allowed", true)
		return websocket.New(handleWebSocket)(c)
	}
	return fiber.ErrUpgradeRequired
}

func handleWebSocket(c *websocket.Conn) {
	log.Printf("New client connected")

	defer func() {
		connMutex.Lock()
		delete(activeConnections, c)
		connMutex.Unlock()
		c.Close()
		log.Printf("Client disconnected")
	}()

	connMutex.Lock()
	activeConnections[c] = &PrinterConnection{
		Conn:     c,
		Printers: make(map[string]bool),
	}
	connMutex.Unlock()

	for {
		messageType, message, err := c.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}

		if messageType == websocket.TextMessage {
			// ตรวจสอบประเภทของ message
			var baseMsg struct {
				Type string `json:"type"`
			}
			if err := json.Unmarshal(message, &baseMsg); err != nil {
				log.Printf("Error unmarshaling base message: %v", err)
				continue
			}

			switch baseMsg.Type {
			case "update_printers":
				var msg ClientMessage
				if err := json.Unmarshal(message, &msg); err != nil {
					log.Printf("Error unmarshaling printer message: %v", err)
					continue
				}
				handlePrinterUpdate(c, msg)

			case "update_print_status":
				var msg PrintStatusMessage
				if err := json.Unmarshal(message, &msg); err != nil {
					log.Printf("Error unmarshaling status message: %v", err)
					continue
				}
				handlePrintStatus(msg)
			}
		}
	}
}

func handlePrintStatus(msg PrintStatusMessage) {
	// อัพเดทสถานะงานพิมพ์ในฐานข้อมูล
	var job models.PrintJob
	if err := db.DB.First(&job, msg.JobID).Error; err != nil {
		log.Printf("Error finding print job %d: %v", msg.JobID, err)
		return
	}

	// อัพเดทสถานะและเวลา
	updates := map[string]interface{}{
		"status":     msg.Status,
		"updated_at": time.Now(),
	}

	if err := db.DB.Model(&job).Updates(updates).Error; err != nil {
		log.Printf("Error updating print job status: %v", err)
		return
	}

	log.Printf("Updated print job %d status to %s", msg.JobID, msg.Status)
}

func handlePrinterUpdate(c *websocket.Conn, msg ClientMessage) {
	connMutex.Lock()
	conn := activeConnections[c]
	conn.Printers = make(map[string]bool)

	for _, printer := range msg.Printers {
		conn.Printers[printer.IP] = true
	}
	connMutex.Unlock()

	updatePrinters(msg.Printers)
}

func updatePrinters(printers []PrinterInfo) {
	for _, p := range printers {
		lastSeen, _ := time.Parse("2006-01-02 15:04:05", p.LastSeen)

		var printer models.Printer
		result := db.DB.Where("ip_address = ?", p.IP).First(&printer)

		if result.Error != nil {
			printer = models.Printer{
				Name:       p.Name,
				IPAddress:  p.IP,
				Port:       p.Port,
				Department: p.Department,
				Status:     p.Status,
				LastSeen:   lastSeen,
			}
			db.DB.Create(&printer)
		} else {
			db.DB.Model(&printer).Updates(map[string]interface{}{
				"name":       p.Name,
				"department": p.Department,
				"status":     p.Status,
				"last_seen":  lastSeen,
			})
		}
	}
}