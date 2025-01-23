package api_handlers

import (
	"log"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// ตัวแปรเก็บการเชื่อมต่อของพนักงาน
var staffClients = make(map[*websocket.Conn]bool)
var mutex = sync.Mutex{}

// ฟังก์ชัน WebSocket Handler สำหรับพนักงาน
func StaffWebSocket(c *websocket.Conn) {
	mutex.Lock()
	staffClients[c] = true
	mutex.Unlock()

	defer func() {
		mutex.Lock()
		delete(staffClients, c)
		mutex.Unlock()
		c.Close()
	}()

	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			log.Println("Disconnected:", err)
			break
		}
	}
}

type CallRequest struct {
	TableNumber string `json:"table_number"`
}

// @Summary Call staff
// @Description ลูกค้าเรียกพนักงาน
// @Tags Notifications
// @Accept json
// @Produce json
// @Param status body CallRequest true "โต๊ะที่เรียก"
// @Success 200 {object} map[string]string "Notification sent to staff"
// @Failure 400 {object} map[string]string "Invalid request"
// @Router /api/notifications/call-staff [post]
func CallStaff(c *fiber.Ctx) error {

	var req CallRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request"})
	}

	// สร้างข้อความแจ้งเตือน
	message := "ลูกค้าจากโต๊ะ " + req.TableNumber + " ต้องการเรียกพนักงาน"

	// ส่งแจ้งเตือนไปยังพนักงานทุกคน
	NotifyStaff(message)

	return c.JSON(fiber.Map{"status": "Notification sent to staff"})
}

// ฟังก์ชันสำหรับกระจายแจ้งเตือนไปยังพนักงานทุกคน
func NotifyStaff(message string) {
	mutex.Lock()
	defer mutex.Unlock()

	for client := range staffClients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("Failed to send notification:", err)
			client.Close()
			delete(staffClients, client)
		}
	}
}
