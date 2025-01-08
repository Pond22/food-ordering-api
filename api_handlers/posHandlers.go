package api_handlers

import (
	"food-ordering-api/db"
	"food-ordering-api/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

// สร้าง session ใหม่
func StartPOSSession(c *fiber.Ctx) error {
	var req struct {
		StaffID uint `json:"staff_id"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// สร้าง session ใหม่
	session := models.POSSession{
		StaffID:    req.StaffID,
		StartTime:  time.Now(),
		Status:     "active",
		LoginToken: uuid.New().String(),
	}

	if err := db.DB.Create(&session).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create session",
		})
	}

	// โหลดข้อมูล staff
	if err := db.DB.Preload("Staff").First(&session, session.ID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to load session data",
		})
	}

	return c.JSON(session)
}

// จบ session
func EndPOSSession(c *fiber.Ctx) error {
	id := c.Params("id")
	now := time.Now()

	if err := db.DB.Model(&models.POSSession{}).
		Where("id = ? AND status = ?", id, "active").
		Updates(map[string]interface{}{
			"status":   "ended",
			"end_time": now,
		}).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to end session",
		})
	}

	return c.SendStatus(200)
}

// ตรวจสอบ session
func ValidatePOSSession(c *fiber.Ctx) error {
	id := c.Params("id")

	var session models.POSSession
	if err := db.DB.Where("id = ? AND status = ?", id, "active").
		First(&session).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid session",
		})
	}

	return c.SendStatus(200)
}