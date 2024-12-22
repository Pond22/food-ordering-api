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
	Description string                  `json:"description"`
	Type        string                  `json:"type" example:"single_discount"`
	StartDate   time.Time               `json:"start_date"`
	EndDate     time.Time               `json:"end_date"`
	IsActive    bool                    `json:"is_active"`
	Items       []PromotionItemResponse `json:"items"`
}

type createPromo_req struct {
	Name        string    `json:"name" binding:"required"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date" binding:"required"`
	EndDate     time.Time `json:"end_date" binding:"required"`
	Price       float64   `json:"price" binding:"required"`
	Items       []struct {
		MenuItemID uint `json:"menu_item_id" binding:"required"`
		Quantity   int  `json:"quantity" binding:"required,min=1"`
	} `json:"items" binding:"required"`
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
	// now := time.Now()

	if err := db.DB.Preload("PromotionItems.MenuItem").
		Where("is_active = ?", true).Find(&promotions).Error; err != nil {
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