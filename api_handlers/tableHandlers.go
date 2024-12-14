package api_handlers

import (
	"errors"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"slices"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Table struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"`
}

// @Summary แก้ไขข้อมูลโต๊ะ
// @Description แก้ไขข้อมูลโต๊ะ โดยโต๊ะต้องไม่อยู่ในระหว่างการให้บริการ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID โต๊ะ"
// @Param request body Table true "ข้อมูลโต๊ะที่ต้องการแก้ไข"
// @Success 200 {object} models.Table "รายละเอียดของโต๊ะที่แก้ไขแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 404 {object} map[string]interface{} "ไม่พบโต๊ะที่ต้องการแก้ไข"
// @Failure 409 {object} map[string]interface{} "ชื่อโต๊ะซ้ำกับที่มีอยู่แล้ว"
// @Failure 422 {object} map[string]interface{} "ไม่สามารถแก้ไขโต๊ะที่กำลังใช้งานอยู่"
// @Router /api/table/{id} [put]
// @Tags Table
func UpdateTable(c *fiber.Ctx) error {
	id := c.Params("id")
	tableID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ID ต้องเป็นตัวเลข",
		})
	}

	var req Table
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
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

	var table models.Table
	if err := db.DB.First(&table, tableID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "ไม่พบโต๊ะที่ต้องการแก้ไข",
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการค้นหาโต๊ะ",
		})
	}

	if table.Status == "occupied" {
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "ไม่สามารถแก้ไขโต๊ะที่กำลังใช้งานอยู่",
		})
	}

	var existingTable models.Table
	if err := db.DB.Where("name = ? AND id != ?", req.Name, tableID).First(&existingTable).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "มีโต๊ะชื่อนี้อยู่แล้ว",
		})
	}

	if req.Status != "string" {
		table.Status = req.Status
	}

	table.Name = req.Name
	table.Capacity = req.Capacity

	if err := db.DB.Save(&table).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการแก้ไขข้อมูลโต๊ะ",
		})
	}

	return c.JSON(table)
}

// @Summary ลบโต๊ะ
// @Description เป็นการ softdelete โต๊ะโดยโต๊ะต้องอยุ่ในสถานะพร้อมให้บริการ หรือ ไม่พร้อมให้บริการ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID โต๊ะ"
// @Success 200 {object} map[string]interface{} "ลบโต๊ะสำเร็จ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 409 {object} map[string]interface{} "ชื่อโต๊ะซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างโต๊ะ"
// @Router /api/table/{id} [delete]
// @Tags Table
func DeleteTable(c *fiber.Ctx) error {
	R := c.Params("id")
	id, err := strconv.Atoi(R)

	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "id ผิดพลาดในการแปลงหรือไม่ได้ส่ง id มา",
		})
	}

	var table models.Table

	if err := db.DB.Where("id = ?", id).First(&table).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "หาโต๊ะไม่เจอ",
		})
	}

	if !slices.Contains([]string{"available", "unavailable"}, table.Status) {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "โต๊ะต้องอยู่ในสถานะว่างหรือไม่พร้อมให้บริการเท่านั้น จึงสามารถลบได้",
		})
	}

	if table.GroupID != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "โต๊ะต้องไม่อยู่ในกลุ่มใดๆ",
		})
	}

	if err := db.DB.Delete(&table).Error; err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการลบ",
		})
	}

	return c.Status(200).JSON(fiber.Map{
		"ok": "ok",
	})
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

type req_merge struct {
	TableIDs []uint `json:"table_ids"`
	// StaffID  uint   `json:"staff_id"`
}

// @Summary รวมโต๊ะ
// @Description รวมโต๊ะ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body req_merge true "ข้อมูลโต๊ะ"
// @Success 200 {object} models.MenuItem "รายละเอียดของโต๊ะ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการรวมโต๊ะ"
// @Router /api/table/mergeTable [post]
// @Tags Table
func MergeTables(c *fiber.Ctx) error {
	var req req_merge

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request format"})
	}

	if len(req.TableIDs) < 2 {
		return c.Status(400).JSON(fiber.Map{"error": "Need at least 2 tables to merge"})
	}

	tx := db.DB.Begin()

	// 1. ตรวจสอบว่าโต๊ะทั้งหมดมีอยู่จริง
	var tables []models.Table
	if err := tx.Where("id IN ?", req.TableIDs).Find(&tables).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Tables not found"})
	}

	if len(tables) != len(req.TableIDs) {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Some tables not found"})
	}

	// 2. ตรวจสอบว่าทุกโต๊ะพร้อมให้บริการ
	for _, table := range tables {
		if table.Status != "available" {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{
				"error": fmt.Sprintf("Table %s is not available", table.Name),
			})
		}
		// ตรวจสอบว่าโต๊ะไม่ได้อยู่ในกลุ่มอื่น
		if table.GroupID != nil {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{
				"error": fmt.Sprintf("Table %s is already in another group", table.Name),
			})
		}
	}

	// 3. สร้าง Group ID และผูกโต๊ะเข้าด้วยกัน
	groupID := uuid.New().String()
	mainTableID := req.TableIDs[0]

	for _, table := range tables {
		updates := map[string]interface{}{
			"group_id":  groupID,
			"parent_id": mainTableID,
		}
		if table.ID == mainTableID {
			updates["parent_id"] = nil
		}

		if err := tx.Model(&models.Table{}).Where("id = ?", table.ID).Updates(updates).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Failed to merge tables"})
		}
	}

	// if err := tx.Model(&models.QRCode{}).
	// 	Where("table_id IN ? AND is_active = ?", req.TableIDs[1:], true).
	// 	Update("is_active", false).Error; err != nil {
	// 	tx.Rollback()
	// 	return c.Status(500).JSON(fiber.Map{"error": "Failed to update QR codes"})
	// }

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit changes"})
	}

	return c.JSON(fiber.Map{
		"message":       "Tables merged successfully",
		"group_id":      groupID,
		"main_table_id": mainTableID,
	})
}

type move_req struct {
	FromTableID uint `json:"from_table_id"`
	ToTableID   uint `json:"to_table_id"`
}

// @Summary ย้ายโต๊ะ
// @Description ย้ายโต๊ะโดยจะย้ายได้ต่อเมื่อโต๊ะอยู่ในสถานะ occupied หรือกำลังใช้งานอยู่เท่านั้นเป็นการย้ายโต๊ะที่ลูกค้่่ากำลังใช้งานไปอีกโต๊ะนึง
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body move_req true "ID ของโต๊ะที่จะย้าย ID ของโต๊ะเป้าหมายที่จะย้ายไป"
// @Success 200 {object} models.MenuItem "รายละเอียดของโต๊ะ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการย้ายโต๊ะ"
// @Router /api/table/moveTable [post]
// @Tags Table
func MoveTable(c *fiber.Ctx) error {
	var req move_req

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid request format"})
	}

	tx := db.DB.Begin()

	// 1. ตรวจสอบโต๊ะต้นทาง
	var fromTable models.Table
	if err := tx.First(&fromTable, req.FromTableID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Source table not found"})
	}

	// ตรวจสอบว่าโต๊ะต้นทางมีลูกค้านั่งอยู่จริง
	if fromTable.Status != "occupied" {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "Source table is not occupied"})
	}

	// 2. ตรวจสอบโต๊ะปลายทาง
	var toTable models.Table
	if err := tx.First(&toTable, req.ToTableID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Destination table not found"})
	}

	// ตรวจสอบว่าโต๊ะปลายทางว่างอยู่
	if toTable.Status != "available" {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "Destination table is not available"})
	}

	// 3. ดึง QR code ที่กำลังใช้งานของโต๊ะต้นทาง
	var currentQR models.QRCode
	if err := tx.Where("table_id = ? AND is_active = ?", fromTable.ID, true).First(&currentQR).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Could not find active QR code for source table"})
	}

	// 4. อัพเดตสถานะโต๊ะ
	if err := tx.Model(&fromTable).Update("status", "available").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update source table status"})
	}

	if err := tx.Model(&toTable).Update("status", "occupied").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update destination table status"})
	}

	// 5. อัพเดต table_id ใน QR code
	if err := tx.Model(&currentQR).Update("table_id", toTable.ID).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update QR code table reference"})
	}

	//เหลือทำให้มันจัดการออเดอร์ของโต๊ะ
	//
	//
	//
	//--------------------------

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit changes"})
	}

	return c.JSON(fiber.Map{
		"message":    "Table moved successfully",
		"from_table": fromTable,
		"to_table":   toTable,
		"qr_code":    currentQR,
	})
}

// @Summary แยกโต๊ะ
// @Description แยกโต๊ะหลังจากที่รวมแล้ว
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID ของกลุ่มโต๊ะนั้นๆ"
// @Success 200 {object} models.MenuItem "รายละเอียดของโต๊ะ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการแยกโต๊ะ"
// @Router /api/table/splitTable/{id} [post]
// @Tags Table
func SplitTables(c *fiber.Ctx) error {
	req := c.Params("id")
	if req == "" {
		return c.Status(400).JSON(fiber.Map{"error": "GroupTable ID Require"})
	}

	tx := db.DB.Begin()

	// 1. หาโต๊ะทั้งหมดในกลุ่ม
	var tables []models.Table
	if err := tx.Where("group_id = ?", req).Find(&tables).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Tables not found"})
	}

	if len(tables) < 2 {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "No merged tables found for this group"})
	}

	for _, table := range tables {
		if table.Status != "available" {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{"error": "กลุ่มนี้อยู่ในระหว่างการใช้งานไม่สามรถแยกได้ตอนนี้"})
		}
	}

	// 2. คืนสถานะโต๊ะให้เป็นโต๊ะเดี่ยว
	for _, table := range tables {
		if err := tx.Model(&table).Updates(map[string]interface{}{
			"group_id":  nil,
			"parent_id": nil,
		}).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Failed to split tables"})
		}

		// เปิดใช้งาน QR Code ของแต่ละโต๊ะ
		if err := tx.Model(&models.QRCode{}).
			Where("table_id = ?", table.ID).
			Update("is_active", true).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update QR codes"})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit changes"})
	}

	return c.JSON(fiber.Map{
		"message": "Tables split successfully",
	})
}