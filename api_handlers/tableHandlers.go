package api_handlers

import (
	"errors"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"slices"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Table struct {
	Name     string `json:"name"`
	Capacity int    `json:"capacity"`
	Status   string `json:"status"`
}

func hasGroupID(groupID *string) bool {
	return groupID != nil && *groupID != ""
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

	if hasGroupID(table.GroupID) {
		return c.Status(http.StatusUnprocessableEntity).JSON(fiber.Map{
			"error": "ไม่สามารถแก้ไขโต๊ะที่กำลังรวมอยู่",
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

type Req_merge struct {
	TableIDs []uint `json:"table_ids"`
	// StaffID  uint   `json:"staff_id"`
}

// @Summary รวมโต๊ะ
// @Description รวมโต๊ะ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body Req_merge true "ข้อมูลโต๊ะ"
// @Success 200 {object} models.MenuItem "รายละเอียดของโต๊ะ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการรวมโต๊ะ"
// @Router /api/table/mergeTable [post]
// @Tags Table
func MergeTables(c *fiber.Ctx) error {
	var req Req_merge

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
	ForceMove   bool `json:"force_move,omitempty"` // true = ยอมรับความจุน้อยกว่า ยืดหยุ่่นท่องไว้
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

	if req.FromTableID == req.ToTableID { //อย่าหาย้ายไปโต๊ะเดิม
		return c.Status(400).JSON(fiber.Map{
			"error": "Cannot move table to itself",
		})
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

	// ตรวจสอบความจุโต๊ะปลายทางแบบยืดหยุ่น
	if toTable.Capacity < fromTable.Capacity && !req.ForceMove {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"error":                "Destination table capacity is too small. Set force_move=true to override",
			"current_capacity":     fromTable.Capacity,
			"destination_capacity": toTable.Capacity,
		})
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

	// ย้ายเฉพาะออเดอร์ของ QR code ปัจจุบัน
	if err := tx.Model(&models.Order{}).
		Where("table_id = ? AND uuid = ? AND status != ?",
			fromTable.ID,
			currentQR.UUID,
			"cancelled").
		Update("table_id", toTable.ID).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update orders",
		})
	}

	var updatedFromTable, updatedToTable models.Table
	var updatedQR models.QRCode

	tx.First(&updatedFromTable, fromTable.ID)
	tx.First(&updatedToTable, toTable.ID)
	tx.First(&updatedQR, currentQR.ID)

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit changes"})
	}

	return c.JSON(fiber.Map{
		"message":    "Table moved successfully",
		"from_table": updatedFromTable,
		"to_table":   updatedToTable,
		"qr_code":    updatedQR,
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

type ReservationRequest struct {
	CustomerName string    `json:"customer_name"`
	PhoneNumber  string    `json:"phone_number"`
	GuestCount   int       `json:"guest_count"`
	ReservedFor  time.Time `json:"reserved_for"`
}

// @Summary จองโต๊ะ
// @Description จองโต๊ะโดยใช้กฎการจองที่กำหนดไว้
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ReservationRequest true "ข้อมูลการจองโต๊ะ"
// @Param id path string true "ID ของโต๊ะ"
// @Success 200 {object} map[string]interface{} "จองโต๊ะสำเร็จ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 404 {object} map[string]interface{} "ไม่พบโต๊ะ"
// @Failure 409 {object} map[string]interface{} "มีการจองซ้ำซ้อน"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดภายในระบบ"
// @Router /api/table/reservedTable/{id} [post]
// @Tags Table
func ReservedTable(c *fiber.Ctx) error {
	tableID := c.Params("id")
	if tableID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "กรุณาระบุหมายเลขโต๊ะ",
		})
	}

	// รับข้อมูลการจอง
	var req ReservationRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "ข้อมูลการจองไม่ถูกต้อง",
		})
	}

	// Validation
	if req.CustomerName == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุชื่อลูกค้า"})
	}
	if req.PhoneNumber == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุเบอร์โทรศัพท์"})
	}
	if req.GuestCount <= 0 {
		return c.Status(400).JSON(fiber.Map{"error": "จำนวนลูกค้าต้องมากกว่า 0"})
	}
	if req.ReservedFor.Before(time.Now()) {
		return c.Status(400).JSON(fiber.Map{"error": "ไม่สามารถจองย้อนหลังได้"})
	}

	tx := db.DB.Begin()

	// ดึง active rule
	var activeRule models.ReservationRules
	if err := tx.Where("is_active = ?", true).First(&activeRule).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่พบกฎการจองที่ใช้งานอยู่",
		})
	}

	// ตรวจสอบโต๊ะ
	var table models.Table
	if err := tx.Where("id = ?", tableID).First(&table).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบโต๊ะที่ระบุ"})
	}

	if table.Status != "available" {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "โต๊ะไม่ว่าง"})
	}

	if table.Capacity < req.GuestCount {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{
			"error":            "จำนวนลูกค้าเกินความจุของโต๊ะ",
			"table_capacity":   table.Capacity,
			"requested_guests": req.GuestCount,
		})
	}

	// คำนวณเวลา
	now := time.Now()
	tableBlockedFrom := req.ReservedFor.Add(-time.Duration(activeRule.PreReservationMinutes) * time.Minute)
	gracePeriodUntil := req.ReservedFor.Add(time.Duration(activeRule.GracePeriodMinutes) * time.Minute)

	// ตรวจสอบการจองซ้ำซ้อน
	var conflictReservation models.TableReservation
	if err := tx.Where(`
			table_id = ? AND status = 'active' AND (
					(table_blocked_from BETWEEN ? AND ?) OR
					(grace_period_until BETWEEN ? AND ?) OR
					(? BETWEEN table_blocked_from AND grace_period_until)
			)`,
		tableID, tableBlockedFrom, gracePeriodUntil,
		tableBlockedFrom, gracePeriodUntil,
		req.ReservedFor,
	).First(&conflictReservation).Error; err == nil {
		tx.Rollback()
		return c.Status(409).JSON(fiber.Map{
			"error": "มีการจองในช่วงเวลานี้แล้ว",
		})
	}

	// สร้างการจอง
	reservation := models.TableReservation{
		TableID:          uint(table.ID),
		CustomerName:     req.CustomerName,
		PhoneNumber:      req.PhoneNumber,
		GuestCount:       req.GuestCount,
		ReservedFor:      req.ReservedFor,
		TableBlockedFrom: tableBlockedFrom,
		GracePeriodUntil: gracePeriodUntil,
		Status:           "active",
	}

	if err := tx.Create(&reservation).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถบันทึกการจองได้"})
	}

	// อัพเดทสถานะโต๊ะเฉพาะเมื่อถึงเวลา tableBlockedFrom
	if now.After(tableBlockedFrom) || now.Equal(tableBlockedFrom) {
		if err := tx.Model(&table).Update("status", "reserved").Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถอัพเดทสถานะโต๊ะได้"})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "เกิดข้อผิดพลาดในการบันทึกข้อมูล"})
	}

	return c.Status(200).JSON(fiber.Map{
		"message":            "จองโต๊ะสำเร็จ",
		"reservation_id":     reservation.ID,
		"table_blocked_from": tableBlockedFrom,
		"reserved_for":       req.ReservedFor,
		"grace_period_until": gracePeriodUntil,
	})
}

// @Summary ยกเลืกจองโต๊ะ
// @Description ยกเลืกจองโต๊ะโดยโต๊ะต้องอยู่ในถานะพร้อมให้บริการถึงยกเลิกจองได้
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "ID ของกลุ่มโต๊ะนั้นๆ"
// @Success 200 {object} map[string]interface{} "เค"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการยกเลืกจองโต๊ะ"
// @Router /api/table/unreservedTable/{id} [post]
// @Tags Table
func UnreservedTable(c *fiber.Ctx) error {
	tableID := c.Params("id")
	if tableID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "กรุณาระบุหมายเลขโต๊ะ"})
	}

	tx := db.DB.Begin()

	// ตรวจสอบโต๊ะ
	var table models.Table
	if err := tx.Where("id = ?", tableID).First(&table).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "ไม่พบโต๊ะที่ระบุ"})
	}

	if table.Status != "reserved" {
		tx.Rollback()
		return c.Status(400).JSON(fiber.Map{"error": "โต๊ะนี้ไม่ได้ถูกจอง"})
	}

	// อัพเดทสถานะการจอง
	if err := tx.Model(&models.TableReservation{}).
		Where("table_id = ? AND status = ?", tableID, "active").
		Update("status", "cancelled").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถยกเลิกการจองได้"})
	}

	// อัพเดทสถานะโต๊ะ - รวมถึงโต๊ะที่อยู่ในกลุ่มเดียวกัน
	if hasGroupID(table.GroupID) {
		if err := tx.Model(&models.Table{}).
			Where("group_id = ?", table.GroupID).
			Update("status", "available").Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถอัพเดทสถานะโต๊ะรวมได้"})
		}
	} else {
		if err := tx.Model(&table).Update("status", "available").Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "ไม่สามารถอัพเดทสถานะโต๊ะได้"})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "เกิดข้อผิดพลาดในการบันทึกข้อมูล"})
	}

	return c.Status(200).JSON(fiber.Map{
		"message": "ยกเลิกการจองสำเร็จ",
	})
}

// @Summary สลับสถานะพร้อมใช้งาน/ไม่พร้อมใช้งานของโต๊ะ
// @Description สลับสถานะระหว่างพร้อมใช้งาน (available) และไม่พร้อมใช้งาน (unavailable) โดยโต๊ะต้องไม่อยู่ในระหว่างการให้บริการ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID โต๊ะ"
// @Success 200 {object} map[string]interface{} "สลับสถานะสำเร็จ"
// @Failure 404 {object} map[string]interface{} "ไม่พบโต๊ะ"
// @Failure 422 {object} map[string]interface{} "ไม่สามารถสลับสถานะได้"
// @Router /api/table/setstatus/{id} [put]
// @Tags Table
func ToggleTableStatus(c *fiber.Ctx) error {
	id := c.Params("id")
	if id == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "ต้องระบุ ID โต๊ะ",
		})
	}

	tx := db.DB.Begin()

	var table models.Table
	if err := tx.First(&table, id).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{
			"error": "ไม่พบโต๊ะ",
		})
	}

	// ตรวจสอบว่าโต๊ะไม่ได้อยู่ในระหว่างใช้งาน
	if table.Status == "occupied" || table.Status == "reserved" {
		tx.Rollback()
		return c.Status(422).JSON(fiber.Map{
			"error": "ไม่สามารถเปลี่ยนสถานะโต๊ะที่กำลังใช้งานหรือจองอยู่",
		})
	}

	if hasGroupID(table.GroupID) {
		tx.Rollback()
		return c.Status(422).JSON(fiber.Map{
			"error": "ไม่สามารถเปลี่ยนสถานะโต๊ะที่อยู่ในกลุ่ม",
		})
	}

	// สลับสถานะ
	newStatus := "unavailable"
	if table.Status == "unavailable" {
		newStatus = "available"
	}

	if err := tx.Model(&table).Update("status", newStatus).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการอัพเดตสถานะ",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการบันทึกข้อมูล",
		})
	}

	return c.JSON(fiber.Map{
		"message":    "สลับสถานะสำเร็จ",
		"table":      table,
		"new_status": newStatus,
	})
}

// @Summary ดึงรายการอาหารที่ต้องคิดเงินตาม UUID
// @Description ดึงรายการอาหารที่มีสถานะ served และ pending สำหรับการคิดเงิน
// @Accept json
// @Produce json
// @Param uuid path string true "UUID ของโต๊ะ"
// @Success 200 {object} map[string]interface{} "รายการอาหารที่ต้องคิดเงิน"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 404 {object} map[string]interface{} "ไม่พบข้อมูล"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูล"
// @Router /api/table/billable/{uuid} [get]
// @Tags Table
func GetBillableItems(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	if uuid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "กรุณาระบุ UUID",
		})
	}

	// ดึงรายการอาหารที่ต้องคิดเงิน
	var orderItems []models.OrderItem
	if err := db.DB.Joins("Order").
		Joins("MenuItem").
		Joins("MenuItem.Category").
		Preload("Options.MenuOption").
		Preload("PromotionUsage.Promotion"). // เพิ่ม Preload สำหรับโปรโมชั่น
		Where("\"Order\".uuid = ? AND order_items.status IN ?", uuid, []string{"served", "pending"}).
		Find(&orderItems).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถดึงข้อมูลรายการอาหารได้",
		})
	}

	if len(orderItems) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"message": "ไม่พบรายการอาหารที่ต้องคิดเงินสำหรับ UUID นี้",
		})
	}

	var response struct {
		UUID  string         `json:"uuid"`
		Items []BillableItem `json:"items"`
		Total float64        `json:"total"`
	}

	response.UUID = uuid
	var total float64

	// แปลงข้อมูลรายการอาหาร
	for _, item := range orderItems {
		billableItem := BillableItem{
			ID:        item.ID,
			OrderID:   item.OrderID,
			Name:      item.MenuItem.Name,
			Category:  item.MenuItem.Category.Name,
			Quantity:  item.Quantity,
			Price:     item.Price,
			Status:    item.Status,
			Notes:     item.Notes,
			CreatedAt: item.CreatedAt,
			ItemTotal: float64(item.Quantity) * item.Price,
			Options:   make([]OptionInfo, 0),
		}

		// ตรวจสอบและเพิ่มข้อมูลโปรโมชั่น
		if item.PromotionUsage != nil && item.PromotionUsage.Promotion.ID > 0 {
			billableItem.Promotion = &PromotionInfo{
				ID:          item.PromotionUsage.Promotion.ID,
				Name:        item.PromotionUsage.Promotion.Name,
				Description: item.PromotionUsage.Promotion.Description,
			}
		}
		// เพิ่มข้อมูลตัวเลือกเสริม
		for _, opt := range item.Options {
			billableItem.Options = append(billableItem.Options, OptionInfo{
				ID:       opt.ID,
				Name:     opt.MenuOption.Name,
				Price:    opt.Price,
				Quantity: opt.Quantity,
			})
			billableItem.ItemTotal += opt.Price * float64(opt.Quantity)
		}

		response.Items = append(response.Items, billableItem)
		total += billableItem.ItemTotal
	}

	response.Total = total
	return c.JSON(response)
}

// โครงสร้างข้อมูลสำหรับการส่งกลับ
type BillableItem struct {
	ID        uint           `json:"id"`
	OrderID   uint           `json:"order_id"`
	Name      string         `json:"name"`
	Category  string         `json:"category"`
	Quantity  int            `json:"quantity"`
	Price     float64        `json:"price"`
	Status    string         `json:"status"`
	Notes     string         `json:"notes"`
	CreatedAt time.Time      `json:"created_at"`
	ItemTotal float64        `json:"item_total"`
	Options   []OptionInfo   `json:"options"`
	Promotion *PromotionInfo `json:"promotion,omitempty"`
}

type OptionInfo struct {
	ID       uint    `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

type ReservationResponse struct {
	models.TableReservation
	TableName string `json:"table_name"`
}

type PromotionInfo struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

// @Summary ดึงข้อมูลการจองทั้งหมด
// @Description ดึงข้อมูลการจองโต๊ะทั้งหมดพร้อมชื่อโต๊ะ เรียงตามวันที่จองจากใหม่ไปเก่า
// @Tags Reservation
// @Accept json
// @Produce json
// @Success 200 {array} ReservationResponse "รายการการจองทั้งหมด"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูล"
// @Router /api/table/reservations [get]
// @Tags Table
func GetAllReservations(c *fiber.Ctx) error {
	var reservations []struct {
		models.TableReservation
		TableName string `json:"table_name"`
	}

	if err := db.DB.Table("table_reservations").
		Select("table_reservations.*, tables.name as table_name").
		Joins("LEFT JOIN tables ON tables.id = table_reservations.table_id").
		Order("reserved_for DESC").
		Scan(&reservations).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการดึงข้อมูล",
		})
	}

	return c.JSON(reservations)
}
