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
	"math"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	api_v2 "food-ordering-api/api_v2"

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
	if len(req.CategoryIDs) > 0 {
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
	}

	// ลบความสัมพันธ์เก่าทั้งหมด
	if err := tx.Exec("DELETE FROM printer_categories WHERE printer_id = ?", printer.ID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to remove old category assignments",
		})
	}

	// เพิ่มความสัมพันธ์ใหม่
	for _, catID := range req.CategoryIDs {
		if err := tx.Exec("INSERT INTO printer_categories (printer_id, category_id) VALUES (?, ?)",
			printer.ID, catID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to assign categories",
			})
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

		// สร้าง map เพื่อจัดกลุ่มรายการที่เหมือนกันทุกประการ
		type OrderItemKey struct {
			MenuItemID uint
			Notes      string
			Options    string // จะเก็บ options ในรูปแบบ string ที่เรียงลำดับแล้ว
		}

		itemGroups := make(map[OrderItemKey]struct {
			MenuItem models.MenuItem
			Quantity int
			Price    float64
			Options  []models.OrderItemOption
			Notes    string
		})

		// จัดกลุ่มรายการที่เหมือนกันทุกประการ
		for _, order := range job.Receipt.Orders {
			for _, item := range order.Items {
				if item.Status != "cancelled" {
					// สร้าง options string ที่เรียงลำดับแล้ว
					var optStrings []string
					for _, opt := range item.Options {
						optStrings = append(optStrings, fmt.Sprintf("%d:%d:%.2f",
							opt.MenuOptionID, opt.Quantity, opt.Price))
					}
					sort.Strings(optStrings)
					optionsStr := strings.Join(optStrings, "|")

					key := OrderItemKey{
						MenuItemID: item.MenuItemID,
						Notes:      item.Notes,
						Options:    optionsStr,
					}

					if group, exists := itemGroups[key]; exists {
						group.Quantity += item.Quantity
						group.Price += item.Price * float64(item.Quantity)
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
							Price:    item.Price * float64(item.Quantity),
							Options:  item.Options,
							Notes:    item.Notes,
						}
					}
				}
			}
		}

		// แปลง map เป็น slice เพื่อเรียงลำดับ
		type GroupedItem struct {
			MenuItem models.MenuItem
			Quantity int
			Price    float64
			Options  []models.OrderItemOption
			Notes    string
		}

		var groupedItems []GroupedItem
		for _, group := range itemGroups {
			groupedItems = append(groupedItems, GroupedItem{
				MenuItem: group.MenuItem,
				Quantity: group.Quantity,
				Price:    group.Price,
				Options:  group.Options,
				Notes:    group.Notes,
			})
		}

		// เรียงลำดับตามชื่อเมนู
		sort.Slice(groupedItems, func(i, j int) bool {
			return groupedItems[i].MenuItem.Name < groupedItems[j].MenuItem.Name
		})

		// พิมพ์รายการที่จัดกลุ่มแล้ว
		for i, group := range groupedItems {
			itemLine := fmt.Sprintf("%d. %s", i+1, group.MenuItem.Name)
			if group.Quantity > 1 {
				itemLine += fmt.Sprintf(" x%d", group.Quantity)
			}
			itemLine += fmt.Sprintf("   ฿%.2f", group.Price)
			content.WriteString(cleanText(itemLine) + "\n")

			// พิมพ์ options
			for _, opt := range group.Options {
				optionLine := fmt.Sprintf("   • %s   ฿%.2f",
					cleanText(opt.MenuOption.Name),
					opt.Price)
				content.WriteString(optionLine + "\n")
			}

			// พิมพ์โน้ต
			if group.Notes != "" {
				content.WriteString(fmt.Sprintf("   [หมายเหตุ: %s]\n", cleanText(group.Notes)))
			}
		}

		content.WriteString("----------------------------------------\n")

		// พิมพ์สรุปยอด
		subTotalLine := fmt.Sprintf("ยอดรวม: ฿%.2f", job.Receipt.SubTotal)
		discountLine := fmt.Sprintf("ส่วนลด: ฿%.2f", job.Receipt.DiscountTotal)
		extraChargesLine := fmt.Sprintf("ค่าใช้จ่ายเพิ่มเติม: ฿%.2f", job.Receipt.ChargeTotal)
		VatLine := fmt.Sprintf("Vat 7%%: ฿%.2f", job.Receipt.ServiceCharge)
		totalLine := fmt.Sprintf("ยอดสุทธิ: ฿%.2f", job.Receipt.Total)

		summaryLines := []string{
			subTotalLine,
			discountLine,
			extraChargesLine,
			VatLine,
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

func prepareQRCodePrintContent(job models.PrintJob, png []byte) ([]byte, error) {
	var content bytes.Buffer

	// เพิ่มส่วนหัว
	headerLines := []string{
		"===== QR Code สำหรับโต๊ะ ",
		fmt.Sprint("1 ====="),
		fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05")),
		"========================",
	}

	for _, line := range headerLines {
		content.WriteString(cleanText(line) + "\n")
	}

	// เพิ่มบรรทัดว่างสำหรับ spacing ก่อน QR Code
	content.WriteString("\n\n")

	// เตรียมข้อมูล PNG สำหรับการพิมพ์
	var pngContent []byte
	if isPNG(png) {
		pngContent = png
	}

	// เตรียมข้อความ
	textContent := content.Bytes()

	// รวมข้อความและ PNG
	var finalContent []byte
	finalContent = append(finalContent, textContent...)

	if pngContent != nil {
		finalContent = append(finalContent, pngContent...)
	}

	// เพิ่มส่วนท้าย
	footerLines := []string{
		"\n========================",
		"โปรดสแกน QR Code เพื่อเข้าสู่ระบบ",
		"========================",
	}

	for _, line := range footerLines {
		content.WriteString(cleanText(line) + "\n")
	}

	finalContent = append(finalContent, content.Bytes()...)

	return finalContent, nil
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
	JobType   string          `json:"job_type"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Order     *models.Order   `json:"order,omitempty"`
	Receipt   *models.Receipt `json:"receipt,omitempty"`
	Printer   PrinterResponse `json:"printer"`
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
	printerIP := c.Query("printer_ip")
	vendorID := c.Query("vendor_id")
	productID := c.Query("product_id")

	var printer models.Printer
	var err error

	if vendorID != "" && productID != "" {
		err = db.DB.Where("type = ? AND vendor_id = ? AND product_id = ?",
			"usb", vendorID, productID).First(&printer).Error
	} else if printerIP != "" {
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
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error":   "Printer not found",
				"details": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Database error",
			"details": err.Error(),
		})
	}

	var jobs []models.PrintJob
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch print jobs",
			"details": err.Error(),
		})
	}

	// กำหนดจำนวน workers
	numWorkers := 10

	// สร้าง channels
	jobsChan := make(chan models.PrintJob, len(jobs))
	resultsChan := make(chan PrintJobResponse, len(jobs))

	// สร้าง WaitGroup สำหรับ workers
	var wg sync.WaitGroup

	// เริ่ม workers
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for job := range jobsChan {
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
					log.Printf("Failed to convert job %d to bitmap: %v", job.ID, err)
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
					Receipt:   job.Receipt,
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

				resultsChan <- response
			}
		}()
	}

	// ส่งงานเข้า jobs channel
	for _, job := range jobs {
		jobsChan <- job
	}
	close(jobsChan)

	// รอให้ workers ทำงานเสร็จ
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// รวบรวมผลลัพธ์
	var responses []PrintJobResponse
	for response := range resultsChan {
		responses = append(responses, response)
	}

	// เรียงลำดับตาม CreatedAt
	sort.Slice(responses, func(i, j int) bool {
		return responses[i].CreatedAt.Before(responses[j].CreatedAt)
	})

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
	// เพิ่ม log เพื่อตรวจสอบ
	log.Printf("Decoding PNG with length: %d", len(pngData))

	// อ่านรูป PNG
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		log.Printf("PNG Decode Error: %v", err)
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
	// เพิ่ม query params สำหรับ pagination
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	jobType := c.Query("type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	// สร้าง base query พร้อม preload ที่จำเป็น
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
		Preload("Receipt.Staff").
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

	// เพิ่มเงื่อนไขการค้นหา
	if jobType != "" {
		if jobType == "others" {
			// กรณีเลือก "อื่นๆ"
			standardTypes := []string{"order", "receipt", "cancelation", "qr_code", "shift_report"}
			query = query.Where("job_type NOT IN (?)", standardTypes).
				Where("order_id IS NULL").
				Where("receipt_id IS NULL").
				Where("job_type NOT LIKE ?", "order%").   // ป้องกันกรณี job_type มีคำว่า order
				Where("job_type NOT LIKE ?", "receipt%"). // ป้องกันกรณี job_type มีคำว่า receipt
				Where("job_type NOT LIKE ?", "cancel%")   // ป้องกันกรณี job_type มีคำว่า cancel
		} else {
			// กรณีเลือกประเภทอื่นๆ
			query = query.Where("job_type = ?", jobType)
		}
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

	// นับจำนวนรายการทั้งหมดก่อนทำ pagination
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to count total records",
		})
	}

	// เพิ่ม pagination และการเรียงลำดับ
	query = query.Where("status = ?", "completed").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit)

	var jobs []models.PrintJob
	if err := query.Find(&jobs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error":   "Failed to fetch print jobs",
			"details": err.Error(),
		})
	}

	// แปลงข้อมูลเป็น response
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
				content, err = api_v2.PrepareReceiptPrintContent(job)
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

	// ส่งข้อมูลกลับพร้อม pagination info
	return c.JSON(fiber.Map{
		"data": responses,
		"pagination": fiber.Map{
			"current_page": page,
			"total_pages":  int(math.Ceil(float64(total) / float64(limit))),
			"total_items":  total,
			"per_page":     limit,
			"has_more":     (page*limit) < int(total) && len(responses) == limit,
		},
	})
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
		JobType:   newJob.JobType,
		CreatedAt: newJob.CreatedAt,
		UpdatedAt: newJob.UpdatedAt,
		Order:     newJob.Order,
		Receipt:   newJob.Receipt,
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

// ฟังก์ชันสำหรับพิมพ์ใบรายการอาหารก่อนชำระเงิน
func prepareBillCheckPrintContent(orders []models.Order, tableIDs []string, discounts []PaymentDiscountRequest, extraCharges []PaymentExtraChargeRequest, serviceChargePercent float64) ([]byte, error) {
	var content bytes.Buffer

	// ส่วนหัว
	headerLines := []string{
		"***** ใบรายการอาหาร *****",
		fmt.Sprintf("โต๊ะ: %s", strings.Join(tableIDs, ", ")),
		"----------------------------------------",
		fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05")),
		"----------------------------------------",
		"[ รายการอาหาร ]",
		"----------------------------------------",
	}

	for _, line := range headerLines {
		content.WriteString(cleanText(line) + "\n")
	}

	// จัดกลุ่มรายการที่เหมือนกัน
	type OrderItemKey struct {
		MenuItemID uint
		Notes      string
		Options    string
	}

	type GroupedItem struct {
		MenuItem models.MenuItem
		Quantity int
		Price    float64
		Options  []models.OrderItemOption
		Notes    string
	}

	itemGroups := make(map[OrderItemKey]GroupedItem)

	// คำนวณยอดรวม
	var subTotal float64
	for _, order := range orders {
		for _, item := range order.Items {
			if item.Status != "cancelled" {
				var optStrings []string
				for _, opt := range item.Options {
					optStrings = append(optStrings, fmt.Sprintf("%d:%d:%.2f",
						opt.MenuOptionID, opt.Quantity, opt.Price))
				}
				sort.Strings(optStrings)
				optionsStr := strings.Join(optStrings, "|")

				key := OrderItemKey{
					MenuItemID: item.MenuItemID,
					Notes:      item.Notes,
					Options:    optionsStr,
				}

				// คำนวณราคารวมของรายการ
				itemTotal := float64(item.Quantity) * item.Price
				for _, opt := range item.Options {
					itemTotal += opt.Price * float64(opt.Quantity)
				}

				// ตรวจสอบและปรับราคาตามโปรโมชั่น
				if item.PromotionUsage != nil && item.PromotionUsage.Promotion.ID > 0 {
					itemTotal = item.PromotionUsage.Promotion.Price
				}

				if group, exists := itemGroups[key]; exists {
					group.Quantity += item.Quantity
					group.Price += itemTotal
					itemGroups[key] = group
				} else {
					itemGroups[key] = GroupedItem{
						MenuItem: item.MenuItem,
						Quantity: item.Quantity,
						Price:    itemTotal,
						Options:  item.Options,
						Notes:    item.Notes,
					}
				}

				subTotal += itemTotal
			}
		}
	}

	// แปลงเป็น slice และเรียงตามชื่อเมนู
	var groupedItems []GroupedItem
	for _, group := range itemGroups {
		groupedItems = append(groupedItems, group)
	}

	sort.Slice(groupedItems, func(i, j int) bool {
		return groupedItems[i].MenuItem.Name < groupedItems[j].MenuItem.Name
	})

	// พิมพ์รายการ
	for i, group := range groupedItems {
		itemLine := fmt.Sprintf("%d. %s", i+1, group.MenuItem.Name)
		if group.Quantity > 1 {
			itemLine += fmt.Sprintf(" x%d", group.Quantity)
		}
		itemLine += fmt.Sprintf("   ฿%.2f", group.Price)
		content.WriteString(cleanText(itemLine) + "\n")

		for _, opt := range group.Options {
			optionLine := fmt.Sprintf("   • %s   ฿%.2f",
				cleanText(opt.MenuOption.Name),
				opt.Price*float64(opt.Quantity))
			content.WriteString(optionLine + "\n")
		}

		if group.Notes != "" {
			content.WriteString(fmt.Sprintf("   [หมายเหตุ: %s]\n", cleanText(group.Notes)))
		}
	}

	content.WriteString("----------------------------------------\n")
	content.WriteString(fmt.Sprintf("ยอดรวม: ฿%.2f\n", subTotal))

	// คำนวณและแสดงส่วนลด
	var totalDiscount float64
	if len(discounts) > 0 {
		content.WriteString("ส่วนลด:\n")
		for _, discount := range discounts {
			var discountType models.DiscountType
			if err := db.DB.First(&discountType, discount.DiscountTypeID).Error; err != nil {
				continue
			}
			var discountAmount float64
			if discountType.Type == "percentage" {
				discountAmount = (subTotal * discountType.Value) / 100
			} else {
				discountAmount = discountType.Value
			}
			totalDiscount += discountAmount
			content.WriteString(fmt.Sprintf("- %s: ฿%.2f\n", discountType.Name, discountAmount))
		}
	}

	// คำนวณและแสดงค่าใช้จ่ายเพิ่มเติม
	var totalExtraCharge float64
	if len(extraCharges) > 0 {
		content.WriteString("ค่าใช้จ่ายเพิ่มเติม:\n")
		for _, charge := range extraCharges {
			var chargeType models.AdditionalChargeType
			if err := db.DB.First(&chargeType, charge.ChargeTypeID).Error; err != nil {
				continue
			}
			chargeAmount := chargeType.DefaultAmount * float64(charge.Quantity)
			totalExtraCharge += chargeAmount
			content.WriteString(fmt.Sprintf("+ %s (x%d): ฿%.2f\n", chargeType.Name, charge.Quantity, chargeAmount))
		}
	}

	// คำนวณ Service Charge
	// serviceChargeAmount := (subTotal * serviceChargePercent) / 100
	// if serviceChargePercent > 0 {
	// 	content.WriteString(fmt.Sprintf("ค่าบริการ %.1f%%: ฿%.2f\n", serviceChargePercent, serviceChargeAmount))
	// }

	// คำนวณ VAT
	subTotalAfterDiscount := subTotal - totalDiscount
	// vatAmount := (subTotalAfterDiscount + serviceChargeAmount + totalExtraCharge) * 0.07
	vatAmount := (subTotalAfterDiscount + totalExtraCharge) * 0.07
	content.WriteString(fmt.Sprintf("VAT 7%%: ฿%.2f\n", vatAmount))

	// แสดงยอดรวมสุทธิ
	// netTotal := subTotalAfterDiscount + serviceChargeAmount + totalExtraCharge + vatAmount
	netTotal := subTotalAfterDiscount + totalExtraCharge + vatAmount
	content.WriteString("----------------------------------------\n")
	content.WriteString(fmt.Sprintf("ยอดรวมสุทธิ: ฿%.2f\n", netTotal))
	content.WriteString("----------------------------------------\n")
	content.WriteString("** กรุณาตรวจสอบรายการให้ครบถ้วน **\n")
	content.WriteString("========================================\n")

	return content.Bytes(), nil
}

// @Summary พิมพ์ใบรายการอาหารก่อนชำระเงิน
// @Description พิมพ์ใบรายการอาหารสำหรับตรวจทานก่อนชำระเงิน สามารถพิมพ์ได้ทั้งแบบโต๊ะเดี่ยวและรวมโต๊ะ
// @Accept json
// @Produce json
// @Param request body PrintBillCheckRequest true "ข้อมูลสำหรับพิมพ์ใบรายการอาหาร"
// @Success 200 {object} map[string]interface{}
// @Router /api/printers/bill-check [post]
// @Tags Printer
type PrintBillCheckRequest struct {
	TableIDs      []uint                          `json:"table_ids" binding:"required,min=1"`
	ServiceCharge float64                         `json:"service_charge"` //ควรเป็นค่า VAT 7%
	Discounts     []api_v2.PrintBillCheckDiscount `json:"discounts,omitempty"`
	ExtraCharges  []api_v2.PrintBillCheckCharge   `json:"extra_charges,omitempty"`
}

func PrintBillCheck(c *fiber.Ctx) error {
	var req PrintBillCheckRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "รูปแบบข้อมูลไม่ถูกต้อง",
		})
	}

	// ตรวจสอบโต๊ะ
	var tables []models.Table
	if err := db.DB.Where("id IN ?", req.TableIDs).Find(&tables).Error; err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบข้อมูลโต๊ะบางโต๊ะ",
		})
	}

	if len(tables) != len(req.TableIDs) {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "พบข้อมูลโต๊ะไม่ครบตามที่ระบุ",
		})
	}

	// ดึงข้อมูลออเดอร์ที่ยังไม่ได้ชำระจากทุกโต๊ะ
	var allOrders []models.Order
	for _, tableID := range req.TableIDs {
		var orders []models.Order
		if err := db.DB.Preload("Items", "status != ?", "cancelled").
			Preload("Items.MenuItem").
			Preload("Items.MenuItem.Category").
			Preload("Items.Options.MenuOption").
			Preload("Items.PromotionUsage.Promotion").
			Where("table_id = ? AND status NOT IN (?, ?) AND receipt_id IS NULL",
				tableID, "completed", "cancelled").
			Find(&orders).Error; err != nil {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "ไม่พบออเดอร์สำหรับโต๊ะบางโต๊ะ",
			})
		}
		allOrders = append(allOrders, orders...)
	}

	if len(allOrders) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบรายการอาหารที่ต้องชำระเงิน",
		})
	}

	// คำนวณยอดรวม
	var subTotal float64
	for _, order := range allOrders {
		for _, item := range order.Items {
			// คำนวณราคาพื้นฐานของรายการ
			itemTotal := float64(item.Quantity) * item.Price

			// เพิ่มราคาตัวเลือกเสริม
			for _, opt := range item.Options {
				itemTotal += opt.Price * float64(opt.Quantity)
			}

			// ตรวจสอบและปรับราคาตามโปรโมชั่น
			if item.PromotionUsage != nil && item.PromotionUsage.Promotion.ID > 0 {
				// หากมีโปรโมชั่น ใช้ราคาโปรโมชั่นแทน
				itemTotal = item.PromotionUsage.Promotion.Price
			}

			subTotal += itemTotal
		}
	}

	// ดึงข้อมูลและคำนวณส่วนลด
	var totalDiscount float64
	var discountDetails []models.DiscountType
	for _, discount := range req.Discounts {
		var discountType models.DiscountType
		if err := db.DB.First(&discountType, discount.DiscountTypeID).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "ประเภทส่วนลดไม่ถูกต้อง",
			})
		}
		discountDetails = append(discountDetails, discountType)

		var discountAmount float64
		if discountType.Type == "percentage" {
			discountAmount = (subTotal * discountType.Value) / 100
		} else {
			discountAmount = discountType.Value
		}
		totalDiscount += discountAmount
	}

	// ดึงข้อมูลและคำนวณค่าใช้จ่ายเพิ่มเติม
	var totalExtraCharge float64
	var chargeDetails []models.AdditionalChargeType
	for _, charge := range req.ExtraCharges {
		var chargeType models.AdditionalChargeType
		if err := db.DB.First(&chargeType, charge.ChargeTypeID).Error; err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "ประเภทค่าใช้จ่ายเพิ่มเติมไม่ถูกต้อง",
			})
		}
		chargeDetails = append(chargeDetails, chargeType)

		chargeAmount := chargeType.DefaultAmount * float64(charge.Quantity)
		totalExtraCharge += chargeAmount
	}

	// คำนวณ VAT 7%
	subTotalAfterDiscount := subTotal - totalDiscount
	vatAmount := (subTotalAfterDiscount + totalExtraCharge) * 0.07

	// คำนวณยอดรวมสุทธิ
	netTotal := subTotalAfterDiscount + totalExtraCharge + vatAmount

	// ค้นหาเครื่องพิมพ์หลัก
	var printer models.Printer
	if err := db.DB.Where("name = ?", "main").First(&printer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่พบเครื่องพิมพ์หลัก",
		})
	}

	// สร้างเนื้อหาสำหรับพิมพ์
	tableNames := make([]string, len(tables))
	for i, table := range tables {
		tableNames[i] = table.Name
	}

	// แปลงข้อมูลส่วนลดและค่าใช้จ่ายเพิ่มเติมให้อยู่ในรูปแบบที่ถูกต้อง
	var discountsForPrint []api_v2.PrintBillCheckDiscount
	for _, discount := range req.Discounts {
		discountsForPrint = append(discountsForPrint, api_v2.PrintBillCheckDiscount{
			DiscountTypeID: discount.DiscountTypeID,
			Reason:         discount.Reason,
		})
	}

	var chargesForPrint []api_v2.PrintBillCheckCharge
	for _, charge := range req.ExtraCharges {
		chargesForPrint = append(chargesForPrint, api_v2.PrintBillCheckCharge{
			ChargeTypeID: charge.ChargeTypeID,
			Quantity:     charge.Quantity,
			Note:         charge.Note,
		})
	}

	content, err := api_v2.V2_prepareBillCheckPrintContent(allOrders, tableNames, discountsForPrint, chargesForPrint, req.ServiceCharge, subTotal, totalDiscount, totalExtraCharge, vatAmount, netTotal)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถสร้างเนื้อหาสำหรับพิมพ์ได้",
		})
	}

	// สร้าง print job
	printJob := models.PrintJob{
		PrinterID: printer.ID,
		Content:   content,
		JobType:   "bill_check",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := db.DB.Create(&printJob).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถสร้างงานพิมพ์ได้",
		})
	}

	return c.JSON(fiber.Map{
		"message":       "สร้างงานพิมพ์ใบรายการอาหารสำเร็จ",
		"job_id":        printJob.ID,
		"sub_total":     subTotal,
		"discount":      totalDiscount,
		"extra_charges": totalExtraCharge,
		"vat":           vatAmount,
		"net_total":     netTotal,
	})
}
