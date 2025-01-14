package api_handlers

import (
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type PaymentRequest struct {
	UUID          string                      `json:"uuid" binding:"required"`     // UUID ของ QR Code
	TableID       uint                        `json:"table_id" binding:"required"` // ID ของโต๊ะ
	PaymentMethod string                      `json:"payment_method" binding:"required"`
	ServiceCharge float64                     `json:"service_charge"`
	Discounts     []PaymentDiscountRequest    `json:"discounts,omitempty"`
	ExtraCharges  []PaymentExtraChargeRequest `json:"extra_charges,omitempty"`
	StaffID       uint                        `json:"staff_id,omitempty"`
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	tx := db.DB.Begin()

	// 1. ตรวจสอบ QR Code
	var qrCode models.QRCode
	if err := tx.Where("uuid = ? AND table_id = ? AND is_active = ? AND expiry_at > ?",
		req.UUID, req.TableID, true, time.Now()).First(&qrCode).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Invalid or expired QR code",
		})
	}

	// 2. ดึงทุก orders ที่ยังไม่ได้ชำระเงิน
	var orders []models.Order
	if err := tx.Preload("Items.MenuItem").
		Preload("Items.Options.MenuOption").
		Where("uuid = ? AND table_id = ? AND status != ? AND status != ? AND receipt_id IS NULL",
			req.UUID, req.TableID, "completed", "cancelled").
		Find(&orders).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No active orders found",
		})
	}

	if len(orders) == 0 {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "No active orders found for this table",
		})
	}

	// 3. คำนวณยอดรวมทั้งหมด
	var subTotal float64
	for _, order := range orders {
		subTotal += order.Total
	}

	// 4. สร้างใบเสร็จ
	receipt := models.Receipt{
		UUID:          req.UUID,
		TableID:       int(req.TableID),
		SubTotal:      subTotal,
		ServiceCharge: req.ServiceCharge,
		PaymentMethod: req.PaymentMethod,
		StaffID:       req.StaffID,
		CreatedAt:     time.Now(),
	}

	// 5. บันทึกใบเสร็จ
	if err := tx.Create(&receipt).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create receipt",
		})
	}

	// 6. คำนวณและบันทึกส่วนลด
	var totalDiscount float64
	for _, discount := range req.Discounts {
		var discountType models.DiscountType
		if err := tx.First(&discountType, discount.DiscountTypeID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
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

		// บันทึกส่วนลดในใบเสร็จ
		receiptDiscount := models.ReceiptDiscount{
			ReceiptID:      receipt.ID,
			DiscountTypeID: discount.DiscountTypeID,
			Value:          discountAmount,
			StaffID:        req.StaffID,
			Reason:         discount.Reason,
			CreatedAt:      time.Now(),
		}

		if err := tx.Create(&receiptDiscount).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save discount",
			})
		}
	}

	// 7. คำนวณและบันทึกค่าใช้จ่ายเพิ่มเติม
	var totalExtraCharge float64
	for _, charge := range req.ExtraCharges {
		var chargeType models.AdditionalChargeType
		if err := tx.First(&chargeType, charge.ChargeTypeID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid charge type",
			})
		}

		chargeAmount := charge.Amount * float64(charge.Quantity)
		totalExtraCharge += chargeAmount

		// บันทึกค่าใช้จ่ายเพิ่มเติมในใบเสร็จ
		receiptCharge := models.ReceiptCharge{
			ReceiptID:    receipt.ID,
			ChargeTypeID: charge.ChargeTypeID,
			Amount:       charge.Amount,
			Quantity:     charge.Quantity,
			StaffID:      req.StaffID,
			Note:         charge.Note,
			CreatedAt:    time.Now(),
		}

		if err := tx.Create(&receiptCharge).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to save extra charge",
			})
		}
	}

	// 8. อัพเดทยอดรวมในใบเสร็จ
	receipt.DiscountTotal = totalDiscount
	receipt.ChargeTotal = totalExtraCharge
	receipt.Total = subTotal - totalDiscount + totalExtraCharge + req.ServiceCharge

	if err := tx.Save(&receipt).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update receipt total",
		})
	}

	// 9. อัพเดท Orders เพื่อเชื่อมกับใบเสร็จและเปลี่ยนสถานะ
	for _, order := range orders {
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"receipt_id": receipt.ID,
			"status":     "completed",
		}).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update order",
			})
		}
	}

	// 10. Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	// 11. ดึงข้อมูลใบเสร็จที่สมบูรณ์
	var completeReceipt models.Receipt
	if err := db.DB.Preload("Staff").
		Preload("Orders", func(db *gorm.DB) *gorm.DB {
			return db.Order("created_at ASC")
		}).
		Preload("Orders.Items.MenuItem").
		Preload("Orders.Items.Options.MenuOption").
		Preload("Discounts.DiscountType").
		Preload("Charges.ChargeType").
		First(&completeReceipt, receipt.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
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
