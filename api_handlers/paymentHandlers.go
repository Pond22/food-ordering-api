package api_handlers

import (
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// Request structs
type PaymentRequest struct {
	OrderID       uint                        `json:"order_id" binding:"required"`
	PaymentMethod string                      `json:"payment_method" binding:"required"`
	ServiceCharge float64                     `json:"service_charge"`
	Discounts     []PaymentDiscountRequest    `json:"discounts,omitempty"`
	ExtraCharges  []PaymentExtraChargeRequest `json:"extra_charges,omitempty"`
	StaffID       uint                        `json:"staff_id,omitempty"` //อย่าลืมแก้ binding:"required"`
}

type PaymentDiscountRequest struct {
	DiscountTypeID uint    `json:"discount_type_id" binding:"required"`
	Value          float64 `json:"value" binding:"required"`
	Reason         string  `json:"reason,omitempty"`
}

type PaymentExtraChargeRequest struct {
	ChargeTypeID uint    `json:"charge_type_id" binding:"required"`
	Amount       float64 `json:"amount" binding:"required"`
	Quantity     int     `json:"quantity" binding:"required,min=1"`
	Note         string  `json:"note,omitempty"`
}

// @Summary ชำระเงิน
// @Description ชำระเงินสำหรับออเดอร์ พร้อมส่วนลดและค่าใช้จ่ายเพิ่มเติม
// @Accept json
// @Produce json
// @Param payment body PaymentRequest true "ข้อมูลการชำระเงิน"
// @Success 200 {object} models.Receipt
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 404 {object} map[string]interface{} "ไม่พบออเดอร์"
// @Router /api/payment/process [post]
// @Tags Payment
func ProcessPayment(c *fiber.Ctx) error {
	var req PaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	tx := db.DB.Begin()

	// 1. ตรวจสอบและดึงข้อมูลออเดอร์
	var order models.Order
	if err := tx.Preload("Items.MenuItem").
		Preload("Items.Options").
		First(&order, req.OrderID).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// 2. คำนวณยอดรวมก่อนส่วนลด
	var subTotal float64 = order.Total

	// 3. คำนวณส่วนลดทั้งหมด
	var totalDiscount float64 = 0
	var discounts []models.OrderDiscount

	for _, discount := range req.Discounts {
		var discountType models.DiscountType
		if err := tx.First(&discountType, discount.DiscountTypeID).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid discount type",
			})
		}

		// คำนวณส่วนลด
		var discountAmount float64
		if discountType.Type == "percentage" {
			discountAmount = (subTotal * discount.Value) / 100
		} else {
			discountAmount = discount.Value
		}

		totalDiscount += discountAmount

		// สร้างรายการส่วนลด
		orderDiscount := models.OrderDiscount{
			OrderID:        order.ID,
			DiscountTypeID: discount.DiscountTypeID,
			Value:          discountAmount,
			StaffID:        req.StaffID,
			Reason:         discount.Reason,
			CreatedAt:      time.Now(),
		}
		discounts = append(discounts, orderDiscount)
	}

	// 4. คำนวณค่าใช้จ่ายเพิ่มเติม
	var totalExtraCharge float64 = 0
	var extraCharges []models.OrderAdditionalCharge

	for _, charge := range req.ExtraCharges {
		var chargeType models.AdditionalChargeType
		if err := tx.First(&chargeType, charge.ChargeTypeID).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid charge type",
			})
		}

		chargeAmount := charge.Amount * float64(charge.Quantity)
		totalExtraCharge += chargeAmount

		// สร้างรายการค่าใช้จ่ายเพิ่มเติม
		orderCharge := models.OrderAdditionalCharge{
			OrderID:      order.ID,
			ChargeTypeID: charge.ChargeTypeID,
			Amount:       charge.Amount,
			Quantity:     charge.Quantity,
			StaffID:      req.StaffID,
			Note:         charge.Note,
			CreatedAt:    time.Now(),
		}
		extraCharges = append(extraCharges, orderCharge)
	}

	// 5. สร้างใบเสร็จ
	receipt := models.Receipt{
		OrderID:       order.ID,
		UUID:          order.UUID,
		TableID:       order.TableID,
		SubTotal:      subTotal,
		DiscountTotal: totalDiscount,
		ChargeTotal:   totalExtraCharge,
		ServiceCharge: req.ServiceCharge,
		Total:         subTotal - totalDiscount + totalExtraCharge + req.ServiceCharge,
		PaymentMethod: req.PaymentMethod,
		StaffID:       req.StaffID,
		CreatedAt:     time.Now(),
	}

	// 6. บันทึกข้อมูลทั้งหมด
	if err := tx.Create(&receipt).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create receipt",
		})
	}

	for _, discount := range discounts {
		if err := tx.Create(&discount).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save discounts",
			})
		}
	}

	for _, charge := range extraCharges {
		if err := tx.Create(&charge).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save extra charges",
			})
		}
	}

	// 7. อัพเดทสถานะออเดอร์เป็นชำระเงินแล้ว
	if err := tx.Model(&order).Update("status", "completed").Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order status",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	// ดึงข้อมูลใบเสร็จที่สมบูรณ์
	var completeReceipt models.Receipt
	if err := db.DB.Preload("Staff").
		Preload("Order.Items.MenuItem").
		Preload("Discounts.DiscountType").
		Preload("Charges.ChargeType").
		First(&completeReceipt, receipt.ID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load complete receipt",
		})
	}

	return c.JSON(completeReceipt)
}

// @Summary ดึงประเภทส่วนลดที่ใช้งานได้
// @Description ดึงรายการประเภทส่วนลดที่เปิดใช้งาน
// @Produce json
// @Success 200 {array} models.DiscountType
// @Router /api/payment/discount-types [get]
// @Tags Payment
func GetActiveDiscountTypes(c *fiber.Ctx) error {
	var discountTypes []models.DiscountType
	if err := db.DB.Where("is_active = ?", true).Find(&discountTypes).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch discount types",
		})
	}
	return c.JSON(discountTypes)
}

// @Summary ดึงประเภทค่าใช้จ่ายเพิ่มเติมที่ใช้งานได้
// @Description ดึงรายการประเภทค่าใช้จ่ายเพิ่มเติมที่เปิดใช้งาน
// @Produce json
// @Success 200 {array} models.AdditionalChargeType
// @Router /api/payment/charge-types [get]
// @Tags Payment
func GetActiveChargeTypes(c *fiber.Ctx) error {
	var chargeTypes []models.AdditionalChargeType
	if err := db.DB.Where("is_active = ?", true).Find(&chargeTypes).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch charge types",
		})
	}
	return c.JSON(chargeTypes)
}

// @Summary ดึงใบเสร็จ
// @Description ดึงข้อมูลใบเสร็จตาม ID
// @Produce json
// @Param id path int true "Receipt ID"
// @Success 200 {object} models.Receipt
// @Router /api/payment/receipt/{id} [get]
// @Tags Payment
func GetReceipt(c *fiber.Ctx) error {
	id := c.Params("id")

	var receipt models.Receipt
	if err := db.DB.Preload("Staff").
		Preload("Order.Items.MenuItem").
		Preload("Discounts.DiscountType").
		Preload("Charges.ChargeType").
		First(&receipt, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Receipt not found",
		})
	}

	return c.JSON(receipt)
}

// @Summary สร้างประเภทส่วนลดใหม่
// @Description สร้างประเภทส่วนลดใหม่ (สำหรับผู้จัดการ)
// @Accept json
// @Produce json
// @Param discount body models.DiscountType true "ข้อมูลประเภทส่วนลด"
// @Success 200 {object} models.DiscountType
// @Router /api/payment/discount-types [post]
// @Tags Payment
func CreateDiscountType(c *fiber.Ctx) error {
	var discountType models.DiscountType
	if err := c.BodyParser(&discountType); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if err := db.DB.Create(&discountType).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create discount type",
		})
	}

	return c.JSON(discountType)
}

// @Summary สร้างประเภทค่าใช้จ่ายเพิ่มเติมใหม่
// @Description สร้างประเภทค่าใช้จ่ายเพิ่มเติมใหม่ (สำหรับผู้จัดการ)
// @Accept json
// @Produce json
// @Param charge body models.AdditionalChargeType true "ข้อมูลประเภทค่าใช้จ่าย"
// @Success 200 {object} models.AdditionalChargeType
// @Router /api/payment/charge-types [post]
// @Tags Payment
func CreateChargeType(c *fiber.Ctx) error {
	var chargeType models.AdditionalChargeType
	if err := c.BodyParser(&chargeType); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if err := db.DB.Create(&chargeType).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create charge type",
		})
	}

	return c.JSON(chargeType)
}