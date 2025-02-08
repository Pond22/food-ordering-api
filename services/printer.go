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
	Type      string `json:"type"`           // "network" หรือ "usb"
	IP        string `json:"ip,omitempty"`   // สำหรับ network printer
	Port      int    `json:"port,omitempty"` // เปลี่ยนจาก *int เป็น int
	Name      string `json:"name"`
	PaperSize string `json:"paper_size"`
	Status    string `json:"status"`
	LastSeen  string `json:"last_seen"`
	VendorID  string `json:"vendor_id,omitempty"`  // สำหรับ USB printer
	ProductID string `json:"product_id,omitempty"` // สำหรับ USB printer
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
		if printer.Type == "network" {
			conn.Printers[printer.IP] = true
		} else if printer.Type == "usb" {
			// สำหรับ USB printer ใช้ vendor_id และ product_id เป็น key
			key := "USB_" + printer.VendorID + "_" + printer.ProductID
			conn.Printers[key] = true
		}
	}
	connMutex.Unlock()

	updatePrinters(msg.Printers)
}

func updatePrinters(printers []PrinterInfo) {
	for _, p := range printers {
		lastSeen, _ := time.Parse("2006-01-02 15:04:05", p.LastSeen)

		var printer models.Printer
		var result error

		if p.Type == "usb" {
			result = db.DB.Where("type = ? AND vendor_id = ? AND product_id = ?",
				"usb", p.VendorID, p.ProductID).First(&printer).Error

			log.Printf("Looking for USB printer: VID=%s, PID=%s", p.VendorID, p.ProductID)
		} else {
			result = db.DB.Where("type = ? AND ip_address = ?",
				"network", p.IP).First(&printer).Error

			log.Printf("Looking for Network printer: IP=%s", p.IP)
		}

		if result != nil {
			// สร้างเครื่องพิมพ์ใหม่
			printer = models.Printer{
				Type:      p.Type,
				Name:      p.Name,
				PaperSize: p.PaperSize,
				Status:    p.Status,
				LastSeen:  lastSeen,
			}

			if p.Type == "usb" {
				printer.VendorID = p.VendorID
				printer.ProductID = p.ProductID
				log.Printf("Creating new USB printer: VID=%s, PID=%s", p.VendorID, p.ProductID)
			} else {
				printer.IPAddress = p.IP
				printer.Port = p.Port // ใช้ค่า int โดยตรง
				log.Printf("Creating new Network printer: IP=%s", p.IP)
			}

			if err := db.DB.Create(&printer).Error; err != nil {
				log.Printf("Error creating printer: %v", err)
				continue
			}
		} else {
			// อัพเดทข้อมูลที่มีอยู่
			updates := map[string]interface{}{
				"name":       p.Name,
				"paper_size": p.PaperSize,
				"status":     p.Status,
				"last_seen":  lastSeen,
			}

			if p.Type == "usb" {
				updates["vendor_id"] = p.VendorID
				updates["product_id"] = p.ProductID
				log.Printf("Updating USB printer: VID=%s, PID=%s", p.VendorID, p.ProductID)
			} else {
				updates["port"] = p.Port // ใช้ค่า int โดยตรง
				log.Printf("Updating Network printer: IP=%s", p.IP)
			}

			if err := db.DB.Model(&printer).Updates(updates).Error; err != nil {
				log.Printf("Error updating printer: %v", err)
				continue
			}
		}
	}
}
