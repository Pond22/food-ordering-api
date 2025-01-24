package api_handlers

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"image"
	"image/draw"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"github.com/golang/freetype/truetype"
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
	Department string            `json:"department"`
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

	var printer models.Printer
	if err := tx.First(&printer, printerID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Printer not found",
		})
	}

	// ตรวจสอบและดึงหมวดหมู่ที่ระบุ
	var categories []models.Category
	if err := tx.Find(&categories, req.CategoryIDs).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid category IDs",
		})
	}

	if len(categories) != len(req.CategoryIDs) {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Some categories not found",
		})
	}

	// อัพเดทความสัมพันธ์
	if err := tx.Model(&printer).Association("Categories").Replace(&categories); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update printer categories",
		})
	}

	// โหลดข้อมูลที่อัพเดทแล้ว
	if err := tx.Preload("Categories").First(&printer, printerID).Error; err != nil {
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

	return c.JSON(printer)
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

	if job.Receipt != nil {
		// หัวข้อใบเสร็จ
		headerLines := []string{
			"***** ใบเสร็จรับเงิน *****",
			"Receipt #" + fmt.Sprintf("%d", job.Receipt.ID),
			"โต๊ะ: " + fmt.Sprintf("%d", job.Receipt.TableID),
			"----------------------------------------",
			fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05")),
			"----------------------------------------",
		}

		for _, line := range headerLines {
			content.WriteString(cleanText(line) + "\n")
		}

		// แสดงรายการอาหาร
		content.WriteString("[ รายการอาหาร ]\n")
		content.WriteString("----------------------------------------\n")

		for i, order := range job.Receipt.Orders {
			for _, item := range order.Items {
				// แสดงรายการอาหาร
				itemLine := fmt.Sprintf("%d. %s", i+1, item.MenuItem.Name)
				if item.Quantity > 1 {
					itemLine += fmt.Sprintf(" x%d", item.Quantity)
				}
				itemLine += fmt.Sprintf("   ฿%.2f", item.Price)
				content.WriteString(cleanText(itemLine) + "\n")

				// ตัวเลือกเพิ่มเติม
				for _, opt := range item.Options {
					optionLine := fmt.Sprintf("   • %s: %s",
						cleanText(opt.MenuOption.OptionGroup.Name),
						cleanText(opt.MenuOption.Name))
					content.WriteString(optionLine + "\n")
				}

				// หมายเหตุพิเศษ
				if item.Notes != "" {
					content.WriteString(fmt.Sprintf("   [หมายเหตุ: %s]\n", cleanText(item.Notes)))
				}
			}
		}

		content.WriteString("----------------------------------------\n")

		// แสดงผลรวม
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

		// ข้อมูลการชำระเงิน
		paymentInfo := fmt.Sprintf("ชำระโดย: %s", cleanText(job.Receipt.PaymentMethod))
		content.WriteString(paymentInfo + "\n")

		// ข้อมูลพนักงาน
		employeeInfo := fmt.Sprintf("พนักงาน: %d", job.Receipt.StaffID)
		content.WriteString(employeeInfo + "\n")

		// ข้อความขอบคุณ
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

type PrintJobResponse struct {
	ID        uint            `json:"id"`
	PrinterID uint            `json:"printer_id"`
	OrderID   *uint           `json:"order_id,omitempty"`
	Content   []byte          `json:"content"`
	Status    string          `json:"status"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
	Order     *models.Order   `json:"order,omitempty"`
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
	// ตรวจสอบ params
	printerIP := c.Query("printer_ip")
	vendorID := c.Query("vendor_id")
	productID := c.Query("product_id")

	// ค้นหาเครื่องพิมพ์
	var printer models.Printer
	var err error

	if printerIP != "" {
		err = db.DB.Where("ip_address = ?", printerIP).First(&printer).Error
	} else {
		err = db.DB.Where("type = ? AND vendor_id = ? AND product_id = ?",
			"usb", vendorID, productID).First(&printer).Error
	}

	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Printer not found",
		})
	}

	// ดึงงานพิมพ์ด้วย printer_id
	var jobs []models.PrintJob
	// err = db.DB.Where("printer_id = ? AND status = ?", printer.ID, "pending").
	// 	Preload("Order").
	// 	Preload("Order.Items", "status = ?", "pending").
	// 	Preload("Order.Items.MenuItem").
	// 	Preload("Order.Items.Options").
	// 	Preload("Order.Items.Options.MenuOption").
	// 	Preload("Order.Items.Options.MenuOption.OptionGroup").
	// 	Order("created_at ASC").
	// 	Find(&jobs).Error

	err = db.DB.Where("printer_id = ? AND status = ?", printer.ID, "pending").
		Preload("Order").
		Preload("Order.Items", "status = ?", "pending").
		Preload("Order.Items.MenuItem").
		Preload("Order.Items.Options").
		Preload("Order.Items.Options.MenuOption").
		Preload("Order.Items.Options.MenuOption.OptionGroup").
		Preload("Receipt"). // ถ้าเป็นไปได้ควรแยกคิวรี่กัน N+1
		Preload("Receipt.Orders.Items.MenuItem").
		Preload("Receipt.Discounts.DiscountType").
		Preload("Receipt.Charges.ChargeType").
		Order("created_at ASC").
		Find(&jobs).Error
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch print jobs",
		})
	}

	// เตรียม response
	var responses []PrintJobResponse
	for _, job := range jobs {
		var preparedContent []byte
		var err error

		if job.OrderID != nil {
			// ใบแจ้งออเดอร์
			preparedContent, err = prepareOrderPrintContent(job)
			if err != nil {
				log.Printf("Failed to prepare order content for job %d: %v", job.ID, err)
				continue
			}
		} else if job.ReceiptID != nil {
			// ใบเสร็จ
			preparedContent, err = prepareReceiptPrintContent(job)
			if err != nil {
				log.Printf("Failed to prepare receipt content for job %d: %v", job.ID, err)
				continue
			}
		} else {
			// หากไม่มี order_id หรือ receipt_id ให้ข้าม
			log.Printf("Skipping job %d due to missing identifiers", job.ID)
			continue
		}

		// แปลงเป็น bitmap
		bitmapImage, err := convertToBitmap(preparedContent)
		if err != nil {
			log.Printf("Failed to convert job %d to bitmap: %v", job.ID, err)
			continue
		}

		// สร้าง response
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
				IPAddress:  printer.IPAddress,
				Port:       printer.Port,
				Department: printer.Department,
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
