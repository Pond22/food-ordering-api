package api_v2

import (
	"bytes"
	"errors"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"math"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// @Summary รับงานพิมพ์ที่รอดำเนินการ
// @Description ดึงรายการงานพิมพ์ที่ยังไม่ได้พิมพ์สำหรับเครื่องพิมพ์ที่ระบุ (รองรับทั้ง IP และ USB)
// @Produce json
// @Param printer_ip query string false "Printer IP Address (สำหรับเครื่องพิมพ์เครือข่าย)"
// @Param vendor_id query string false "Vendor ID (สำหรับเครื่องพิมพ์ USB)"
// @Param product_id query string false "Product ID (สำหรับเครื่องพิมพ์ USB)"
// @Success 200 {array} models.PrintJob
// @Router /api/v2/printers/pending-jobs [get]
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
					"error": "รูปแบบตัวระบุเครื่องพิมพ์ USB ไม่ถูกต้อง",
				})
			}
		} else {
			err = db.DB.Where("type = ? AND ip_address = ?",
				"network", printerIP).First(&printer).Error
		}
	} else {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "กรุณาระบุพารามิเตอร์สำหรับระบุเครื่องพิมพ์",
		})
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": "ไม่พบเครื่องพิมพ์",
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการค้นหาเครื่องพิมพ์",
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
			"error": "เกิดข้อผิดพลาดในการดึงข้อมูลงานพิมพ์",
		})
	}

	// กำหนดจำนวน workers
	numWorkers := 10
	jobsChan := make(chan models.PrintJob, len(jobs))
	resultsChan := make(chan models.PrintJob, len(jobs))
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
					preparedContent, err = prepareOrderPrintContent(job)
				case "receipt":
					preparedContent, err = prepareReceiptPrintContent(job)
				case "cancelation":
					preparedContent, err = prepareCancelPrintContent(job)
				case "qr_code":
					preparedContent = job.Content
				default:
					preparedContent = job.Content
				}

				if err != nil {
					continue
				}

				// แปลงเนื้อหาเป็น bitmap ตามขนาดกระดาษ
				bitmapContent, err := convertToBitmap(preparedContent, printer.PaperSize)
				if err != nil {
					continue
				}

				job.Content = bitmapContent
				resultsChan <- job
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
	var results []models.PrintJob
	for job := range resultsChan {
		results = append(results, job)
	}

	return c.JSON(results)
}

// @Summary รีปริ้นเอกสาร
// @Description สั่งพิมพ์เอกสารซ้ำ (รองรับ order, receipt, qr_code)
// @Accept json
// @Produce json
// @Param id path int true "Print Job ID"
// @Success 200 {object} models.PrintJobฤ
// @Router /api/v2/printers/reprint/{id} [post]
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
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบงานพิมพ์ที่ระบุ",
		})
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการเตรียมเนื้อหาสำหรับพิมพ์",
		})
	}

	// แปลงเนื้อหาเป็น bitmap ตามขนาดกระดาษ
	bitmapContent, err := convertToBitmap(newJob.Content, originalJob.Printer.PaperSize)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการแปลงเนื้อหาเป็น bitmap",
		})
	}

	newJob.Content = bitmapContent

	if err := db.DB.Create(&newJob).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการสร้างงานพิมพ์ใหม่",
		})
	}

	return c.JSON(newJob)
}

// @Summary ดึงรายการงานพิมพ์ที่สามารถรีปริ้นได้
// @Description ดึงรายการงานพิมพ์ที่สามารถรีปริ้นได้ โดยสามารถกรองตามประเภทและช่วงเวลาได้
// @Produce json
// @Param type query string false "ประเภทงานพิมพ์ (order, receipt, cancelation, qr_code)"
// @Param start_date query string false "วันที่เริ่มต้น (YYYY-MM-DD)"
// @Param end_date query string false "วันที่สิ้นสุด (YYYY-MM-DD)"
// @Success 200 {array} models.PrintJob
// @Router /api/v2/printers/reprintable-jobs [get]
// @Tags Printer
func GetReprintableJobs(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	limit := c.QueryInt("limit", 10)
	offset := (page - 1) * limit

	jobType := c.Query("type")
	startDate := c.Query("start_date")
	endDate := c.Query("end_date")

	query := db.DB.Model(&models.PrintJob{}).
		Preload("Printer").
		Preload("Order.Items", func(db *gorm.DB) *gorm.DB {
			return db.Order("id ASC")
		}).
		Preload("Order.Items.MenuItem").
		Preload("Order.Items.MenuItem.Category").
		Preload("Order.Items.Options.MenuOption").
		Preload("Order.Items.Options.MenuOption.OptionGroup").
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

	if jobType != "" {
		if jobType == "others" {
			standardTypes := []string{"order", "receipt", "cancelation", "qr_code", "shift_report"}
			query = query.Where("job_type NOT IN (?)", standardTypes).
				Where("order_id IS NULL").
				Where("receipt_id IS NULL").
				Where("job_type NOT LIKE ?", "order%").
				Where("job_type NOT LIKE ?", "receipt%").
				Where("job_type NOT LIKE ?", "cancel%")
		} else {
			query = query.Where("job_type = ?", jobType)
		}
	}

	if startDate != "" {
		startTime, err := time.Parse("2006-01-02", startDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "รูปแบบวันที่เริ่มต้นไม่ถูกต้อง",
			})
		}
		query = query.Where("created_at >= ?", startTime)
	}

	if endDate != "" {
		endTime, err := time.Parse("2006-01-02", endDate)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "รูปแบบวันที่สิ้นสุดไม่ถูกต้อง",
			})
		}
		endTime = endTime.Add(24 * time.Hour)
		query = query.Where("created_at < ?", endTime)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการนับจำนวนรายการ",
		})
	}

	query = query.Where("status = ?", "completed").
		Order("created_at DESC").
		Offset(offset).
		Limit(limit)

	var jobs []models.PrintJob
	if err := query.Find(&jobs).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการดึงข้อมูลงานพิมพ์",
		})
	}

	// เตรียมเนื้อหาสำหรับแต่ละงานพิมพ์
	for i := range jobs {
		var content []byte
		var err error

		switch jobs[i].JobType {
		case "order":
			content, err = prepareOrderPrintContent(jobs[i])
		case "receipt":
			content, err = prepareReceiptPrintContent(jobs[i])
		case "cancelation":
			content, err = prepareCancelPrintContent(jobs[i])
		case "qr_code":
			content = jobs[i].Content
		default:
			content = jobs[i].Content
		}

		if err != nil {
			continue
		}

		// แปลงเนื้อหาเป็น bitmap ตามขนาดกระดาษ
		jobs[i].Content, _ = convertToBitmap(content, jobs[i].Printer.PaperSize)
	}

	return c.JSON(fiber.Map{
		"data": jobs,
		"pagination": fiber.Map{
			"current_page": page,
			"total_pages":  int(math.Ceil(float64(total) / float64(limit))),
			"total_items":  total,
			"per_page":     limit,
			"has_more":     (page * limit) < int(total),
		},
	})
}

// PrintBillCheckRequest สำหรับรับข้อมูลการพิมพ์ใบรายการอาหาร
type PrintBillCheckRequest struct {
	TableIDs []uint `json:"table_ids" binding:"required,min=1"`
}

// @Summary พิมพ์ใบรายการอาหารก่อนชำระเงิน
// @Description พิมพ์ใบรายการอาหารสำหรับตรวจทานก่อนชำระเงิน สามารถพิมพ์ได้ทั้งแบบโต๊ะเดี่ยวและรวมโต๊ะ
// @Accept json
// @Produce json
// @Param request body PrintBillCheckRequest true "ข้อมูลสำหรับพิมพ์ใบรายการอาหาร"
// @Success 200 {object} map[string]interface{}
// @Router /api/v2/printers/bill-check [post]
// @Tags Printer
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

	// ดึงข้อมูลออเดอร์ที่ยังไม่ได้ชำระเงินของทุกโต๊ะ (เฉพาะสถานะ pending และ completed)
	var orders []models.Order
	if err := db.DB.Where("table_id IN ? AND status IN (?, ?) AND receipt_id IS NULL",
		req.TableIDs, "pending", "completed").
		Preload("Items", "status IN (?)", []string{"pending", "completed"}).
		Preload("Items.MenuItem").
		Preload("Items.Options.MenuOption").
		Find(&orders).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถดึงข้อมูลออเดอร์ได้",
		})
	}

	if len(orders) == 0 {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบรายการอาหารที่ยังไม่ได้ชำระเงิน",
		})
	}

	// แปลง table IDs เป็น string slice สำหรับแสดงในใบพิมพ์
	tableNames := make([]string, len(tables))
	for i, table := range tables {
		tableNames[i] = table.Name
	}

	// ค้นหาเครื่องพิมพ์หลัก
	var printer models.Printer
	if err := db.DB.Where("name = ?", "main").First(&printer).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่พบเครื่องพิมพ์หลัก",
		})
	}

	// สร้างเนื้อหาสำหรับพิมพ์
	content, err := prepareBillCheckPrintContent(orders, tableNames)
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
		"message": "สร้างงานพิมพ์ใบรายการอาหารสำเร็จ",
		"job_id":  printJob.ID,
	})
}

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

func prepareBillCheckPrintContent(orders []models.Order, tableIDs []string) ([]byte, error) {
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

	itemGroups := make(map[OrderItemKey]struct {
		MenuItem models.MenuItem
		Quantity int
		Price    float64
		Options  []models.OrderItemOption
		Notes    string
	})

	var subTotal float64

	// รวมรายการที่เหมือนกัน
	for _, order := range orders {
		for _, item := range order.Items {
			// เปลี่ยนเงื่อนไขให้รวมทั้ง pending และ completed
			if item.Status == "pending" || item.Status == "completed" {
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

				subTotal += item.Price * float64(item.Quantity)
			}
		}
	}

	// แปลงเป็น slice และเรียงตามชื่อเมนู
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
				opt.Price)
			content.WriteString(optionLine + "\n")
		}

		if group.Notes != "" {
			content.WriteString(fmt.Sprintf("   [หมายเหตุ: %s]\n", cleanText(group.Notes)))
		}
	}

	content.WriteString("----------------------------------------\n")
	content.WriteString(fmt.Sprintf("ยอดรวม: ฿%.2f\n", subTotal))
	content.WriteString("----------------------------------------\n")
	content.WriteString("** กรุณาตรวจสอบรายการให้ครบถ้วน **\n")
	content.WriteString("========================================\n")

	return content.Bytes(), nil
}
