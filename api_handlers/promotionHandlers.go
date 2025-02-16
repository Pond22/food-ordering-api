package api_handlers

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
)

type MenuItemBasic struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Price int16  `json:"price"`
}

type PromotionItemResponse struct {
	ID              uint          `json:"id"`
	MenuItem        MenuItemBasic `json:"menu_item"`
	DiscountPercent float64       `json:"discount_percent,omitempty"`
	DiscountAmount  float64       `json:"discount_amount,omitempty"`
	BundlePrice     float64       `json:"bundle_price,omitempty"`
	Quantity        int           `json:"quantity,omitempty"`
}

type PromotionResponse struct {
	ID          uint                    `json:"id" example:"1"`
	Name        string                  `json:"name" example:"ลด 20% ทุกเมนูข้าว"`
	NameEn      string                  `json:"nameEn"`
	NameCh      string                  `json:"nameCh"`
	Description string                  `json:"description"`
	Type        string                  `json:"type" example:"single_discount"`
	StartDate   time.Time               `json:"start_date"`
	EndDate     time.Time               `json:"end_date"`
	IsActive    bool                    `json:"is_active"`
	Items       []PromotionItemResponse `json:"items"`
}

type createPromo_req struct {
	Name          string    `json:"name" binding:"required"`
	NameEn        string    `json:"nameEn"`
	NameCh        string    `json:"nameCh"`
	Description   string    `json:"description"`
	DescriptionEn string    `json:"descriptionEn"`
	DescriptionCh string    `json:"descriptionCh"`
	StartDate     time.Time `json:"start_date" form:"2006-01-02 15:04:05Z07:00"`
	EndDate       time.Time `json:"end_date" form:"2006-01-02 15:04:05Z07:00"`
	Price         float64   `json:"price" binding:"required"`
	Items         []struct {
		MenuItemID uint `json:"menu_item_id" binding:"required"`
		Quantity   int  `json:"quantity" binding:"required,min=1"`
	} `json:"items" binding:"required"`
	MaxSelections int `json:"max_selections,omitempty"`
	MinSelections int `json:"min_selections,omitempty"`
}

type updatePromo_req struct {
	Name          string     `json:"name,omitempty"`
	NameEn        string     `json:"nameEn,omitempty"`
	NameCh        string     `json:"nameCh,omitempty"`
	Description   string     `json:"description,omitempty"`
	DescriptionEn string     `json:"descriptionEn,omitempty"`
	DescriptionCh string     `json:"descriptionCh,omitempty"`
	StartDate     *time.Time `json:"start_date,omitempty"`
	EndDate       *time.Time `json:"end_date,omitempty"`
	Price         *float64   `json:"price,omitempty"`
}

type UpdateStatusRequest struct {
	IsActive bool `json:"is_active" example:"false"`
}

type ErrorResponse struct {
	Error string `json:"error" example:"Invalid input format"`
}

type SuccessResponse struct {
	Message string `json:"message" example:"Status updated successfully"`
}

// @Summary สร้างโปรโมชั่นใหม่
// @Description สร้างโปรโมชั่นใหม่พร้อมรายการสินค้าที่ร่วมรายการ สามารถเป็นได้ทั้งแบบส่วนลดและแบบบันเดิล
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param promotion body createPromo_req true "ข้อมูลโปรโมชั่น"
// @Success 200 {object} PromotionResponse "รายละเอียดของโปรโมชั่นที่สร้างสำเร็จ"
// @Failure 400 {object} ErrorResponse "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 403 {object} ErrorResponse "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions [post]
func CreatePromotion(c *fiber.Ctx) error {
	var req createPromo_req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err})
	}

	// if req.MaxSelections == 0 {
	// 	req.MaxSelections = len(req.Items)
	// }
	// if req.MinSelections == 0 {
	// 	req.MinSelections = len(req.Items)
	// }

	tx := db.DB.Begin()

	// สร้างโปรโมชั่น
	promo := models.Promotion{
		Name:          req.Name,
		NameEn:        req.NameEn,
		NameCh:        req.NameCh,
		Description:   req.Description,
		DescriptionEn: req.DescriptionEn,
		DescriptionCh: req.DescriptionCh,
		StartDate:     req.StartDate,
		EndDate:       req.EndDate,
		Price:         req.Price, // เพิ่มราคาโปรโมชั่น
		IsActive:      true,
		MaxSelections: req.MaxSelections,
		MinSelections: req.MinSelections,
		TotalItems:    len(req.Items),
	}

	if err := tx.Create(&promo).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to create promotion"})
	}

	// ตรวจสอบและเพิ่มรายการในโปรโมชั่น
	for _, item := range req.Items {
		// ตรวจสอบว่ามีเมนูนี้อยู่จริง
		var menuItem models.MenuItem
		if err := tx.First(&menuItem, item.MenuItemID).Error; err != nil {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{
				"error": fmt.Sprintf("Menu item ID %d not found", item.MenuItemID),
			})
		}

		promoItem := models.PromotionItem{
			PromotionID: promo.ID,
			MenuItemID:  item.MenuItemID,
			Quantity:    item.Quantity,
		}

		if err := tx.Create(&promoItem).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create promotion items"})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	// ดึงข้อมูลโปรโมชั่นที่สร้างพร้อมรายการอาหาร
	var completePromo models.Promotion
	if err := db.DB.Preload("Items.MenuItem").First(&completePromo, promo.ID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load created promotion"})
	}

	return c.JSON(completePromo)
}

type Active_PromotionResponse struct {
	ID            uint                    `json:"id" example:"1"`
	Name          string                  `json:"name" example:"ลด 20% ทุกเมนูข้าว"`
	NameEn        string                  `json:"nameEn"`
	NameCh        string                  `json:"nameCh"`
	Description   string                  `json:"description"`
	Type          string                  `json:"type" example:"single_discount"`
	StartDate     time.Time               `json:"start_date"`
	EndDate       time.Time               `json:"end_date"`
	IsActive      bool                    `json:"is_active"`
	MaxSelections int                     `json:"max_selections"`
	MinSelections int                     `json:"min_selections"`
	Items         []PromotionItemResponse `json:"items"`
}

// @Summary ดึงรายการโปรโมชั่นที่กำลังใช้งาน
// @Description ดึงรายการโปรโมชั่นทั้งหมดที่กำลังใช้งานอยู่ (IsActive=true) และอยู่ในช่วงวันที่ที่กำหนด
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} Active_PromotionResponse "รายการโปรโมชั่นที่กำลังใช้งาน"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 403 {object} ErrorResponse "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions/Active [get]
func GetActivePromotions(c *fiber.Ctx) error {
	var promotions []models.Promotion
	now := time.Now()

	if err := db.DB.Preload("Items").
		Preload("Items.MenuItem").
		Preload("Items.MenuItem.Category").
		Preload("Items.MenuItem.OptionGroups").
		Preload("Items.MenuItem.OptionGroups.Options").
		Where("is_active = ? AND start_date <= ? AND end_date >= ? AND deleted_at IS NULL", true, now, now).Find(&promotions).Error; err != nil {
		// Where("is_active = ? AND start_date <= ? AND end_date >= ?", true, now, now).
		return c.Status(500).JSON(fiber.Map{"error": "Failed to fetch promotions"})
	}

	return c.JSON(promotions)
}

// @Summary อัพเดทสถานะการใช้งานโปรโมชั่น
// @Description เปิดหรือปิดการใช้งานโปรโมชั่น โดยการอัพเดทค่า IsActive
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID ของโปรโมชั่น"
// @Param status body UpdateStatusRequest true "สถานะใหม่ของโปรโมชั่น"
// @Success 200 {object} SuccessResponse "อัพเดทสถานะสำเร็จ"
// @Failure 400 {object} ErrorResponse "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 403 {object} ErrorResponse "ไม่มีสิทธิ์เข้าถึง"
// @Failure 404 {object} ErrorResponse "ไม่พบโปรโมชั่นที่ระบุ"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions/status/{id} [patch]
func UpdatePromotionStatus(c *fiber.Ctx) error {
	promoID := c.Params("id")
	var req struct {
		IsActive bool `json:"is_active"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	if err := db.DB.Model(&models.Promotion{}).
		Where("id = ?", promoID).
		Update("is_active", req.IsActive).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update status"})
	}

	return c.JSON(fiber.Map{"message": "Status updated successfully"})
}

// @Summary อัพเดตข้อมูลโปรโมชั่น
// @Description อัพเดตข้อมูลพื้นฐานของโปรโมชั่น (ไม่รวมรายการอาหาร)
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID ของโปรโมชั่น"
// @Param promotion body updatePromo_req true "ข้อมูลโปรโมชั่นที่ต้องการอัพเดต"
// @Success 200 {object} PromotionResponse "รายละเอียดของโปรโมชั่นที่อัพเดตสำเร็จ"
// @Failure 400 {object} ErrorResponse "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 404 {object} ErrorResponse "ไม่พบโปรโมชั่นที่ระบุ"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions/{id} [put]
func UpdatePromotion(c *fiber.Ctx) error {
	promoID := c.Params("id")
	var req updatePromo_req

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input format"})
	}

	tx := db.DB.Begin()

	var promo models.Promotion
	if err := tx.First(&promo, promoID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Promotion not found"})
	}

	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.NameEn != "" {
		updates["NameEn"] = req.NameEn
	}
	if req.NameCh != "" {
		updates["NameCh"] = req.NameCh
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}
	if req.DescriptionEn != "" {
		updates["DescriptionEn"] = req.DescriptionEn
	}
	if req.DescriptionCh != "" {
		updates["DescriptionCh"] = req.DescriptionCh
	}
	if req.StartDate != nil {
		updates["start_date"] = req.StartDate
	}
	if req.EndDate != nil {
		updates["end_date"] = req.EndDate
	}
	if req.Price != nil {
		updates["price"] = req.Price
	}

	// อัพเดทข้อมูลโปรโมชัน
	if err := tx.Model(&promo).Updates(updates).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update promotion"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	// ดึงข้อมูลที่อัพเดทแล้วมาแสดง
	var updatedPromo models.Promotion
	if err := db.DB.Preload("Items.MenuItem").First(&updatedPromo, promoID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load updated promotion"})
	}

	return c.JSON(updatedPromo)
}

// @Summary ลบโปรโมชั่น (Soft Delete)
// @Description ลบโปรโมชั่นแบบ Soft Delete โดยการตั้งค่า DeletedAt
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID ของโปรโมชั่น"
// @Success 200 {object} SuccessResponse "ลบโปรโมชั่นสำเร็จ"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 404 {object} ErrorResponse "ไม่พบโปรโมชั่นที่ระบุ"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions/{id} [delete]
func DeletePromotion(c *fiber.Ctx) error {
	promoID := c.Params("id")

	var promo models.Promotion
	if err := db.DB.First(&promo, promoID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Promotion not found"})
	}

	tx := db.DB.Begin()

	// Soft delete promotion items first
	if err := tx.Model(&models.PromotionItem{}).Where("promotion_id = ?", promo.ID).Update("deleted_at", time.Now()).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete promotion items"})
	}

	// Then soft delete the promotion itself
	if err := tx.Model(&promo).Update("deleted_at", time.Now()).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete promotion"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.JSON(fiber.Map{"message": "Promotion deleted successfully"})
}

// @Summary ดึงข้อมูลโปรโมชั่นตาม ID
// @Description ดึงข้อมูลรายละเอียดของโปรโมชั่นตาม ID ที่ระบุ
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID ของโปรโมชั่น"
// @Success 200 {object} PromotionResponse "รายละเอียดของโปรโมชั่น"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 404 {object} ErrorResponse "ไม่พบโปรโมชั่นที่ระบุ"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions/{id} [get]
func GetPromotionByID(c *fiber.Ctx) error {
	promoID := c.Params("id")

	fmt.Println(promoID)
	var promotion models.Promotion
	if err := db.DB.Preload("Items.MenuItem").
		Preload("Items.MenuItem.Category").
		Preload("Items.MenuItem.OptionGroups").
		Preload("Items.MenuItem.OptionGroups.Options").
		Where("id = ?", promoID).First(&promotion).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Promotion not found"})
	}

	return c.JSON(promotion)
}

// @Summary ดึงข้อมูลโปรโมชั่นAll
// @Description ดึงข้อมูลรายละเอียดของโปรโมชั่นทั้งหมด
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} PromotionResponse "รายละเอียดของโปรโมชั่น"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 404 {object} ErrorResponse "ไม่พบโปรโมชั่นที่ระบุ"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions [get]
func GetAllPromotion(c *fiber.Ctx) error {
	var promotions []models.Promotion
	if err := db.DB.Preload("Items.MenuItem").
		Preload("Items.MenuItem.Category").
		Preload("Items.MenuItem.OptionGroups").
		Preload("Items.MenuItem.OptionGroups.Options").
		Where("deleted_at IS NULL").
		Find(&promotions).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Promotion not found"})
	}

	return c.JSON(promotions)
}

// @Summary อัพเดทรูปภาพโปรโมชั่น
// @Description อัพเดทรูปภาพของโปรโมชั่น รองรับไฟล์ JPG และ PNG ขนาดไม่เกิน 5MB
// @Tags promotions
// @Accept multipart/form-data
// @Produce json
// @Param id path integer true "ID ของโปรโมชั่น"
// @Param image formData file true "รูปภาพโปรโมชั่น"
// @Success 200 {object} models.Promotion "รายละเอียดของโปรโมชั่นที่อัพเดทรูปภาพแล้ว"
// @Failure 400 {object} ErrorResponse "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 404 {object} ErrorResponse "ไม่พบโปรโมชั่นที่ระบุ"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions/image/{id} [put]
func UpdatePromotionImage(c *fiber.Ctx) error { //Security BearerAuth
	id := c.Params("id")
	promoID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid promotion ID format",
		})
	}

	// ตรวจสอบว่ามีโปรโมชั่นอยู่จริง
	var existingPromotion models.Promotion
	if err := db.DB.First(&existingPromotion, promoID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Promotion not found",
		})
	}

	// รับไฟล์รูปภาพ
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Error getting image file",
		})
	}

	// ตรวจสอบขนาดไฟล์ (จำกัดขนาดไม่เกิน 5MB)
	if file.Size > 5*1024*1024 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file size must be less than 5MB",
		})
	}

	// ตรวจสอบนามสกุลไฟล์
	filename := file.Filename
	ext := filepath.Ext(filename)
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	if !validExts[strings.ToLower(ext)] {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Only JPG and PNG files are allowed",
		})
	}

	// เปิดไฟล์
	fileContent, err := file.Open()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error opening image file",
		})
	}
	defer fileContent.Close()

	// อ่านข้อมูลไฟล์
	buffer := make([]byte, file.Size)
	if _, err := fileContent.Read(buffer); err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error reading image file",
		})
	}

	// อัพเดทรูปภาพในฐานข้อมูล
	if err := db.DB.Model(&existingPromotion).Update("image", buffer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update promotion image",
		})
	}

	// ดึงข้อมูลที่อัพเดทแล้วมาแสดง
	var updatedPromotion models.Promotion
	if err := db.DB.Preload("Items.MenuItem").First(&updatedPromotion, promoID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading updated promotion",
		})
	}

	return c.JSON(updatedPromotion)
}

// @Summary ลบรายการอาหารในโปรโมชั่น
// @Description ลบรายการอาหารที่ระบุออกจากโปรโมชั่น
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID ของโปรโมชั่น"
// @Param item_id path int true "ID ของรายการอาหารในโปรโมชั่น"
// @Success 200 {object} SuccessResponse "ลบรายการอาหารสำเร็จ"
// @Failure 400 {object} ErrorResponse "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 404 {object} ErrorResponse "ไม่พบรายการที่ระบุ"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions/{id}/items/{item_id} [delete]
func DeletePromotionItem(c *fiber.Ctx) error {
	promoID := c.Params("id")
	itemID := c.Params("item_id")

	tx := db.DB.Begin()

	// ตรวจสอบว่ามีโปรโมชั่นนี้อยู่จริง
	var promo models.Promotion
	if err := tx.First(&promo, promoID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Promotion not found"})
	}

	// ลบรายการอาหาร
	if err := tx.Where("id = ? AND promotion_id = ?", itemID, promoID).Delete(&models.PromotionItem{}).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to delete promotion item"})
	}

	// นับจำนวนรายการที่เหลือ
	var totalItems int64
	if err := tx.Model(&models.PromotionItem{}).Where("promotion_id = ?", promoID).Count(&totalItems).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count remaining items"})
	}

	// อัพเดท total_items ในโปรโมชั่น
	if err := tx.Model(&promo).Update("total_items", totalItems).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update total items"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	return c.JSON(fiber.Map{"message": "Promotion item deleted successfully"})
}

// @Summary เพิ่มรายการอาหารในโปรโมชั่น
// @Description เพิ่มรายการอาหารใหม่เข้าไปในโปรโมชั่นที่มีอยู่แล้ว
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID ของโปรโมชั่น"
// @Param items body []struct{MenuItemID uint "json:\"menu_item_id\"" binding:"required" Quantity int "json:\"quantity\"" binding:"required,min=1"} true "รายการอาหารที่ต้องการเพิ่ม"
// @Success 200 {object} PromotionResponse "รายละเอียดของโปรโมชั่นที่อัพเดทแล้ว"
// @Failure 400 {object} ErrorResponse "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 404 {object} ErrorResponse "ไม่พบโปรโมชั่นที่ระบุ"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions/{id}/items [post]
func AddPromotionItems(c *fiber.Ctx) error {
	promoID := c.Params("id")
	var req struct {
		Items []struct {
			MenuItemID uint `json:"menu_item_id" binding:"required"`
			Quantity   int  `json:"quantity" binding:"required,min=1"`
		} `json:"items" binding:"required"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input format"})
	}

	tx := db.DB.Begin()

	// ตรวจสอบว่ามีโปรโมชั่นนี้อยู่จริง
	var promo models.Promotion
	if err := tx.First(&promo, promoID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Promotion not found"})
	}

	// ตรวจสอบและเพิ่มรายการในโปรโมชั่น
	for _, item := range req.Items {
		// ตรวจสอบว่ามีเมนูนี้อยู่จริง
		var menuItem models.MenuItem
		if err := tx.First(&menuItem, item.MenuItemID).Error; err != nil {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{
				"error": fmt.Sprintf("Menu item ID %d not found", item.MenuItemID),
			})
		}

		// ตรวจสอบว่าเมนูนี้ยังไม่มีในโปรโมชั่น
		var existingItem models.PromotionItem
		if err := tx.Where("promotion_id = ? AND menu_item_id = ?", promoID, item.MenuItemID).First(&existingItem).Error; err == nil {
			tx.Rollback()
			return c.Status(400).JSON(fiber.Map{
				"error": fmt.Sprintf("Menu item ID %d already exists in this promotion", item.MenuItemID),
			})
		}

		promoItem := models.PromotionItem{
			PromotionID: promo.ID,
			MenuItemID:  item.MenuItemID,
			Quantity:    item.Quantity,
		}

		if err := tx.Create(&promoItem).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Failed to create promotion items"})
		}
	}

	// นับจำนวนรายการทั้งหมดที่มีอยู่
	var totalItems int64
	if err := tx.Model(&models.PromotionItem{}).Where("promotion_id = ?", promoID).Count(&totalItems).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to count total items"})
	}

	// อัพเดท TotalItems
	if err := tx.Model(&promo).Update("total_items", totalItems).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update total items"})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to commit transaction"})
	}

	// ดึงข้อมูลที่อัพเดทแล้วมาแสดง
	var updatedPromo models.Promotion
	if err := db.DB.Preload("Items.MenuItem").First(&updatedPromo, promoID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Failed to load updated promotion"})
	}

	return c.JSON(updatedPromo)
}
