package api_handlers

import (
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

type Table struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"`
}

// @Summary สร้างโต๊ะใหม่
// @Description สร้างโต๊ะใหม่
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Table true "ข้อมูลโต๊ะ"
// @Success 200 {object} models.MenuItem "รายละเอียดของโต๊ะ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 409 {object} map[string]interface{} "ชื่อโต๊ะซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างโต๊ะ"
// @Router /api/table [post]
// @Tags Table
func Addtable(c *fiber.Ctx) error {
	var req Table
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	if req.Name == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "กรุณาระบุชื่อโต๊ะ",
		})
	}

	if req.Capacity <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "จำนวนที่นั่งต้องไม่ต่ำกว่า 0",
		})
	}

	var existingTable models.Table
	if err := db.DB.Where("name = ?", req.Name).First(&existingTable).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "มีโต๊ะชื่อนี้อยู่แล้ว",
		})
	}

	table := models.Table{
		Name:     req.Name,
		Capacity: req.Capacity,
		Status:   "available",
	}

	if err := db.DB.Create(&table).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ไม่สามารถเพิ่่มโต๊ะได้",
		})
	}

	return c.JSON(table)
}

// func MergeTables(c *fiber.Ctx) error {
// 	var req struct {
//         TableIDs []uint `json:"table_ids"`
//         StaffID  uint   `json:"staff_id"`
//     }

// }