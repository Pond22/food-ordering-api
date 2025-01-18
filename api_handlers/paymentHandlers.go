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
// @Router /api/payment/discount-types/Active [get]
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
// @Router /api/payment/charge-types/Active [get]
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

type UpdateDiscountTypeRequest struct {
	Name     string   `json:"name,omitempty"`
	Type     string   `json:"type,omitempty"`  // percentage/amount
	Value    *float64 `json:"value,omitempty"` // pointer เพื่อให้รู้ว่ามีการส่งค่ามาจริงๆ
	IsActive *bool    `json:"is_active,omitempty"`
}

type UpdateChargeTypeRequest struct {
	Name          string   `json:"name,omitempty"`
	DefaultAmount *float64 `json:"default_amount,omitempty"`
	IsActive      *bool    `json:"is_active,omitempty"`
}

type CreateDiscountTypeRequest struct {
	Name     string  `json:"name" binding:"required"`
	Type     string  `json:"type" binding:"required,oneof=percentage amount"` // รับได้แค่ percentage หรือ amount
	Value    float64 `json:"value" binding:"required"`                        // ถ้าเป็น percentage ต้อง 0-100
	IsActive bool    `json:"isActive,omitempty"`
}

// DiscountType Handlers

// @Summary สร้างประเภทส่วนลดใหม่
// @Description สร้างประเภทส่วนลดใหม่ (สำหรับผู้จัดการ) ฟิลด์ Type รับได้แค่ percentage หรือ amount โดย percentage คือส่วนลดเป็นเปอร์เซ็น amount เป็นจำนวนเงินตรงๆ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param discountType body CreateDiscountTypeRequest true "ข้อมูลประเภทส่วนลด"
// @Success 201 {object} models.DiscountType "ประเภทส่วนลดที่สร้างสำเร็จ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 409 {object} map[string]interface{} "ชื่อประเภทส่วนลดซ้ำ"
// @Router /api/payment/discount-types [post]
// @Tags Payment Types
func CreateDiscountType(c *fiber.Ctx) error {
	var req CreateDiscountTypeRequest
	if err := c.BodyParser(&req); err != nil {
		// กรณีที่ parse request body ไม่สำเร็จ จะส่ง 400 Bad Request
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// Validation
	if req.Type == "percentage" && (req.Value < 0 || req.Value > 100) {
		// กรณีที่เป็นเปอร์เซ็นต์แต่ค่าไม่อยู่ในช่วง 0-100
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Percentage value must be between 0 and 100",
		})
	}

	// สร้าง model จาก request
	discountType := models.DiscountType{
		Name:      req.Name,
		Type:      req.Type,
		Value:     req.Value,
		IsActive:  req.IsActive,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// บันทึกลงฐานข้อมูล
	if err := db.DB.Create(&discountType).Error; err != nil {
		// กรณีเกิด error ในการบันทึก เช่น
		// - ชื่อซ้ำ (unique constraint)
		// - การเชื่อมต่อ DB มีปัญหา
		// - อื่นๆ
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create discount type: " + err.Error(),
		})
	}

	// กรณีสำเร็จ:
	// - ใช้ Status 201 (Created) เพราะเป็นการสร้างข้อมูลใหม่
	// - ส่งข้อมูลที่สร้างกลับไป รวมถึง ID และ timestamps ที่ระบบสร้างให้
	return c.Status(fiber.StatusCreated).JSON(discountType)
}

// @Summary อัพเดตข้อมูลประเภทส่วนลด
// @Description อัพเดตข้อมูลประเภทส่วนลดตาม ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "Discount Type ID"
// @Param discountType body UpdateDiscountTypeRequest true "ข้อมูลที่ต้องการอัพเดต"
// @Success 200 {object} models.DiscountType "ประเภทส่วนลดที่อัพเดตแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 404 {object} map[string]interface{} "ไม่พบประเภทส่วนลด"
// @Failure 409 {object} map[string]interface{} "ชื่อประเภทส่วนลดซ้ำ"
// @Router /api/payment/discount-types/{id} [put]
// @Tags Payment Types
func UpdateDiscountType(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateDiscountTypeRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// ค้นหาข้อมูลเดิม
	var discountType models.DiscountType
	if err := db.DB.First(&discountType, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Discount type not found",
		})
	}

	// อัพเดตข้อมูล
	if req.Name != "" {
		// ตรวจสอบชื่อซ้ำ
		var existing models.DiscountType
		if err := db.DB.Where("name = ? AND id != ?", req.Name, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Discount type name already exists",
			})
		}
		discountType.Name = req.Name
	}

	if req.Type != "" {
		if req.Type != "percentage" && req.Type != "amount" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Type must be either 'percentage' or 'amount'",
			})
		}
		discountType.Type = req.Type
	}

	if req.Value != nil {
		if discountType.Type == "percentage" && (*req.Value < 0 || *req.Value > 100) {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Percentage value must be between 0 and 100",
			})
		}
		discountType.Value = *req.Value
	}

	if req.IsActive != nil {
		discountType.IsActive = *req.IsActive
	}

	discountType.UpdatedAt = time.Now()

	if err := db.DB.Save(&discountType).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update discount type",
		})
	}

	return c.JSON(discountType)
}

// AdditionalChargeType Handlers

type CreateChargeTypeRequest struct {
	Name          string  `json:"name" binding:"required"`
	DefaultAmount float64 `json:"defaultAmount" binding:"required,min=0"`
	IsActive      bool    `json:"isActive,omitempty"` // optional, จะใช้ค่า default จาก model ถ้าไม่ได้ส่งมา
}

// @Summary สร้างประเภทค่าใช้จ่ายเพิ่มเติมใหม่
// @Description สร้างประเภทค่าใช้จ่ายเพิ่มเติมใหม่ (สำหรับผู้จัดการ)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param chargeType body CreateChargeTypeRequest true "ข้อมูลประเภทค่าใช้จ่าย"
// @Success 201 {object} models.AdditionalChargeType "ประเภทค่าใช้จ่ายที่สร้างสำเร็จ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 409 {object} map[string]interface{} "ชื่อประเภทค่าใช้จ่ายซ้ำ"
// @Router /api/payment/charge-types [post]
// @Tags Payment Types
func CreateChargeType(c *fiber.Ctx) error {
	var req CreateChargeTypeRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// สร้าง model จาก request
	chargeType := models.AdditionalChargeType{
		Name:          req.Name,
		DefaultAmount: req.DefaultAmount,
		IsActive:      req.IsActive, // ถ้าไม่ได้ส่งมา จะใช้ค่า default จาก model
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := db.DB.Create(&chargeType).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create charge type",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(chargeType)
}

// @Summary อัพเดตข้อมูลประเภทค่าใช้จ่าย
// @Description อัพเดตข้อมูลประเภทค่าใช้จ่ายตาม ID
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "Charge Type ID"
// @Param chargeType body UpdateChargeTypeRequest true "ข้อมูลที่ต้องการอัพเดต"
// @Success 200 {object} models.AdditionalChargeType "ประเภทค่าใช้จ่ายที่อัพเดตแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 404 {object} map[string]interface{} "ไม่พบประเภทค่าใช้จ่าย"
// @Failure 409 {object} map[string]interface{} "ชื่อประเภทค่าใช้จ่ายซ้ำ"
// @Router /api/payment/charge-types/{id} [put]
// @Tags Payment Types
func UpdateChargeType(c *fiber.Ctx) error {
	id := c.Params("id")
	var req UpdateChargeTypeRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// ค้นหาข้อมูลเดิม
	var chargeType models.AdditionalChargeType
	if err := db.DB.First(&chargeType, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Charge type not found",
		})
	}

	// อัพเดตข้อมูล
	if req.Name != "" {
		// ตรวจสอบชื่อซ้ำ
		var existing models.AdditionalChargeType
		if err := db.DB.Where("name = ? AND id != ?", req.Name, id).First(&existing).Error; err == nil {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"error": "Charge type name already exists",
			})
		}
		chargeType.Name = req.Name
	}

	if req.DefaultAmount != nil {
		if *req.DefaultAmount < 0 {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Default amount cannot be negative",
			})
		}
		chargeType.DefaultAmount = *req.DefaultAmount
	}

	if req.IsActive != nil {
		chargeType.IsActive = *req.IsActive
	}

	chargeType.UpdatedAt = time.Now()

	if err := db.DB.Save(&chargeType).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update charge type",
		})
	}

	return c.JSON(chargeType)
}

// @Summary ลบประเภทส่วนลด
// @Description ลบประเภทส่วนลดตาม ID (Soft Delete)
// @Produce json
// @Security BearerAuth
// @Param id path integer true "Discount Type ID"
// @Success 200 {object} map[string]interface{} "ลบสำเร็จ"
// @Failure 404 {object} map[string]interface{} "ไม่พบประเภทส่วนลด"
// @Router /api/payment/discount-types/{id} [delete]
// @Tags Payment Types
func DeleteDiscountType(c *fiber.Ctx) error {
	id := c.Params("id")

	// ตรวจสอบว่ามีข้อมูลอยู่หรือไม่
	var discountType models.DiscountType
	if err := db.DB.First(&discountType, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Discount type not found",
		})
	}

	// Soft delete โดยการปิดการใช้งาน
	discountType.IsActive = false
	discountType.UpdatedAt = time.Now()

	if err := db.DB.Save(&discountType).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete discount type",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Discount type deleted successfully",
	})
}

// @Summary ลบประเภทค่าใช้จ่ายเพิ่มเติม
// @Description ลบประเภทค่าใช้จ่ายเพิ่มเติมตาม ID (Soft Delete)
// @Produce json
// @Security BearerAuth
// @Param id path integer true "Charge Type ID"
// @Success 200 {object} map[string]interface{} "ลบสำเร็จ"
// @Failure 404 {object} map[string]interface{} "ไม่พบประเภทค่าใช้จ่าย"
// @Router /api/payment/charge-types/{id} [delete]
// @Tags Payment Types
func DeleteChargeType(c *fiber.Ctx) error {
	id := c.Params("id")

	// ตรวจสอบว่ามีข้อมูลอยู่หรือไม่
	var chargeType models.AdditionalChargeType
	if err := db.DB.First(&chargeType, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Charge type not found",
		})
	}

	// Soft delete โดยการปิดการใช้งาน
	chargeType.IsActive = false
	chargeType.UpdatedAt = time.Now()

	if err := db.DB.Save(&chargeType).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to delete charge type",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Charge type deleted successfully",
	})
}

// @Summary ดึงข้อมูลประเภทส่วนลดตาม ID
// @Description ดึงข้อมูลรายละเอียดของประเภทส่วนลดตาม ID
// @Produce json
// @Security BearerAuth
// @Param id path integer true "Discount Type ID"
// @Success 200 {object} models.DiscountType "ข้อมูลประเภทส่วนลด"
// @Failure 404 {object} map[string]interface{} "ไม่พบประเภทส่วนลด"
// @Router /api/payment/discount-types/{id} [get]
// @Tags Payment Types
func GetDiscountType(c *fiber.Ctx) error {
	id := c.Params("id")

	var discountType models.DiscountType
	if err := db.DB.First(&discountType, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Discount type not found",
		})
	}

	return c.JSON(discountType)
}

// @Summary ดึงข้อมูลประเภทค่าใช้จ่ายตาม ID
// @Description ดึงข้อมูลรายละเอียดของประเภทค่าใช้จ่ายตาม ID
// @Produce json
// @Security BearerAuth
// @Param id path integer true "Charge Type ID"
// @Success 200 {object} models.AdditionalChargeType "ข้อมูลประเภทค่าใช้จ่าย"
// @Failure 404 {object} map[string]interface{} "ไม่พบประเภทค่าใช้จ่าย"
// @Router /api/payment/charge-types/{id} [get]
// @Tags Payment Types
func GetChargeType(c *fiber.Ctx) error {
	id := c.Params("id")

	var chargeType models.AdditionalChargeType
	if err := db.DB.First(&chargeType, id).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Charge type not found",
		})
	}

	return c.JSON(chargeType)
}

// @Summary ดึงประเภทส่วนลดทั้งหมด
// @Description ดึงรายการประเภทส่วนลดทั้งหมด ทั้งที่เปิดและปิดใช้งาน
// @Produce json
// @Success 200 {array} models.DiscountType "รายการประเภทส่วนลดทั้งหมด"
// @Failure 500 {object} map[string]interface{} "มีข้อผิดพลาดในการดึงข้อมูล"
// @Router /api/payment/discount-types [get]
// @Tags Payment Types
func GetAllDiscountTypes(c *fiber.Ctx) error {
	var discountTypes []models.DiscountType
	if err := db.DB.Find(&discountTypes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch discount types",
		})
	}
	return c.JSON(discountTypes)
}

// @Summary ดึงประเภทค่าใช้จ่ายเพิ่มเติมทั้งหมด
// @Description ดึงรายการประเภทค่าใช้จ่ายเพิ่มเติมทั้งหมด ทั้งที่เปิดและปิดใช้งาน
// @Produce json
// @Success 200 {array} models.AdditionalChargeType "รายการประเภทค่าใช้จ่ายทั้งหมด"
// @Failure 500 {object} map[string]interface{} "มีข้อผิดพลาดในการดึงข้อมูล"
// @Router /api/payment/charge-types [get]
// @Tags Payment Types
func GetAllChargeTypes(c *fiber.Ctx) error {
	var chargeTypes []models.AdditionalChargeType
	if err := db.DB.Find(&chargeTypes).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch charge types",
		})
	}
	return c.JSON(chargeTypes)
}
