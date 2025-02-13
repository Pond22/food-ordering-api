package api_handlers

import (
	"food-ordering-api/db"
	"food-ordering-api/models"
	"time"

	"encoding/json"

	"regexp"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
)

// สำหรับรับข้อมูลการเรียกพนักงาน
type CallRequest struct {
	TableID string `json:"table_id"`
	Type    string `json:"type"` // "call_staff" หรือ "payment"
}

// เพิ่ม struct สำหรับรับข้อมูลการรับทราบ
type AcknowledgeRequest struct {
	NotificationID string `json:"notification_id"`
	StaffID        uint   `json:"staff_id"`
}

// สำหรับส่งผ่าน WebSocket
type WSNotification struct {
	ID        uint      `json:"id"`
	Type      string    `json:"type"`
	TableID   string    `json:"table_id"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	StaffID   *uint     `json:"staff_id,omitempty"` // เพิ่ม field นี้
}

var (
	staffConnections = make(map[*websocket.Conn]bool)
	connectionMutex  = sync.Mutex{}
)

// @Summary เรียกพนักงาน
// @Tags Notifications
// @Accept json
// @Produce json
// @Param request body CallRequest true "ข้อมูลการเรียก"
// @Success 200 {object} models.Notification
// @Router /api/notifications/call [post]
func HandleCall(c *fiber.Ctx) error {
	var req CallRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request format"})
	}

	// ตรวจสอบข้อมูลที่จำเป็น
	if req.TableID == "" || (req.Type != "call_staff" && req.Type != "payment") {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request data"})
	}

	// ตรวจสอบการแจ้งเตือนที่ยังไม่ได้อ่าน
	var existingNotification models.Notification
	if err := db.DB.Where(
		"message LIKE ? AND status = ? AND created_at > ?",
		"โต๊ะ "+req.TableID+"%",
		"unread",
		time.Now().Add(-15*time.Minute),
	).First(&existingNotification).Error; err == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "มีการแจ้งเตือนที่ยังไม่ได้รับการตอบรับอยู่แล้ว",
		})
	}

	message := "โต๊ะ " + req.TableID
	if req.Type == "payment" {
		message += " ต้องการชำระเงิน"
	} else {
		message += " ต้องการเรียกพนักงาน"
	}

	notification := models.Notification{
		UserID:    1,
		Message:   message,
		Status:    "unread",
		CreatedAt: time.Now(),
	}

	if err := db.DB.Create(&notification).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create notification"})
	}

	// ส่งการแจ้งเตือนผ่าน WebSocket
	wsNotification := WSNotification{
		ID:        notification.ID,
		Type:      req.Type,
		TableID:   req.TableID,
		Message:   message,
		Status:    "unread",
		CreatedAt: notification.CreatedAt,
	}

	broadcastToStaff(wsNotification)
	return c.JSON(notification)
}

// WebSocket handler
func HandleWebSocket(c *websocket.Conn) {
	// เพิ่ม connection ใหม่
	connectionMutex.Lock()
	staffConnections[c] = true
	connectionMutex.Unlock()

	// Cleanup เมื่อ function จบการทำงาน
	defer func() {
		connectionMutex.Lock()
		delete(staffConnections, c)
		connectionMutex.Unlock()
		c.Close()
	}()

	// ส่งการแจ้งเตือนที่ยังไม่ได้อ่านทั้งหมดเมื่อเชื่อมต่อครั้งแรก
	var unreadNotifications []models.Notification
	if err := db.DB.Where("status = ?", "unread").
		Order("created_at DESC").
		Find(&unreadNotifications).Error; err == nil {
		for _, notification := range unreadNotifications {
			wsNotification := WSNotification{
				ID:        notification.ID,
				Type:      getNotificationType(notification.Message),
				TableID:   getTableID(notification.Message),
				Message:   notification.Message,
				Status:    notification.Status,
				CreatedAt: notification.CreatedAt,
			}
			data, _ := json.Marshal(wsNotification)
			c.WriteMessage(websocket.TextMessage, data)
		}
	}

	// รอรับข้อความจาก client
	for {
		_, msg, err := c.ReadMessage()
		if err != nil {
			break
		}

		var ackReq AcknowledgeRequest
		if err := json.Unmarshal(msg, &ackReq); err != nil {
			continue
		}

		// จัดการการรับทราบการแจ้งเตือน
		if ackReq.NotificationID != "" && ackReq.StaffID != 0 {
			handleNotificationAcknowledgment(ackReq.NotificationID, ackReq.StaffID)
		}
	}
}

func broadcastToStaff(notification WSNotification) {
	data, err := json.Marshal(notification)
	if err != nil {
		return
	}

	connectionMutex.Lock()
	defer connectionMutex.Unlock()

	for conn := range staffConnections {
		if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
			delete(staffConnections, conn)
			conn.Close()
		}
	}
}

func handleNotificationAcknowledgment(notificationID string, staffID uint) {
	var notification models.Notification
	if err := db.DB.First(&notification, notificationID).Error; err != nil {
		return
	}

	now := time.Now()
	notification.Status = "read"
	notification.AcknowledgedBy = &staffID
	notification.AcknowledgedAt = &now

	if err := db.DB.Save(&notification).Error; err != nil {
		return
	}

	// แจ้งทุกคนว่าการแจ้งเตือนถูกรับทราบแล้ว
	wsNotification := WSNotification{
		ID:        notification.ID,
		Type:      getNotificationType(notification.Message),
		TableID:   getTableID(notification.Message),
		Message:   notification.Message,
		Status:    "read",
		CreatedAt: notification.CreatedAt,
		StaffID:   &staffID,
	}
	broadcastToStaff(wsNotification)
}

// @Summary อ่านการแจ้งเตือน
// @Tags Notifications
// @Accept json
// @Produce json
// @Param id path int true "Notification ID"
// @Success 200 {object} models.Notification
// @Router /api/notifications/{id}/read [post]
func MarkAsRead(c *fiber.Ctx) error {
	notificationID := c.Params("id")

	var notification models.Notification
	if err := db.DB.First(&notification, notificationID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Notification not found",
		})
	}

	notification.Status = "read"
	if err := db.DB.Save(&notification).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update notification",
		})
	}

	return c.JSON(notification)
}

// @Summary ดึงการแจ้งเตือนที่ยังไม่ได้อ่าน
// @Tags Notifications
// @Produce json
// @Success 200 {array} models.Notification
// @Router /api/notifications/unread [get]
func GetUnreadNotifications(c *fiber.Ctx) error {
	var notifications []models.Notification
	if err := db.DB.Where("status = ?", "unread").
		Order("created_at DESC").
		Find(&notifications).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch notifications",
		})
	}
	return c.JSON(notifications)
}

// Helper functions
func getNotificationType(message string) string {
	if strings.Contains(message, "ชำระเงิน") {
		return "payment"
	}
	return "call_staff"
}

func getTableID(message string) string {
	re := regexp.MustCompile(`โต๊ะ (\d+)`)
	matches := re.FindStringSubmatch(message)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
