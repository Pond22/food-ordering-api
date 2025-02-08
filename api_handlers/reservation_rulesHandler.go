package api_handlers

import (
	"errors"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type RuleRequest struct {
	GracePeriodMinutes    int `json:"grace_period_minutes" example:"15"`
	PreReservationMinutes int `json:"pre_reservation_minutes" example:"30"`
}

// @Summary ตั้งค่ากฎการจองโต๊ะ
// @Description กำหนดกฎการจองโต๊ะ เช่น เวลาสายที่ยอมรับได้ และเวลากั้นโต๊ะล่วงหน้า
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body RuleRequest true "กฎการจองโต๊ะ"
// @Success 200 {object} models.ReservationRules "กฎการจองที่ตั้งค่า"
// @Router /api/reservation/rules [post]
func SetReservationRules(c *fiber.Ctx) error {
	var req RuleRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	// Validation
	if req.GracePeriodMinutes < 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "เวลาสายต้องไม่ติดลบ",
		})
	}

	if req.PreReservationMinutes < 0 {
		return c.Status(400).JSON(fiber.Map{
			"error": "เวลากั้นโต๊ะล่วงหน้าต้องไม่ติดลบ",
		})
	}

	tx := db.DB.Begin()

	// ยกเลิก rule เก่า
	if err := tx.Model(&models.ReservationRules{}).
		Where("is_active = ?", true).
		Update("is_active", false).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอัพเดทกฎเก่าได้",
		})
	}

	// สร้าง rule ใหม่
	newRule := models.ReservationRules{
		GracePeriodMinutes:    req.GracePeriodMinutes,
		PreReservationMinutes: req.PreReservationMinutes,
		IsActive:              true,
	}

	if err := tx.Create(&newRule).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถสร้างกฎใหม่ได้",
		})
	}

	tx.Commit()
	return c.JSON(newRule)
}

// @Summary ดึงกฎการจองที่ใช้งานอยู่
// @Description ดึงกฎการจองที่ active อยู่ในปัจจุบัน ถ้าไม่มีจะส่งค่า default
// @Tags ReservationRules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} models.ReservationRules "กฎการจองที่ใช้งานอยู่"
// @Success 200 {object} map[string]interface{} "ค่า default เมื่อไม่มีกฎที่ active"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดภายในระบบ"
// @Router /api/reservation/rules/active [get]
func GetActiveReservationRule(c *fiber.Ctx) error {
	var activeRule models.ReservationRules
	if err := db.DB.Where("is_active = ?", true).First(&activeRule).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// ถ้าไม่พบ rule ที่ active ให้ส่งค่า default
			return c.JSON(fiber.Map{
				"grace_period_minutes":    15, // ค่าเริ่มต้น 15 นาที
				"pre_reservation_minutes": 30, // ค่าเริ่มต้น 30 นาที
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการดึงข้อมูล",
		})
	}

	return c.JSON(activeRule)
}

// @Summary ดึงประวัติการตั้งค่ากฎการจอง
// @Description ดึงประวัติการตั้งค่ากฎการจองทั้งหมด เรียงตามเวลาล่าสุด
// @Tags ReservationRules
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.ReservationRules "ประวัติการตั้งค่ากฎการจอง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดภายในระบบ"
// @Router /api/reservation/rules/history [get]
func GetReservationRulesHistory(c *fiber.Ctx) error {
	var rules []models.ReservationRules
	if err := db.DB.Order("created_at DESC").Find(&rules).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการดึงประวัติ",
		})
	}

	return c.JSON(rules)
}

func AutoManageReservations() {
	for {
		now := time.Now()

		// ค้นหาการจองที่เลยเวลา grace period
		var expiredReservations []models.TableReservation
		db.DB.Where("status = ? AND grace_period_until < ?", "active", now).Find(&expiredReservations)

		for _, reservation := range expiredReservations {
			tx := db.DB.Begin()

			// อัพเดทสถานะการจอง
			if err := tx.Model(&reservation).Update("status", "no_show").Error; err != nil {
				tx.Rollback()
				continue
			}

			// ปล่อยโต๊ะ
			if err := tx.Model(&models.Table{}).
				Where("id = ?", reservation.TableID).
				Update("status", "available").Error; err != nil {
				tx.Rollback()
				continue
			}

			tx.Commit()
		}

		time.Sleep(1 * time.Minute) // ตรวจสอบทุก 1 นาที
	}
}

func ManageTableReservationStatus() {
	for {
		now := time.Now()
		tx := db.DB.Begin()

		// อัพเดทสถานะโต๊ะเป็น reserved เมื่อถึงเวลา blocked
		if err := tx.Exec(`
					UPDATE tables 
					SET status = 'reserved'
					WHERE id IN (
							SELECT table_id 
							FROM table_reservations 
							WHERE status = 'active' 
							AND table_blocked_from <= ? 
							AND grace_period_until > ?
					)
					AND status = 'available'
			`, now, now).Error; err != nil {
			tx.Rollback()
			continue
		}

		// คืนสถานะโต๊ะเป็น available เมื่อเลยเวลา grace period
		if err := tx.Exec(`
					UPDATE tables 
					SET status = 'available'
					WHERE id IN (
							SELECT table_id 
							FROM table_reservations 
							WHERE status = 'active' 
							AND grace_period_until <= ?
					)
					AND status = 'reserved'
			`, now).Error; err != nil {
			tx.Rollback()
			continue
		}

		// อัพเดทสถานะการจองเป็น expired เมื่อเลยเวลา grace period
		if err := tx.Model(&models.TableReservation{}).
			Where("status = 'active' AND grace_period_until <= ?", now).
			Update("status", "expired").Error; err != nil {
			tx.Rollback()
			continue
		}

		tx.Commit()
		time.Sleep(1 * time.Minute) // ตรวจสอบทุก 1 นาที
	}
}
