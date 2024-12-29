package api_handlers

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
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
	Name        string    `json:"name" binding:"required"`
	NameEn      string    `json:"nameEn"`
	NameCh      string    `json:"nameCh"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	Price       float64   `json:"price" binding:"required"`
	Items       []struct {
		MenuItemID uint `json:"menu_item_id" binding:"required"`
		Quantity   int  `json:"quantity" binding:"required,min=1"`
	} `json:"items" binding:"required"`
}

type updatePromo_req struct {
	Name        string     `json:"name,omitempty"`
	NameEn      string     `json:"nameEn,omitempty"`
	NameCh      string     `json:"nameCh,omitempty"`
	Description string     `json:"description,omitempty"`
	StartDate   *time.Time `json:"start_date,omitempty"`
	EndDate     *time.Time `json:"end_date,omitempty"`
	Price       *float64   `json:"price,omitempty"`
	Items       []struct {
		ID         uint `json:"id,omitempty"` // สำหรับ item ที่มีอยู่แล้ว
		MenuItemID uint `json:"menu_item_id"`
		Quantity   int  `json:"quantity" binding:"min=1"`
	} `json:"items"`
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
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	tx := db.DB.Begin()

	// สร้างโปรโมชั่น
	promo := models.Promotion{
		Name:        req.Name,
		NameEn:      req.NameEn,
		NameCh:      req.NameCh,
		Description: req.Description,
		StartDate:   req.StartDate,
		EndDate:     req.EndDate,
		Price:       req.Price, // เพิ่มราคาโปรโมชั่น
		IsActive:    true,
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

// @Summary ดึงรายการโปรโมชั่นที่กำลังใช้งาน
// @Description ดึงรายการโปรโมชั่นทั้งหมดที่กำลังใช้งานอยู่ (IsActive=true) และอยู่ในช่วงวันที่ที่กำหนด
// @Tags promotions
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {array} PromotionResponse "รายการโปรโมชั่นที่กำลังใช้งาน"
// @Failure 401 {object} ErrorResponse "ไม่ได้รับอนุญาต"
// @Failure 403 {object} ErrorResponse "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} ErrorResponse "เกิดข้อผิดพลาดภายในเซิร์ฟเวอร์"
// @Router /api/promotions [get]
func GetActivePromotions(c *fiber.Ctx) error {
	var promotions []models.Promotion
	now := time.Now()

	if err := db.DB.Preload("Items").
		Where("is_active = ? AND start_date <= ? AND end_date >= ?", true, now, now).Find(&promotions).Error; err != nil {
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
// @Description อัพเดตข้อมูลโปรโมชั่นและรายการสินค้าที่ร่วมรายการ
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

	// ตรวจสอบว่ามีโปรโมชั่นที่ต้องการอัพเดตหรือไม่
	var promo models.Promotion
	if err := tx.First(&promo, promoID).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{"error": "Promotion not found"})
	}

	// อัพเดตข้อมูลโปรโมชั่น
	updates := map[string]interface{}{}

	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.NameEn != "" {
		updates["name_en"] = req.NameEn
	}
	if req.NameCh != "" {
		updates["name_ch"] = req.NameCh
	}
	if req.Description != "" {
		updates["description"] = req.Description
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

	if err := tx.Model(&promo).Updates(updates).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Failed to update promotion"})
	}

	// ลบรายการเดิมทั้งหมดเฉพาะเมื่อมีการส่ง items มา
	if len(req.Items) > 0 {
		if err := tx.Where("promotion_id = ?", promoID).Delete(&models.PromotionItem{}).Error; err != nil {
			tx.Rollback()
			return c.Status(500).JSON(fiber.Map{"error": "Failed to update promotion items"})
		}
	}

	// เพิ่มรายการใหม่
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

	// ดึงข้อมูลโปรโมชั่นที่อัพเดตพร้อมรายการอาหาร
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

	// Then soft delete the promotion
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

	var promotion models.Promotion
	if err := db.DB.Preload("Items.MenuItem").First(&promotion, promoID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "Promotion not found"})
	}

	return c.JSON(promotion)
}