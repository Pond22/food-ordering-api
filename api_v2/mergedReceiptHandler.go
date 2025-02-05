package api_v2

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MergedPaymentRequest struct {
	TableIDs      []uint                      `json:"table_ids" binding:"required,min=2"`
	PaymentMethod string                      `json:"payment_method" binding:"required"`
	ServiceCharge float64                     `json:"service_charge"`
	Discounts     []PaymentDiscountRequest    `json:"discounts,omitempty"`
	ExtraCharges  []PaymentExtraChargeRequest `json:"extra_charges,omitempty"`
	StaffID       uint                        `json:"staff_id" binding:"required"`
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

// @Summary ชำระเงินรวมหลายโต๊ะ
// @Description รวมบิลและชำระเงินสำหรับหลายโต๊ะพร้อมกัน
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body MergedPaymentRequest true "ข้อมูลการชำระเงิน"
// @Success 200 {object} models.Receipt "รายละเอียดใบเสร็จที่รวมแล้ว"
// @Router /api/v2/payment/merge [post]
// @Tags Payment_V2
func CreateMergedReceipt(c *fiber.Ctx) error {
	var req MergedPaymentRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	tx := db.DB.Begin()

	// 1. ดึงออเดอร์ที่ยังไม่ได้ชำระจากทุกโต๊ะ
	var allOrders []models.Order
	for _, tableID := range req.TableIDs {
		var orders []models.Order
		if err := tx.Preload("Items", "status != ?", "cancelled").
			Preload("Items.MenuItem").
			Preload("Items.Options.MenuOption").
			Where("table_id = ? AND status NOT IN (?, ?) AND receipt_id IS NULL",
				tableID, "completed", "cancelled").
			Find(&orders).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusNotFound).JSON(fiber.Map{
				"error": "ไม่พบออเดอร์สำหรับโต๊ะบางโต๊ะ",
			})
		}
		allOrders = append(allOrders, orders...)
	}

	if len(allOrders) == 0 {
		tx.Rollback()
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบออเดอร์ที่ยังไม่ได้ชำระ",
		})
	}

	// 2. คำนวณยอดรวม
	var subTotal float64
	for _, order := range allOrders {
		subTotal += order.Total
	}

	// 3. สร้างใบเสร็จ
	receipt := models.Receipt{
		UUID:          uuid.New().String(),
		TableID:       int(req.TableIDs[0]),
		SubTotal:      subTotal,
		ServiceCharge: req.ServiceCharge,
		PaymentMethod: req.PaymentMethod,
		StaffID:       req.StaffID,
		CreatedAt:     time.Now(),
	}

	if err := tx.Create(&receipt).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถสร้างใบเสร็จได้",
		})
	}

	// 4. บันทึกส่วนลด
	var totalDiscount float64
	for _, discount := range req.Discounts {
		var discountType models.DiscountType
		if err := tx.First(&discountType, discount.DiscountTypeID).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "ประเภทส่วนลดไม่ถูกต้อง",
			})
		}

		var discountAmount float64
		if discountType.Type == "percentage" {
			discountAmount = (subTotal * discount.Value) / 100
		} else {
			discountAmount = discount.Value
		}
		totalDiscount += discountAmount

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
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถบันทึกส่วนลดได้",
			})
		}
	}

	// 5. บันทึกค่าใช้จ่ายเพิ่มเติม
	var totalExtraCharge float64
	for _, charge := range req.ExtraCharges {
		var chargeType models.AdditionalChargeType
		if err := tx.First(&chargeType, charge.ChargeTypeID).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "ประเภทค่าใช้จ่ายเพิ่มเติมไม่ถูกต้อง",
			})
		}

		chargeAmount := charge.Amount * float64(charge.Quantity)
		totalExtraCharge += chargeAmount

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
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถบันทึกค่าใช้จ่ายเพิ่มเติมได้",
			})
		}
	}

	// 6. อัพเดทยอดรวมในใบเสร็จ
	serviceChargeAmount := (subTotal * req.ServiceCharge) / 100 //เอาไอ้ service ที่ส่งมาจากหน้าบ้านหาเปอร์เซ็น
	receipt.ServiceCharge = serviceChargeAmount
	receipt.DiscountTotal = totalDiscount
	receipt.ChargeTotal = totalExtraCharge
	receipt.Total = subTotal + serviceChargeAmount - totalDiscount + totalExtraCharge

	if err := tx.Save(&receipt).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถอัพเดทยอดรวมได้",
		})
	}

	// 7. อัพเดท orders
	for _, order := range allOrders {
		if err := tx.Model(&order).Updates(map[string]interface{}{
			"receipt_id": receipt.ID,
			"status":     "completed",
		}).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถอัพเดทออเดอร์ได้",
			})
		}
	}

	// 8. อัพเดทสถานะโต๊ะ
	for _, tableID := range req.TableIDs {
		if err := tx.Model(&models.Table{}).
			Where("id = ?", tableID).
			Update("status", "available").Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถอัพเดทสถานะโต๊ะได้",
			})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถบันทึกข้อมูลได้",
		})
	}

	// 9. พิมพ์ใบเสร็จ
	if err := PrintReceipt(receipt.ID); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถพิมพ์ใบเสร็จได้",
		})
	}

	// 10. ดึงข้อมูลใบเสร็จที่สมบูรณ์
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
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถดึงข้อมูลใบเสร็จได้",
		})
	}

	return c.JSON(completeReceipt)
}

func PrintReceipt(receiptID uint) error {
	var receipt models.Receipt
	if err := db.DB.Preload("Orders.Items.MenuItem").
		Preload("Orders.Items.Options.MenuOption").
		Preload("Discounts.DiscountType").
		Preload("Charges.ChargeType").
		First(&receipt, receiptID).Error; err != nil {
		return fmt.Errorf("receipt not found")
	}

	// ค้นหาเครื่องพิมพ์หลักที่มีชื่อ "main"
	var printer models.Printer
	if err := db.DB.Where("name = ?", "main").First(&printer).Error; err != nil {
		return fmt.Errorf("main printer not found")
	}

	// สร้างเนื้อหาใบเสร็จในรูปแบบ ESC/POS
	// content := createReceiptPrintContent(receipt)

	// บันทึกลง print_jobs
	printJob := models.PrintJob{
		PrinterID: printer.ID,
		ReceiptID: &receipt.ID,
		// Content:   content,
		JobType:   "receipt",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.DB.Create(&printJob).Error; err != nil {
		return fmt.Errorf("failed to create print job")
	}

	return nil
}
