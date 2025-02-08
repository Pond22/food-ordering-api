package api_handlers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
	"gorm.io/gorm"
)

// Request structs
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignCategoryRequest struct {
	CategoryIDs []uint `json:"category_ids" binding:"required"`
}

type PrinterResponse struct {
	ID         uint              `json:"id"`
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	IPAddress  string            `json:"ip_address"`
	Port       int               `json:"port"`
	VendorID   string            `json:"vendor_id,omitempty"`
	ProductID  string            `json:"product_id,omitempty"`
	PaperSize  string            `json:"paper_size"`
	Status     string            `json:"status"`
	LastSeen   time.Time         `json:"last_seen"`
	Categories []models.Category `json:"categories"`
}

// // @Summary ดึงรายการหมวดหมู่ทั้งหมด
// // @Description ดึงรายการหมวดหมู่เครื่องพิมพ์ทั้งหมด
// // @Produce json
// // @Success 200 {array} models.Category
// // @Router /api/printers/printer-categories [get]
// // @Tags Printer
// func GetPrinterCategories(c *fiber.Ctx) error {
// 	var categories []models.Category
// 	if err := db.DB.Find(&categories).Error; err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch categories",
// 		})
// 	}
// 	return c.JSON(categories)
// }

// // @Summary อัพเดทข้อมูลหมวดหมู่
// // @Description อัพเดทข้อมูลของหมวดหมู่เครื่องพิมพ์
// // @Accept json
// // @Produce json
// // @Param id path int true "Category ID"
// // @Param category body UpdateCategoryRequest true "ข้อมูลที่ต้องการอัพเดท"
// // @Success 200 {object} models.Category
// // @Router /api/printers/printer-categories/{id} [put]
// // @Tags Printer
// func UpdatePrinterCategory(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	var req UpdateCategoryRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request format",
// 		})
// 	}

// 	var category models.Category
// 	if err := db.DB.First(&category, id).Error; err != nil {
// 		return c.Status(http.StatusNotFound).JSON(fiber.Map{
// 			"error": "Category not found",
// 		})
// 	}

// 	updates := map[string]interface{}{}
// 	if req.Name != "" {
// 		updates["name"] = req.Name
// 	}
// 	if req.Description != "" {
// 		updates["description"] = req.Description
// 	}

// 	if err := db.DB.Model(&category).Updates(updates).Error; err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to update category",
// 		})
// 	}

// 	return c.JSON(category)
// }

// // @Summary ลบหมวดหมู่
// // @Description ลบหมวดหมู่เครื่องพิมพ์
// // @Produce json
// // @Param id path int true "Category ID"
// // @Success 200 {object} map[string]interface{}
// // @Router /api/printers/printer-categories/{id} [delete]
// // @Tags Printer
// func DeletePrinterCategory(c *fiber.Ctx) error {
// 	id := c.Params("id")

// 	// ลบความสัมพันธ์กับเครื่องพิมพ์ก่อน
// 	if err := db.DB.Exec("DELETE FROM printer_categories WHERE category_id = ?", id).Error; err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to remove printer associations",
// 		})
// 	}

// 	if err := db.DB.Delete(&models.Category{}, id).Error; err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to delete category",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"message": "Category deleted successfully",
// 	})
// }

// @Summary กำหนดหมวดหมู่ให้เครื่องพิมพ์
// @Description กำหนดหมวดหมู่ที่เครื่องพิมพ์สามารถพิมพ์ได้
// @Accept json
// @Produce json
// @Param id path int true "Printer ID"
// @Param categories body AssignCategoryRequest true "รายการ ID ของหมวดหมู่"
// @Success 200 {object} models.Printer
// @Router /api/printers/categories/{id} [post]
// @Tags Printer
func AssignPrinterCategories(c *fiber.Ctx) error {
	printerID := c.Params("id")
	var req AssignCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// ตรวจสอบว่ามี category IDs ส่งมาหรือไม่
	if len(req.CategoryIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Category IDs are required",
		})
	}

	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// ตรวจสอบว่ามีเครื่องพิมพ์อยู่จริง
	var printer models.Printer
	if err := tx.First(&printer, printerID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Printer not found",
		})
	}

	// ตรวจสอบว่าหมวดหมู่ทั้งหมดที่ส่งมามีอยู่จริง
	var categories []models.Category
	if err := tx.Find(&categories, req.CategoryIDs).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Failed to fetch categories",
		})
	}

	if len(categories) != len(req.CategoryIDs) {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Some categories not found",
		})
	}

	// ดึงความสัมพันธ์ที่มีอยู่เดิม
	var existingRelations []struct {
		PrinterID  uint
		CategoryID uint
	}
	if err := tx.Table("printer_categories").
		Where("printer_id = ?", printer.ID).
		Find(&existingRelations).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch existing relations",
		})
	}

	// สร้าง map เพื่อเช็คความสัมพันธ์ที่มีอยู่แล้ว
	existingMap := make(map[uint]bool)
	for _, rel := range existingRelations {
		existingMap[rel.CategoryID] = true
	}

	// เพิ่มเฉพาะความสัมพันธ์ใหม่ที่ยังไม่มี
	for _, catID := range req.CategoryIDs {
		if !existingMap[catID] {
			if err := tx.Exec("INSERT INTO printer_categories (printer_id, category_id) VALUES (?, ?)",
				printer.ID, catID).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to assign categories",
				})
			}
		}
	}

	// โหลดข้อมูลที่อัพเดทแล้ว
	var updatedPrinter models.Printer
	if err := tx.Preload("Categories").First(&updatedPrinter, printerID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated printer data",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	return c.JSON(updatedPrinter)
}

// @Summary ดึงรายการหมวดหมู่ของเครื่องพิมพ์
// @Description ดึงรายการหมวดหมู่ที่เครื่องพิมพ์สามารถพิมพ์ได้
// @Produce json
// @Param id path int true "Printer ID"
// @Success 200 {array} models.Category
// @Router /api/printers/categories/{id} [get]
// @Tags Printer
func getPrinterCategories(printer *models.Printer) ([]models.Category, error) {
	var categories []models.Category
	err := db.DB.Model(printer).Association("Categories").Find(&categories)
	return categories, err
}

// @Summary ดึงข้อมูลเครื่องพิมพ์ทั้งหมด
// @Description ดึงรายการเครื่องพิมพ์ทั้งหมดพร้อมหมวดหมู่
// @Produce json
// @Success 200 {array} PrinterResponse
// @Router /api/printers [get]
// @Tags Printer
func GetAllPrinters(c *fiber.Ctx) error {
	var printers []models.Printer
	if err := db.DB.Preload("Categories").Find(&printers).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch printers",
		})
	}

	return c.JSON(printers)
}

// @Summary ดึงข้อมูลเครื่องพิมพ์ตาม ID
// @Description ดึงข้อมูลเครื่องพิมพ์และหมวดหมู่ตาม ID ที่ระบุ
// @Produce json
// @Param id path int true "Printer ID"
// @Success 200 {object} PrinterResponse
// @Router /api/printers/{id} [get]
// @Tags Printer
// func GetPrinterByID(c *fiber.Ctx) error {
// 	id := c.Params("id")

// 	var printer models.Printer
// 	if err := db.DB.Preload("Categories").First(&printer, id).Error; err != nil {
// 		return c.Status(http.StatusNotFound).JSON(fiber.Map{
// 			"error": "Printer not found",
// 		})
// 	}

// 	return c.JSON(printer)
// }

// ฟังก์ชันสำหรับทำความสะอาดข้อความ
func cleanText(text string) string {
	// แทนที่ตัวอักษรควบคุมและจัดการช่องว่าง
	text = strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) || unicode.IsSpace(r) {
			return r
		}
		return -1
	}, text)
	return strings.TrimSpace(text)
}

// ฟังก์ชันสำหรับเตรียมเนื้อหาสำหรับพิมพ์ออเดอร์ไปครัว
func prepareOrderPrintContent(job models.PrintJob) ([]byte, error) {
	var content bytes.Buffer

	if job.Order != nil {
		headerLines := []string{
			"Order #" + fmt.Sprintf("%d", job.Order.ID),
			"โต๊ะ: " + fmt.Sprintf("%d", job.Order.TableID),
			"----------------------------------------",
			fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05")),
			"----------------------------------------",
		}

		for _, line := range headerLines {
			content.WriteString(cleanText(line) + "\n")
		}

		// แยกรายการตามสถานะ
		var newItems []models.OrderItem
		for _, item := range job.Order.Items {
			if item.Status == "pending" {
				newItems = append(newItems, item)
			}
		}

		if len(newItems) > 0 {
			content.WriteString("\n[รายการใหม่]\n")
			content.WriteString("----------------------------------------\n")

			for i, item := range newItems {
				// หมายเลขรายการและชื่ออาหาร
				itemLine := fmt.Sprintf("%d. %s", i+1, item.MenuItem.Name)
				if item.Quantity > 1 {
					itemLine += fmt.Sprintf(" x%d", item.Quantity)
				}
				content.WriteString(cleanText(itemLine) + "\n")

				// ตัวเลือกเพิ่มเติม
				for _, opt := range item.Options {
					// ใช้ MenuOption ที่เชื่อมโยงกับ OrderItemOption
					optionLine := fmt.Sprintf("   • %s: %s",
						cleanText(opt.MenuOption.OptionGroup.Name),
						cleanText(opt.MenuOption.Name))
					content.WriteString(optionLine + "\n")
				}

				// หมายเหตุพิเศษ
				if item.Notes != "" {
					content.WriteString(fmt.Sprintf("   [หมายเหตุ: %s]\n", cleanText(item.Notes)))
				}

				// เว้นบรรทัดระหว่างรายการ
				if i < len(newItems)-1 {
					content.WriteString("----------------\n")
				}
			}
		}

		footerLines := []string{
			"========================================",
			"** กรุณาตรวจสอบรายการให้ครบถ้วน **",
			"========================================",
		}

		for _, line := range footerLines {
			content.WriteString(line + "\n")
		}
	} else {
		content.Write(job.Content)
	}

	return content.Bytes(), nil
}

func prepareReceiptPrintContent(job models.PrintJob) ([]byte, error) {
	var content bytes.Buffer

	tableIDStrings := strings.Split(job.Receipt.TableID, ",")

	var tableIDDisplay string
	if len(tableIDStrings) > 1 {
		lastIndex := len(tableIDStrings) - 1
		tableIDDisplay = strings.Join(tableIDStrings[:lastIndex], ", ") + " และ " + tableIDStrings[lastIndex]
	} else if len(tableIDStrings) == 1 {
		tableIDDisplay = tableIDStrings[0]
	}

	if job.Receipt != nil {
		headerLines := []string{
			"***** ใบเสร็จรับเงิน *****",
			"Receipt #" + fmt.Sprintf("%d", job.Receipt.ID),
			"โต๊ะ: " + tableIDDisplay,
			"----------------------------------------",
			fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05")),
			"----------------------------------------",
		}

		for _, line := range headerLines {
			content.WriteString(cleanText(line) + "\n")
		}

		content.WriteString("[ รายการอาหาร ]\n")
		content.WriteString("----------------------------------------\n")

		// Group similar items
		itemGroups := make(map[string]struct {
			MenuItem models.MenuItem
			Quantity int
			Price    float64
			Options  []models.OrderItemOption
			Notes    string
		})

		for _, order := range job.Receipt.Orders {
			for _, item := range order.Items {
				key := fmt.Sprintf("%d-%s", item.MenuItemID, item.Notes)
				if group, exists := itemGroups[key]; exists {
					group.Quantity += item.Quantity
					group.Price += item.Price
					group.Options = append(group.Options, item.Options...)
					itemGroups[key] = group
				} else {
					itemGroups[key] = struct {
						MenuItem models.MenuItem
						Quantity int
						Price    float64
						Options  []models.OrderItemOption
						Notes    string
					}{
						MenuItem: item.MenuItem,
						Quantity: item.Quantity,
						Price:    item.Price,
						Options:  item.Options,
						Notes:    item.Notes,
					}
				}
			}
		}

		i := 1
		for _, group := range itemGroups {
			itemLine := fmt.Sprintf("%d. %s", i, group.MenuItem.Name)
			if group.Quantity > 1 {
				itemLine += fmt.Sprintf(" x%d", group.Quantity)
			}
			itemLine += fmt.Sprintf("   ฿%.2f", group.Price)
			content.WriteString(cleanText(itemLine) + "\n")

			for _, opt := range group.Options {
				optionLine := fmt.Sprintf("   • %s   ฿%.2f",
					cleanText(opt.MenuOption.Name),
					opt.Price)
				content.WriteString(optionLine + "\n")
			}

			if group.Notes != "" {
				content.WriteString(fmt.Sprintf("   [หมายเหตุ: %s]\n", cleanText(group.Notes)))
			}
			i++
		}

		content.WriteString("----------------------------------------\n")

		subTotalLine := fmt.Sprintf("ยอดรวม: ฿%.2f", job.Receipt.SubTotal)
		discountLine := fmt.Sprintf("ส่วนลด: ฿%.2f", job.Receipt.DiscountTotal)
		extraChargesLine := fmt.Sprintf("ค่าใช้จ่ายเพิ่มเติม: ฿%.2f", job.Receipt.ChargeTotal)
		serviceChargeLine := fmt.Sprintf("ค่าบริการ: ฿%.2f", job.Receipt.ServiceCharge)
		totalLine := fmt.Sprintf("ยอดสุทธิ: ฿%.2f", job.Receipt.Total)

		summaryLines := []string{
			subTotalLine,
			discountLine,
			extraChargesLine,
			serviceChargeLine,
			"----------------------------------------",
			totalLine,
			"----------------------------------------",
		}

		for _, line := range summaryLines {
			content.WriteString(cleanText(line) + "\n")
		}

		paymentInfo := fmt.Sprintf("ชำระโดย: %s", cleanText(job.Receipt.PaymentMethod))
		content.WriteString(paymentInfo + "\n")

		employeeInfo := fmt.Sprintf("พนักงาน: %d", job.Receipt.StaffID)
		content.WriteString(employeeInfo + "\n")

		footerLines := []string{
			"========================================",
			"ขอบคุณที่ใช้บริการ",
			"========================================",
		}

		for _, line := range footerLines {
			content.WriteString(line + "\n")
		}
	} else {
		content.Write(job.Content)
	}

	return content.Bytes(), nil
}

func prepareCancelPrintContent(job models.PrintJob) ([]byte, error) {
	var content bytes.Buffer

	if job.Order != nil {
		content.WriteString(fmt.Sprintf("== ใบแจ้งยกเลิกรายการอาหาร ==\n"))
		content.WriteString(fmt.Sprintf("โต๊ะ: %d\n", job.Order.TableID))
		content.WriteString("----------------------------------------\n")
		content.WriteString(" รายการอาหาร                จำนวนที่ยกเลิก\n")
		content.WriteString("----------------------------------------\n")

		// ดึงข้อมูลจากฟิลด์ Content
		items := strings.Split(string(job.Content), "\n")
		for _, item := range items {
			parts := strings.Split(item, "|")
			if len(parts) == 2 {
				itemName := fmt.Sprintf("%-25s", parts[0])
				quantity := fmt.Sprintf("%5s", parts[1])
				content.WriteString(fmt.Sprintf("%s %s\n", itemName, quantity))
			}
		}

		content.WriteString("----------------------------------------\n")
		content.WriteString(fmt.Sprintf("เวลายกเลิก: %s\n", time.Now().Format("02/01/2006 15:04:05")))
		content.WriteString("========================================\n")
		content.WriteString("กรุณาตรวจสอบการยกเลิก\n")
		content.WriteString("========================================\n")
	}

	return content.Bytes(), nil
}

func prepareQRCodePrintContent(job models.PrintJob) ([]byte, error) {
	var content bytes.Buffer

	// เพิ่มส่วนหัว
	headerLines := []string{
		"===== QR Code สำหรับโต๊ะ =====",
		fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05")),
		"========================",
	}

	for _, line := range headerLines {
		content.WriteString(cleanText(line) + "\n")
	}

	// เพิ่มบรรทัดว่างสำหรับ spacing ก่อน QR Code
	content.WriteString("\n\n")

	// ตรวจสอบว่ามี QR Code ใน job content หรือไม่
	if len(job.Content) > 0 {
		content.Write(job.Content) // เขียน PNG bytes โดยตรง
	} else {
		// กรณีไม่มีข้อมูล QR Code
		content.WriteString("[ไม่พบ QR Code]")
	}

	// เพิ่มส่วนท้าย
	footerLines := []string{
		"\n\n========================",
		"โปรดสแกน QR Code เพื่อเข้าสู่ระบบ",
		"========================",
	}

	for _, line := range footerLines {
		content.WriteString(cleanText(line) + "\n")
	}

	return content.Bytes(), nil
}

// type PrintJobResponse struct {
// 	ID        uint            `json:"id"`
// 	PrinterID uint            `json:"printer_id"`
// 	OrderID   *uint           `json:"order_id,omitempty"`
// 	Content   []byte          `json:"content"`
// 	Status    string          `json:"status"`
// 	CreatedAt time.Time       `json:"created_at"`
// 	UpdatedAt time.Time       `json:"updated_at"`
// 	Order     *models.Order   `json:"order,omitempty"`
// 	Printer   PrinterResponse `json:"printer"`
// }

type PrintJobResponse struct {
	ID        uint            `json:"id"`
	PrinterID uint            `json:"printer_id"`
	OrderID   *uint           `json:"order_id,omitempty"`
	Content   []byte          `json:"content"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Order     *models.Order   `json:"order,omitempty"`
	Receipt   *models.Receipt `json:"receipt,omitempty"`
	Printer   PrinterResponse `json:"printer"`
	JobType   string          `json:"job_type"`
}

// @Summary รับงานพิมพ์ที่รอดำเนินการ
// @Description ดึงรายการงานพิมพ์ที่ยังไม่ได้พิมพ์สำหรับเครื่องพิมพ์ที่ระบุ (รองรับทั้ง IP และ USB)
// @Produce json
// @Param printer_ip query string false "Printer IP Address (สำหรับเครื่องพิมพ์เครือข่าย)"
// @Param vendor_id query string false "Vendor ID (สำหรับเครื่องพิมพ์ USB)"
// @Param product_id query string false "Product ID (สำหรับเครื่องพิมพ์ USB)"
// @Success 200 {array} models.PrintJob
// @Router /api/printers/pending-jobs [get]
// @Tags Printer
func GetPendingPrintJobs(c *fiber.Ctx) error {
	// ตรวจสอบ params
	printerIP := c.Query("printer_ip")
	vendorID := c.Query("vendor_id")
	productID := c.Query("product_id")

	var printer models.Printer
	var err error

	if vendorID != "" && productID != "" {
		// กรณีเครื่องพิมพ์ USB
		// log.Printf("Looking for USB printer - VID: %s, PID: %s", vendorID, productID)
		// log.Printf("Looking for USB printer - VID: %s, PID: %s", vendorID, productID)
		err = db.DB.Where("type = ? AND vendor_id = ? AND product_id = ?",
			"usb", vendorID, productID).First(&printer).Error
	} else if printerIP != "" {
		// กรณีเครื่องพิมพ์เครือข่าย
		// log.Printf("Looking for Network printer - IP: %s", printerIP)
		// log.Printf("Looking for Network printer - IP: %s", printerIP)

		// ถ้า printerIP เริ่มต้นด้วย USB_ ให้ค้นหาจาก vendor_id และ product_id
		if strings.HasPrefix(printerIP, "USB_") {
			parts := strings.Split(printerIP[4:], "_")
			if len(parts) == 2 {
				vid, pid := parts[0], parts[1]
				err = db.DB.Where("type = ? AND vendor_id = ? AND product_id = ?",
					"usb", vid, pid).First(&printer).Error
			} else {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error": "Invalid USB printer identifier format",
				})
			}
		} else {
			// ค้นหาจาก IP address สำหรับ network printer
			err = db.DB.Where("type = ? AND ip_address = ?",
				"network", printerIP).First(&printer).Error
		}
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Missing printer identification parameters",
		})
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.Printf("Printer not found: %v", err)
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Printer not found",
				"details": err.Error(),
			})
		} else {
			log.Printf("Database error: %v", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Database error",
				"details": err.Error(),
			})
		}
	}

	// ดึงงานพิมพ์ที่รอดำเนินการ
	var jobs []models.PrintJob
	// err = db.DB.Where("printer_id = ? AND status = ?", printer.ID, "pending").
	// 	Preload("Order").
	// 	Preload("Order.Items", "status = ?", "pending").
	// 	Preload("Order.Items.MenuItem").
	// 	Preload("Order.Items.Options").
	// 	Preload("Order.Items.Options.MenuOption").
	// 	Preload("Order.Items.Options.MenuOption.OptionGroup").
	// 	Preload("Receipt").
	// 	Preload("Receipt.Orders.Items.MenuItem").
	// 	Preload("Receipt.Discounts.DiscountType").
	// 	Preload("Receipt.Charges.ChargeType").
	// 	Order("created_at ASC").
	// 	Find(&jobs).Error

	query := db.DB.Where("printer_id = ? AND status = ?", printer.ID, "pending").
		Order("created_at ASC")

	err = query.Preload("Order.Items.MenuItem").
		Preload("Order.Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", "pending")
		}).
		Preload("Order.Items.Options.MenuOption.OptionGroup").
		Preload("Receipt.Orders.Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("status != ?", "cancelled")
		}).
		Preload("Receipt.Orders.Items.MenuItem").
		Preload("Receipt.Orders.Items.Options.MenuOption").
		Preload("Receipt.Discounts.DiscountType").
		Preload("Receipt.Charges.ChargeType").
		Find(&jobs).Error

	if err != nil {
		log.Printf("Error fetching print jobs: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch print jobs",
			"details": err.Error(),
		})
	}

	// เตรียม response
	var responses []PrintJobResponse
	for _, job := range jobs {
		var preparedContent []byte
		var err error

		switch job.JobType {
		case "order":
			if job.OrderID != nil {
				preparedContent, err = prepareOrderPrintContent(job)
			}
		case "receipt":
			if job.ReceiptID != nil {
				preparedContent, err = prepareReceiptPrintContent(job)
			}
		case "cancelation":
			preparedContent, err = prepareCancelPrintContent(job)
		case "qr_code":
			preparedContent = job.Content
		default:
			preparedContent = job.Content
		}

		if err != nil {
			log.Printf("Failed to prepare content for job %d: %v", job.ID, err)
			continue
		}
		bitmapImage, err := convertToBitmap(preparedContent)
		if err != nil {
			log.Printf("Failed to convert job %d to bitmap %v", job.ID, err)
			continue
		}
		response := PrintJobResponse{
			ID:        job.ID,
			PrinterID: printer.ID,
			OrderID:   job.OrderID,
			Content:   bitmapImage,
			Status:    job.Status,
			CreatedAt: job.CreatedAt,
			UpdatedAt: job.UpdatedAt,
			Order:     job.Order,
			Printer: PrinterResponse{
				ID:         printer.ID,
				Name:       printer.Name,
				Type:       printer.Type,
				IPAddress:  printer.IPAddress,
				Port:       printer.Port,
				VendorID:   printer.VendorID,
				ProductID:  printer.ProductID,
				PaperSize:  printer.PaperSize,
				Status:     printer.Status,
				Categories: printer.Categories,
			},
		}

		responses = append(responses, response)
	}

	return c.JSON(responses)
}

func findAppropiatePrinter(menuItem models.MenuItem) (*models.Printer, error) {
	// หาเครื่องพิมพ์ที่รองรับหมวดหมู่นี้
	var printer models.Printer
	err := db.DB.Joins("JOIN printer_categories ON printers.id = printer_categories.printer_id").
		Where("printer_categories.category_id = ?", menuItem.CategoryID).
		First(&printer).Error

	if err != nil {
		// ถ้าไม่พบเครื่องพิมพ์ที่รองรับหมวดหมู่นี้ ใช้เครื่องพิมพ์ main
		err = db.DB.Where("name = ?", "main").First(&printer).Error
		if err != nil {
			return nil, errors.New("no printer available for this category")
		}
	}

	return &printer, nil
}

// ฟังก์ชันแปลงเนื้อหาเป็น bitmap
func convertToBitmap(content []byte) ([]byte, error) {

	if isPNG(content) {
		return convertPNGToBitmap(content)
	}
	// กำหนดค่าพื้นฐาน
	width := 576
	dpi := 203.0
	fontSize := 18.0
	lineSpacing := 3.0
	leftPadding := 10
	printableWidth := width - (leftPadding * 2)

	// โหลดฟอนต์
	fontBytes, err := os.ReadFile("THSarabunNew.ttf")
	if err != nil {
		return nil, fmt.Errorf("error loading font: %v", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing font: %v", err)
	}

	// สร้าง font.Face ด้วย hinting ที่เหมาะสมสำหรับภาษาไทย
	face := truetype.NewFace(f, &truetype.Options{
		Size:    fontSize,
		DPI:     dpi,
		Hinting: font.HintingFull,
	})
	defer face.Close()

	d := &font.Drawer{
		Dst:  nil,
		Src:  image.Black,
		Face: face,
	}

	// เตรียม drawer สำหรับวัดความกว้าง
	tempImg := image.NewRGBA(image.Rect(0, 0, width, 1000))
	d.Dst = tempImg

	var allLines []string
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		text := scanner.Text()
		wrappedLines := segmentText(text, printableWidth, d)
		allLines = append(allLines, wrappedLines...)
	}

	// คำนวณความสูงของภาพ
	totalLines := len(allLines)
	height := int(float64(totalLines)*fontSize*lineSpacing) + 100

	// สร้างภาพใหม่
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	d.Dst = img

	// วาดข้อความ
	y := int(fontSize * 3.0)
	for _, text := range allLines {
		d.Dot = fixed.Point26_6{
			X: fixed.I(leftPadding),
			Y: fixed.I(y),
		}
		d.DrawString(text)
		y += int(fontSize * lineSpacing)
	}

	var buf bytes.Buffer

	// Initialize printer
	buf.Write([]byte{0x1B, 0x40})       // Initialize
	buf.Write([]byte{0x1D, 0x21, 0x00}) // Normal size
	buf.Write([]byte{0x1B, 0x4D, 0x00}) // Font A
	buf.Write([]byte{0x1B, 0x33, 60})   // Line spacing
	buf.Write([]byte{0x1D, 0x7C, 0x08}) // Highest density

	// Set bitmap mode
	buf.Write([]byte{0x1D, 0x76, 0x30, 0x00})

	// ขนาด bitmap
	widthBytes := (width + 7) / 8
	buf.WriteByte(byte(widthBytes & 0xFF))
	buf.WriteByte(byte(widthBytes >> 8))
	buf.WriteByte(byte(height & 0xFF))
	buf.WriteByte(byte(height >> 8))

	// แปลงเป็น bitmap ด้วย threshold ที่ปรับแล้ว
	for y := 0; y < height; y++ {
		for x := 0; x < width; x += 8 {
			var b byte
			for bit := 0; bit < 8; bit++ {
				if x+bit < width {
					r, g, b_, _ := img.At(x+bit, y).RGBA()
					// คำนวณค่าความสว่างแบบ weighted
					brightness := (r*299 + g*587 + b_*114) / 1000
					// ปรับ threshold ให้เหมาะสม
					if brightness < 0x7FFF { // ลดลงจาก 0x9FFF
						b |= 1 << (7 - bit)
					}
				}
			}
			buf.WriteByte(b)
		}
	}

	// Feed and cut
	buf.Write([]byte{0x1D, 0x56, 0x41, 0x03}) // Partial cut

	return buf.Bytes(), nil
}
func segmentText(text string, maxWidth int, d *font.Drawer) []string {
	var segments []string
	currentSegment := ""

	// แยกข้อความออกเป็นคำๆ
	words := strings.Fields(text)

	for _, word := range words {
		// ทดสอบความยาวของประโยคปัจจุบัน
		testSegment := currentSegment + word
		if currentSegment != "" {
			testSegment = currentSegment + " " + word
		}

		// วัดความกว้างของข้อความ
		testWidth := d.MeasureString(testSegment).Ceil()

		if testWidth > maxWidth {
			// ถ้าความยาวเกินที่กำหนด ให้บันทึกส่วนปัจจุบัน
			if currentSegment != "" {
				segments = append(segments, strings.TrimSpace(currentSegment))
			}
			currentSegment = word
		} else {
			// ถ้าสามารถใส่คำได้ ให้เพิ่มคำ
			if currentSegment == "" {
				currentSegment = word
			} else {
				currentSegment += " " + word
			}
		}
	}

	// เพิ่มส่วนสุดท้าย
	if currentSegment != "" {
		segments = append(segments, strings.TrimSpace(currentSegment))
	}

	return segments
}

// @Summary อัพเดทสถานะงานพิมพ์
// @Description อัพเดทสถานะของงานพิมพ์ (completed/failed)
// @Accept json
// @Produce json
// @Param id path int true "Print Job ID"
// @Param status body string true "New Status"
// @Success 200 {object} models.PrintJob
// @Router /api/printers/status/{id} [put]
// @Tags Printer
func UpdatePrintJobStatus(c *fiber.Ctx) error {
	jobID := c.Params("id")
	var req struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if req.Status != "completed" && req.Status != "failed" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status. Must be 'completed' or 'failed'",
		})
	}

	var job models.PrintJob
	if err := db.DB.Where("id = ?", jobID).First(&job).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Print job not found",
		})
	}

	// if req.Status == "failed" {
	// 	job.Status = "pending"
	// } else {
	// 	job.Status = req.Status
	// }
	job.Status = req.Status
	if err := db.DB.Save(&job).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update job status",
		})
	}

	return c.JSON(job)
}

func findPrinterByIdentifier(printerID string) (*models.Printer, error) {
	var printer models.Printer
	var err error

	if strings.HasPrefix(printerID, "USB_") {
		// สำหรับเครื่องพิมพ์ USB
		parts := strings.Split(printerID[4:], "_")
		if len(parts) != 2 {
			return nil, errors.New("invalid USB printer ID format")
		}
		vendorID, productID := parts[0], parts[1]

		err = db.DB.Where(
			"type = ? AND vendor_id = ? AND product_id = ?",
			"usb", vendorID, productID,
		).First(&printer).Error
	} else {
		// สำหรับเครื่องพิมพ์เครือข่าย
		err = db.DB.Where(
			"type = ? AND ip_address = ?",
			"network", printerID,
		).First(&printer).Error
	}

	if err != nil {
		return nil, err
	}
	return &printer, nil
}

func getPendingJobs(printer *models.Printer, categories []models.Category) ([]models.PrintJob, error) {
	categoryIDs := make(map[uint]bool)
	for _, cat := range categories {
		categoryIDs[cat.ID] = true
	}

	var jobs []models.PrintJob
	err := db.DB.Where("printer_id = ? AND status = ?", printer.ID, "pending").
		Preload("Order").
		Preload("Order.Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("status = ?", "pending")
		}).
		Preload("Order.Items.MenuItem").
		Preload("Order.Items.MenuItem.Category").
		Preload("Order.Items.Options").
		Preload("Order.Items.Options.MenuOption").
		Preload("Order.Items.Options.MenuOption.OptionGroup").
		Find(&jobs).Error

	if err != nil {
		return nil, err
	}

	// กรองงานตามหมวดหมู่
	var filteredJobs []models.PrintJob
	for _, job := range jobs {
		if shouldPrintJob(job, categoryIDs, printer.Name) {
			filteredJobs = append(filteredJobs, job)
		}
	}

	return filteredJobs, nil
}

func shouldPrintJob(job models.PrintJob, categoryIDs map[uint]bool, printerName string) bool {
	if job.Order == nil {
		return true // สำหรับงานพิมพ์ทั่วไป
	}

	for _, item := range job.Order.Items {
		if item.MenuItem.CategoryID > 0 && categoryIDs[item.MenuItem.CategoryID] {
			return true
		}
	}

	// ใช้ main printer เป็น fallback
	if printerName == "main" {
		var mainPrinter models.Printer
		if db.DB.Where("name = ?", "main").First(&mainPrinter).Error == nil {
			return true
		}
	}

	return false
}

// CategoryResponse สำหรับ swagger
type CategoryResponse struct {
	ID        uint      `json:"id" example:"1"`
	Name      string    `json:"name" example:"ครัวร้อน"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// @Summary ดึงรายการหมวดหมู่ของเครื่องพิมพ์
// @Description ดึงรายการหมวดหมู่ที่เครื่องพิมพ์สามารถพิมพ์ได้
// @Produce json
// @Param id path int true "Printer ID"
// @Success 200 {array} models.Category
// @Router /api/printers/categories/{id} [get]
// @Tags Printer
func GetPrinterCategoriesById(c *fiber.Ctx) error {
	printerID := c.Params("id")

	var printer models.Printer
	if err := db.DB.Preload("Categories").First(&printer, printerID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Printer not found",
		})
	}

	return c.JSON(printer.Categories)
}

func isPNG(data []byte) bool {
	// PNG signature
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	return len(data) > 8 && bytes.Equal(data[:8], pngSignature)
}

func convertPNGToBitmap(pngData []byte) ([]byte, error) {
	// อ่านรูป PNG
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %v", err)
	}

	// กำหนดขนาด
	width := 576
	height := img.Bounds().Dy() * width / img.Bounds().Dx()

	// สร้างภาพ bitmap ใหม่
	bitmapImg := image.NewRGBA(image.Rect(0, 0, width, height))

	// ใช้ nearest neighbor scaling เพื่อรักษาคุณภาพ QR Code
	draw.NearestNeighbor.Scale(bitmapImg, bitmapImg.Bounds(), img, img.Bounds(), draw.Over, nil)

	var buf bytes.Buffer

	// Initialize printer
	buf.Write([]byte{0x1B, 0x40})       // Initialize
	buf.Write([]byte{0x1D, 0x21, 0x00}) // Normal size
	buf.Write([]byte{0x1B, 0x4D, 0x00}) // Font A
	buf.Write([]byte{0x1B, 0x33, 60})   // Line spacing
	buf.Write([]byte{0x1D, 0x7C, 0x08}) // Highest density

	// Set bitmap mode
	buf.Write([]byte{0x1D, 0x76, 0x30, 0x00})

	// ขนาด bitmap
	widthBytes := (width + 7) / 8
	buf.WriteByte(byte(widthBytes & 0xFF))
	buf.WriteByte(byte(widthBytes >> 8))
	buf.WriteByte(byte(height & 0xFF))
	buf.WriteByte(byte(height >> 8))

	// แปลงเป็น bitmap ด้วยการใช้ threshold สูง
	for y := 0; y < height; y++ {
		for x := 0; x < width; x += 8 {
			var b byte
			for bit := 0; bit < 8; bit++ {
				if x+bit < width {
					r, g, b_, _ := bitmapImg.At(x+bit, y).RGBA()

					// สำหรับ QR Code ใช้ threshold ต่ำลง เพื่อให้ได้ภาพคมชัด
					brightness := (r*299 + g*587 + b_*114) / 1000
					if brightness < 0x5FFF { // ลดลงจากเดิม
						b |= 1 << (7 - bit)
					}
				}
			}
			buf.WriteByte(b)
		}
	}

	// Feed and cut
	buf.Write([]byte{0x1D, 0x56, 0x41, 0x03}) // Partial cut

	return buf.Bytes(), nil
}

// @Summary ดึงรายการงานพิมพ์ที่สามารถรีปริ้นได้
// @Description ดึงรายการงานพิมพ์ที่สามารถรีปริ้นได้ โดยสามารถกรองตามประเภทและช่วงเวลาได้
// @Produce json
// @Param type query string false "ประเภทงานพิมพ์ (order, receipt, cancelation, qr_code)"
// @Param start_date query string false "วันที่เริ่มต้น (YYYY-MM-DD)"
// @Param end_date query string false "วันที่สิ้นสุด (YYYY-MM-DD)"
// @Success 200 {array} PrintJobResponse
// @Router /api/printers/reprintable-jobs [get]
// @Tags Printer
// @Summary ดึงรายการงานพิมพ์ที่สามารถรีปริ้นได้
// @Description ดึงรายการงานพิมพ์ที่สามารถรีปริ้นได้ โดยสามารถกรองตามประเภทและช่วงเวลาได้
// @Produce json
// @Param type query string false "ประเภทงานพิมพ์ (order, receipt, cancelation, qr_code)"
// @Param start_date query string false "วันที่เริ่มต้น (YYYY-MM-DD)"
// @Param end_date query string false "วันที่สิ้นสุด (YYYY-MM-DD)"
// @Success 200 {array} PrintJobResponse
// @Router /api/printers/reprintable-jobs [get]
// @Tags Printer
func GetReprintableJobs(c *fiber.Ctx) error {
	jobType := c.Query("type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := db.DB.Model(&models.PrintJob{}).
		Preload("Printer").
		// Order และข้อมูลที่เกี่ยวข้อง
		Preload("Order.Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Order.Items.MenuItem").
		Preload("Order.Items.MenuItem.Category").
		Preload("Order.Items.Options.MenuOption").
		Preload("Order.Items.Options.MenuOption.OptionGroup").
		// Receipt และข้อมูลที่เกี่ยวข้อง
		Preload("Receipt").
		Preload("Receipt.Staff"). // Staff relation มีแค่ใน Receipt
		Preload("Receipt.Orders", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Receipt.Orders.Items", func(db *gorm.DB) *gorm.DB {
			return db.Where("status != ?", "cancelled").Order("id ASC")
		}).
		Preload("Receipt.Orders.Items.MenuItem").
		Preload("Receipt.Orders.Items.MenuItem.Category").
		Preload("Receipt.Orders.Items.Options.MenuOption").
		Preload("Receipt.Orders.Items.Options.MenuOption.OptionGroup").
		Preload("Receipt.Discounts.DiscountType").
		Preload("Receipt.Charges.ChargeType")

	if jobType != "" {
		query = query.Where("job_type = ?", jobType)
	}

	if startDate != "" {
		startTime, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid start date format"})
		}
		query = query.Where("created_at >= ?", startTime)
	}

	if endDate != "" {
		endTime, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Invalid end date format"})
		}
		endTime = endTime.Add(24 * time.Hour)
		query = query.Where("created_at < ?", endTime)
	}

	query = query.Where("status = ?", "completed").Order("created_at DESC")

	var jobs []models.PrintJob
	if err := query.Find(&jobs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch print jobs",
			"details": err.Error(),
		})
	}

	var responses []PrintJobResponse
	for _, job := range jobs {
		// Prepare print content based on job type
		var content []byte
		var err error
		switch job.JobType {
		case "order":
			if job.OrderID != nil {
				content, err = prepareOrderPrintContent(job)
			}
		case "receipt":
			if job.ReceiptID != nil {
				content, err = prepareReceiptPrintContent(job)
			}
		case "cancelation":
			content, err = prepareCancelPrintContent(job)
		case "qr_code":
			content = job.Content
		default:
			content = job.Content
		}

		if err != nil {
			log.Printf("Failed to prepare content for job %d: %v", job.ID, err)
			continue
		}

		response := PrintJobResponse{
			ID:        job.ID,
			PrinterID: job.PrinterID,
			OrderID:   job.OrderID,
			Content:   content,
			Status:    job.Status,
			JobType:   job.JobType,
			CreatedAt: job.CreatedAt,
			UpdatedAt: job.UpdatedAt,
			Order:     job.Order,
			Receipt:   job.Receipt,
			Printer: PrinterResponse{
				ID:         job.Printer.ID,
				Name:       job.Printer.Name,
				Type:       job.Printer.Type,
				IPAddress:  job.Printer.IPAddress,
				Port:       job.Printer.Port,
				VendorID:   job.Printer.VendorID,
				ProductID:  job.Printer.ProductID,
				PaperSize:  job.Printer.PaperSize,
				Status:     job.Printer.Status,
				Categories: job.Printer.Categories,
			},
		}

		responses = append(responses, response)
	}

	return c.JSON(responses)
}

// @Summary รีปริ้นเอกสาร
// @Description สั่งพิมพ์เอกสารซ้ำ (รองรับ order, receipt, qr_code)
// @Accept json
// @Produce json
// @Param id path int true "Print Job ID"
// @Success 200 {object} models.PrintJob
// @Router /api/printers/reprint/{id} [post]
// @Tags Printer
func ReprintDocument(c *fiber.Ctx) error {
	jobID := c.Params("id")

	var originalJob models.PrintJob
	if err := db.DB.Preload("Printer").
		Preload("Order.Items.MenuItem").
		Preload("Order.Items.Options.MenuOption.OptionGroup").
		Preload("Receipt.Orders.Items.MenuItem").
		Preload("Receipt.Orders.Items.Options.MenuOption").
		Preload("Receipt.Discounts.DiscountType").
		Preload("Receipt.Charges.ChargeType").
		First(&originalJob, jobID).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Print job not found"})
	}

	newJob := models.PrintJob{
		PrinterID: originalJob.PrinterID,
		OrderID:   originalJob.OrderID,
		ReceiptID: originalJob.ReceiptID,
		JobType:   originalJob.JobType,
		Status:    "pending",
	}

	var err error
	switch originalJob.JobType {
	case "order":
		newJob.Content, err = prepareOrderPrintContent(originalJob)
	case "receipt":
		newJob.Content, err = prepareReceiptPrintContent(originalJob)
	case "cancelation":
		newJob.Content, err = prepareCancelPrintContent(originalJob)
	case "qr_code":
		newJob.Content = originalJob.Content
	default:
		newJob.Content = originalJob.Content
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to prepare print content", "details": err.Error()})
	}

	if err := db.DB.Create(&newJob).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to create reprint job", "details": err.Error()})
	}

	bitmapContent, err := convertToBitmap(newJob.Content)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to convert content to bitmap", "details": err.Error()})
	}

	response := PrintJobResponse{
		ID:        newJob.ID,
		PrinterID: newJob.PrinterID,
		OrderID:   newJob.OrderID,
		Content:   bitmapContent,
		Status:    newJob.Status,
		CreatedAt: newJob.CreatedAt,
		UpdatedAt: newJob.UpdatedAt,
		Order:     newJob.Order,
		Printer: PrinterResponse{
			ID:         originalJob.Printer.ID,
			Name:       originalJob.Printer.Name,
			Type:       originalJob.Printer.Type,
			IPAddress:  originalJob.Printer.IPAddress,
			Port:       originalJob.Printer.Port,
			VendorID:   originalJob.Printer.VendorID,
			ProductID:  originalJob.Printer.ProductID,
			PaperSize:  originalJob.Printer.PaperSize,
			Status:     originalJob.Printer.Status,
			Categories: originalJob.Printer.Categories,
		},
	}
	return c.JSON(response)
}
