package api_v2

import (
	"errors"
	"food-ordering-api/db"
	"food-ordering-api/models"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// @Summary ยกเลิกการจองโดยตรง
// @Description ยกเลิกการจองโดยอ้างอิงจาก ID การจอง
// @Accept json
// @Produce json
// @Param id path int true "ID ของการจอง"
// @Success 200 {object} map[string]interface{} "ยกเลิกการจองสำเร็จ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 404 {object} map[string]interface{} "ไม่พบการจอง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดภายในระบบ"
// @Router /api/v2/reservation/cancel/{id} [post]
// @Tags Reservation_V2
func CancelReservation(c *fiber.Ctx) error {
	reservationID := c.Params("id")
	if reservationID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "กรุณาระบุ ID การจอง",
		})
	}

	tx := db.DB.Begin()

	// ค้นหาการจอง
	var reservation models.TableReservation
	if err := tx.Where("id = ? AND status = ?", reservationID, "active").First(&reservation).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"error": "ไม่พบการจองที่ระบุ หรือการจองถูกยกเลิกไปแล้ว",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการค้นหาการจอง",
		})
	}

	// อัพเดทสถานะการจอง
	if err := tx.Model(&reservation).Update("status", "cancelled").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถยกเลิกการจองได้",
		})
	}

	// ถ้าโต๊ะถูกกั้นไว้แล้ว (status = reserved) ให้คืนสถานะเป็น available
	if err := tx.Model(&models.Table{}).
		Where("id = ? AND status = ?", reservation.TableID, "reserved").
		Update("status", "available").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอัพเดทสถานะโต๊ะได้",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการบันทึกข้อมูล",
		})
	}

	return c.JSON(fiber.Map{
		"message": "ยกเลิกการจองสำเร็จ",
	})
}
