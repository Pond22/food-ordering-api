package api_v2

import (
	"bufio"
	"bytes"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	service "food-ordering-api/services"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/golang/freetype/truetype"
	"golang.org/x/image/draw"
	"golang.org/x/image/font"
	"golang.org/x/image/math/fixed"
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

type PrintBillCheckDiscount struct {
	DiscountTypeID uint   `json:"discount_type_id" binding:"required"`
	Reason         string `json:"reason"`
}

type PrintBillCheckCharge struct {
	ChargeTypeID uint   `json:"charge_type_id" binding:"required"`
	Quantity     int    `json:"quantity" binding:"required,min=1"`
	Note         string `json:"note"`
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

// ฟังก์ชันสำหรับเตรียมเนื้อหาสำหรับพิมพ์ออเดอร์ไปครัว
func V2_prepareOrderPrintContent(job models.PrintJob) ([]byte, error) {
	var content bytes.Buffer

	if job.Order != nil {
		// ส่วนหัวของใบออเดอร์
		headerLines := []string{
			"",
			fmt.Sprintf("~=== ออเดอร์โต๊ะ %d ===", job.Order.TableID),
			fmt.Sprintf("~Order #%d", job.Order.ID),
			"",
			"====",
			fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05")),
			"====",
		}

		for _, line := range headerLines {
			content.WriteString(line + "\n")
		}

		// แยกรายการตามสถานะ
		var newItems []models.OrderItem
		for _, item := range job.Order.Items {
			if item.Status == "pending" {
				newItems = append(newItems, item)
			}
		}

		if len(newItems) > 0 {
			content.WriteString("\n~รายการใหม่\n")
			content.WriteString("----\n")

			for i, item := range newItems {
				// หมายเลขรายการและชื่ออาหาร
				itemName := fmt.Sprintf("%d. %s", i+1, item.MenuItem.Name)

				// แสดงจำนวนให้ชัดเจนขึ้น
				quantityText := ""
				if item.Quantity > 1 {
					quantityText = fmt.Sprintf("~>>> จำนวน %d ที่ <<<", item.Quantity)
				}

				// ใช้ฟังก์ชัน wrapItemName สำหรับตัดคำ
				itemLines := wrapItemName(itemName, 35)

				// พิมพ์ชื่อรายการ
				for _, line := range itemLines {
					content.WriteString(line + "\n")
				}

				// พิมพ์จำนวน (ถ้ามีมากกว่า 1)
				if quantityText != "" {
					content.WriteString(quantityText + "\n")
				}

				// ตัวเลือกเพิ่มเติม
				if len(item.Options) > 0 {
					content.WriteString("รายละเอียดเพิ่มเติม:\n")
					for _, opt := range item.Options {
						optName := "   • " + opt.MenuOption.OptionGroup.Name + ": " + opt.MenuOption.Name
						// ใช้ wrapItemName สำหรับตัวเลือกด้วย
						optLines := wrapItemName(optName, 35)
						for _, line := range optLines {
							content.WriteString(line + "\n")
						}
					}
				}

				// หมายเหตุพิเศษ
				if item.Notes != "" {
					content.WriteString("~[หมายเหตุ]\n")
					noteLine := "   " + item.Notes
					// ใช้ wrapItemName สำหรับหมายเหตุ
					noteLines := wrapItemName(noteLine, 35)
					for _, line := range noteLines {
						content.WriteString(line + "\n")
					}
				}

				// เว้นบรรทัดระหว่างรายการ
				if i < len(newItems)-1 {
					content.WriteString("\n")
					content.WriteString("====\n")
					content.WriteString("\n")
				}
			}
		}

		footerLines := []string{
			"",
			"====",
			"~โปรดตรวจสอบรายการให้ครบถ้วน",
			fmt.Sprintf("~จำนวนรายการทั้งหมด: %d รายการ", len(newItems)),
			"====",
		}

		for _, line := range footerLines {
			content.WriteString(line + "\n")
		}
	} else {
		content.Write(job.Content)
	}

	return content.Bytes(), nil
}

func PrepareReceiptPrintContent(job models.PrintJob) ([]byte, error) {
	formatter := service.NewPrintFormatter(job.Printer.PaperSize)
	var content bytes.Buffer

	// ส่วนหัวของใบเสร็จ
	headerLines := []string{
		"",
		"~Kaze",
		"",
		"~Grand Kaze Yakiniku and Sushi Bar",
		"~8/3 หมู่ที่ 5 ตำบลท่าศาลา อำเภอเมืองเชียงใหม่ จังหวัด",
		"~เชียงใหม่ 50000, เมืองเชียงใหม่, เชียงใหม่, 50000",
		"~โทรศัพท์: 0952215566",
		"~บริษัท คาเสะกรุ๊ป จำกัด",
		"~--------------------------------",
		"~เลขประจำตัวผู้เสียภาษีอากร: 0505565003291",
		"~ใบเสร็จรับเงิน / ใบกำกับภาษีอย่างย่อ",
		"~ราคาสินค้ายังไม่รวมภาษีมูลค่าเพิ่ม",
		formatter.GetDivider(),
	}

	for _, line := range headerLines {
		content.WriteString(line + "\n")
	}

	// ข้อมูลการขาย
	saleInfo := []string{
		fmt.Sprintf("เลขที่:                           %d", job.Receipt.ID),
		fmt.Sprintf("โต๊ะที่:                          %s", job.Receipt.TableID),
		fmt.Sprintf("พนักงาน:                      %s", job.Receipt.Staff.Name),
		fmt.Sprintf("วันที่:                         %s", time.Now().Format("02-01-2006")),
		fmt.Sprintf("เวลาเข้า:                         %s", job.Receipt.CreatedAt.Format("15:04")),
		fmt.Sprintf("เวลาออก:                         %s", time.Now().Format("15:04")),
		formatter.GetDivider(),
	}

	for _, line := range saleInfo {
		content.WriteString(line + "\n")
	}

	// หัวข้อตาราง
	content.WriteString(fmt.Sprintf("%-35s ~~%5s **%12s\n", "รายการ", "จำนวน", "ราคา"))
	content.WriteString(formatter.GetDivider() + "\n")

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

	// คำนวณยอดรวม
	var totalItems int
	for _, order := range job.Receipt.Orders {
		for _, item := range order.Items {
			if item.Status != "cancelled" {
				totalItems += item.Quantity
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

				itemTotal := item.Price * float64(item.Quantity)
				// เพิ่มราคาตัวเลือกเสริม
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
					itemGroups[key] = struct {
						MenuItem models.MenuItem
						Quantity int
						Price    float64
						Options  []models.OrderItemOption
						Notes    string
					}{
						MenuItem: item.MenuItem,
						Quantity: item.Quantity,
						Price:    itemTotal,
						Options:  item.Options,
						Notes:    item.Notes,
					}
				}
			}
		}
	}

	// แปลง map เป็น slice และเรียงตามชื่อเมนู
	var groupedItems []struct {
		MenuItem models.MenuItem
		Quantity int
		Price    float64
		Options  []models.OrderItemOption
		Notes    string
	}
	for _, group := range itemGroups {
		groupedItems = append(groupedItems, group)
	}

	sort.Slice(groupedItems, func(i, j int) bool {
		return groupedItems[i].MenuItem.Name < groupedItems[j].MenuItem.Name
	})

	// พิมพ์รายการ
	for _, group := range groupedItems {
		itemName := group.MenuItem.Name
		itemLines := wrapItemName(itemName, 35)

		// พิมพ์บรรทัดแรกพร้อมจำนวนและราคา
		content.WriteString(fmt.Sprintf("%-35s ~~%5d **%12.2f\n",
			itemLines[0],
			group.Quantity,
			group.Price))

		// ถ้ามีบรรทัดต่อไป ให้พิมพ์โดยไม่มีจำนวนและราคา
		for i := 1; i < len(itemLines); i++ {
			content.WriteString(fmt.Sprintf("%-35s ~~%5s **%12s\n",
				itemLines[i],
				"",
				""))
		}

		// ตัวเลือกเพิ่มเติม
		for _, opt := range group.Options {
			optName := "  • " + opt.MenuOption.Name
			optPrice := opt.Price * float64(opt.Quantity)

			// ตัดข้อความตัวเลือกที่ยาวเกิน
			optLines := wrapItemName(optName, 35)

			// พิมพ์บรรทัดแรกพร้อมจำนวนและราคา
			content.WriteString(fmt.Sprintf("%-35s ~~%5d **%12.2f\n",
				optLines[0],
				opt.Quantity,
				optPrice))

			// ถ้ามีบรรทัดต่อไป ให้พิมพ์โดยไม่มีจำนวนและราคา
			for i := 1; i < len(optLines); i++ {
				content.WriteString(fmt.Sprintf("%-35s ~~%5s **%12s\n",
					optLines[i],
					"",
					""))
			}
		}

		if group.Notes != "" {
			content.WriteString(fmt.Sprintf("   [หมายเหตุ: %s]\n", group.Notes))
		}
	}

	content.WriteString(formatter.GetDivider() + "\n")

	// สรุปยอด
	summaryLines := []string{
		fmt.Sprintf("%-35s ~~%5d **%12.2f", "ยอดรวม", totalItems, job.Receipt.SubTotal),
	}

	// แสดงส่วนลด
	if job.Receipt.DiscountTotal > 0 {
		for _, discount := range job.Receipt.Discounts {
			summaryLines = append(summaryLines,
				fmt.Sprintf("%-35s ~~%5s **%12.2f",
					fmt.Sprintf("ส่วนลด - %s", discount.DiscountType.Name),
					"",
					-discount.Value))
		}
	}

	// แสดงค่าใช้จ่ายเพิ่มเติม
	if job.Receipt.ChargeTotal > 0 {
		for _, charge := range job.Receipt.Charges {
			chargeAmount := charge.Amount * float64(charge.Quantity)
			summaryLines = append(summaryLines,
				fmt.Sprintf("%-35s ~~%5s **%12.2f",
					fmt.Sprintf("%s x%d", charge.ChargeType.Name, charge.Quantity),
					"",
					chargeAmount))
		}
	}

	// แสดง VAT
	summaryLines = append(summaryLines,
		fmt.Sprintf("%-35s ~~%5s **%12.2f", "ภาษีมูลค่าเพิ่ม 7%", "", job.Receipt.ServiceCharge))

	// ยอดสุทธิ
	summaryLines = append(summaryLines,
		formatter.GetDivider(),
		fmt.Sprintf("%-35s ~~%5s **฿%11.2f", "ยอดรวมสุทธิ", "", job.Receipt.Total),
		fmt.Sprintf("%-35s ~~%5s **฿%11.2f", job.Receipt.PaymentMethod, "", job.Receipt.Total),
		"",
		"~~ขอขอบพระคุณที่มาใช้บริการค่ะ",
	)

	for _, line := range summaryLines {
		content.WriteString(line + "\n")
	}

	return content.Bytes(), nil
}

func prepareCancelPrintContent(job models.PrintJob) ([]byte, error) {
	formatter := service.NewPrintFormatter(job.Printer.PaperSize)
	var content bytes.Buffer

	if job.Order != nil {
		headerLines := []string{
			formatter.CenterText("== ใบแจ้งยกเลิกรายการอาหาร =="),
			formatter.FormatLine(fmt.Sprintf("โต๊ะ: %d", job.Order.TableID)),
			formatter.GetDivider(),
			formatter.FormatLine(" รายการอาหาร                จำนวนที่ยกเลิก"),
			formatter.GetDivider(),
		}

		for _, line := range headerLines {
			content.WriteString(line + "\n")
		}

		// ดึงข้อมูลจากฟิลด์ Content
		items := strings.Split(string(job.Content), "\n")
		for _, item := range items {
			parts := strings.Split(item, "|")
			if len(parts) == 2 {
				content.WriteString(formatter.FormatTableRow([]string{parts[0], parts[1]}, []int{25, 5}) + "\n")
			}
		}

		footerLines := []string{
			formatter.GetDivider(),
			formatter.FormatLine(fmt.Sprintf("เวลายกเลิก: %s", time.Now().Format("02/01/2006 15:04:05"))),
			formatter.GetDoubleDivider(),
			formatter.CenterText("กรุณาตรวจสอบการยกเลิก"),
			formatter.GetDoubleDivider(),
		}

		for _, line := range footerLines {
			content.WriteString(line + "\n")
		}
	}

	return content.Bytes(), nil
}

func prepareQRCodePrintContent(job models.PrintJob, png []byte) ([]byte, error) {
	formatter := service.NewPrintFormatter(job.Printer.PaperSize)
	var content bytes.Buffer

	headerLines := []string{
		formatter.CenterText("===== QR Code สำหรับโต๊ะ ====="),
		formatter.CenterText("1"),
		formatter.FormatLine(fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05"))),
		formatter.GetDoubleDivider(),
	}

	for _, line := range headerLines {
		content.WriteString(line + "\n")
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

	footerLines := []string{
		"\n" + formatter.GetDoubleDivider(),
		formatter.CenterText("โปรดสแกน QR Code เพื่อเข้าสู่ระบบ"),
		formatter.GetDoubleDivider(),
	}

	for _, line := range footerLines {
		content.WriteString(line + "\n")
	}

	finalContent = append(finalContent, content.Bytes()...)

	return finalContent, nil
}

// ฟังก์ชันแปลงเนื้อหาเป็น bitmap
func convertToBitmap(content []byte, paperSize string) ([]byte, error) {
	if isPNG(content) {
		return convertPNGToBitmap(content, paperSize)
	}

	// ตรวจสอบว่าเป็นเนื้อหาใบเสร็จหรือใบรายการอาหารหรือไม่
	if bytes.Contains(content, []byte("GRAND KAZE")) ||
		bytes.Contains(content, []byte("ใบเสร็จรับเงิน")) ||
		bytes.Contains(content, []byte("ใบรายการอาหาร")) {
		return convertReceiptToBitmap(content, paperSize)
	}

	// ใช้ template ตามขนาดกระดาษ
	template := service.Templates[paperSize]
	if template.Width == 0 {
		template = service.Templates["80"] // ใช้ค่าเริ่มต้นถ้าไม่พบ template
	}

	// โหลดฟอนต์
	fontBytes, err := os.ReadFile("THSarabunNew.ttf")
	if err != nil {
		return nil, fmt.Errorf("error loading font: %v", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing font: %v", err)
	}

	// ปรับขนาดฟอนต์ตามขนาดกระดาษ
	face := truetype.NewFace(f, &truetype.Options{
		Size:    template.FontSize,
		DPI:     203.0,
		Hinting: font.HintingFull,
	})
	defer face.Close()

	d := &font.Drawer{
		Dst:  nil,
		Src:  image.Black,
		Face: face,
	}

	// เตรียม drawer สำหรับวัดความกว้าง
	tempImg := image.NewRGBA(image.Rect(0, 0, template.Width, 1000))
	d.Dst = tempImg

	// แยกข้อความเป็นบรรทัด
	var allLines []string
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		text := scanner.Text()

		// ตรวจสอบรูปแบบเส้นคั่น
		if strings.Contains(text, "----") || strings.Contains(text, "====") {
			// สร้างเส้นคั่นใหม่ที่เต็มความกว้าง
			var dividerChar string
			if strings.Contains(text, "====") {
				dividerChar = "="
			} else {
				dividerChar = "-"
			}
			// สร้างเส้นคั่นที่เต็มความกว้างโดยเว้นขอบ 2 ตัวอักษร
			fullWidthDivider := strings.Repeat(dividerChar, template.Width/8-4)
			text = "  " + fullWidthDivider + "  "
		}

		// ใช้ formatter เพื่อจัดรูปแบบข้อความให้พอดีกับความกว้าง
		formatter := service.NewPrintFormatter(paperSize)
		wrappedLines := formatter.WrapText(text)
		allLines = append(allLines, wrappedLines...)
	}

	// คำนวณความสูงของภาพ
	totalLines := len(allLines)
	height := int(float64(totalLines)*template.FontSize*template.LineSpacing) + 100

	// สร้างภาพใหม่
	img := image.NewRGBA(image.Rect(0, 0, template.Width, height))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	d.Dst = img

	// วาดข้อความ
	y := int(template.FontSize * 3.0)
	for _, text := range allLines {
		// ตรวจสอบประเภทของข้อความและกำหนดการจัดวาง
		var x int
		var processedText string = text

		// กรณีมี ~~ หรือ ** อยู่ในข้อความ (ไม่ใช่นำหน้า)
		if !strings.HasPrefix(text, "~") && !strings.HasPrefix(text, "*") {
			// แยกส่วนที่มี ~~ ก่อน
			if strings.Contains(text, "~~") {
				parts := strings.Split(text, "~~")
				if len(parts) == 2 {
					leftPart := parts[0]
					rightPart := parts[1]

					// ตรวจสอบว่าส่วนที่สองมี ** หรือไม่
					if strings.Contains(rightPart, "**") {
						subParts := strings.Split(rightPart, "**")
						if len(subParts) == 2 {
							// วาดส่วนแรกชิดซ้าย
							d.Dot = fixed.Point26_6{
								X: fixed.I(template.LeftPadding),
								Y: fixed.I(y),
							}
							d.DrawString(leftPart)

							// วาดส่วนกลาง
							width := d.MeasureString(subParts[0]).Ceil()
							x = (template.Width - width) / 2
							if x < template.LeftPadding {
								x = template.LeftPadding
							}
							d.Dot = fixed.Point26_6{
								X: fixed.I(x),
								Y: fixed.I(y),
							}
							d.DrawString(subParts[0])

							// วาดส่วนขวา
							width = d.MeasureString(subParts[1]).Ceil()
							x = template.Width - width - template.LeftPadding
							if x < template.LeftPadding {
								x = template.LeftPadding
							}
							d.Dot = fixed.Point26_6{
								X: fixed.I(x),
								Y: fixed.I(y),
							}
							d.DrawString(subParts[1])
							y += int(template.FontSize * template.LineSpacing)
							continue
						}
					}
				}
			}
		}

		// กรณีปกติ (เครื่องหมายนำหน้าหรือไม่มีเครื่องหมาย)
		switch {
		// จัดกึ่งกลาง
		case strings.HasPrefix(text, "~"):
			processedText = strings.TrimPrefix(text, "~")
			width := d.MeasureString(processedText).Ceil()
			x = (template.Width - width) / 2

		// จัดชิดขวา
		case strings.HasPrefix(text, "*"):
			processedText = strings.TrimPrefix(text, "*")
			width := d.MeasureString(processedText).Ceil()
			x = template.Width - width - template.LeftPadding

		// กรณีอื่นๆ จัดชิดซ้าย
		default:
			processedText = text
			x = template.LeftPadding
		}

		// ป้องกันไม่ให้ x น้อยกว่า padding ขั้นต่ำ
		if x < template.LeftPadding {
			x = template.LeftPadding
		}

		d.Dot = fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		}
		d.DrawString(processedText)
		y += int(template.FontSize * template.LineSpacing)
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
	widthBytes := (template.Width + 7) / 8
	buf.WriteByte(byte(widthBytes & 0xFF))
	buf.WriteByte(byte(widthBytes >> 8))
	buf.WriteByte(byte(height & 0xFF))
	buf.WriteByte(byte(height >> 8))

	// แปลงเป็น bitmap
	for y := 0; y < height; y++ {
		for x := 0; x < template.Width; x += 8 {
			var b byte
			for bit := 0; bit < 8; bit++ {
				if x+bit < template.Width {
					r, g, b_, _ := img.At(x+bit, y).RGBA()
					brightness := (r*299 + g*587 + b_*114) / 1000
					if brightness < 0xAFFF { // ลดค่า threshold ให้เท่ากับ convertToBitmap
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

// เพิ่มฟังก์ชันสำหรับโหลดและปรับขนาดโลโก้
func loadAndResizeLogo(template service.PrinterTemplate) (image.Image, error) {
	// ลองโหลดไฟล์โลโก้จากหลายนามสกุล
	possiblePaths := []string{
		"./logo.jpg",
	}

	var logoFile *os.File
	var err error
	var foundPath string

	// ลองเปิดไฟล์จากเส้นทางที่เป็นไปได้
	for _, path := range possiblePaths {
		if logoFile, err = os.Open(path); err == nil {
			foundPath = path
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("error loading logo from any path: %v", err)
	}
	defer logoFile.Close()

	// อ่านรูปภาพตามนามสกุลไฟล์
	var logo image.Image
	if strings.HasSuffix(foundPath, ".png") {
		logo, err = png.Decode(logoFile)
	} else if strings.HasSuffix(foundPath, ".jpg") || strings.HasSuffix(foundPath, ".jpeg") {
		logo, err = jpeg.Decode(logoFile)
	} else {
		return nil, fmt.Errorf("unsupported image format")
	}

	if err != nil {
		return nil, fmt.Errorf("error decoding logo: %v", err)
	}

	// คำนวณขนาดใหม่ของโลโก้ (ให้กว้างประมาณ 80% ของความกว้างกระดาษ)
	targetWidth := int(float64(template.Width) * 0.8)
	ratio := float64(logo.Bounds().Dy()) / float64(logo.Bounds().Dx())
	targetHeight := int(float64(targetWidth) * ratio)

	// สร้างภาพใหม่ตามขนาดที่ต้องการ
	resized := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	draw.CatmullRom.Scale(resized, resized.Bounds(), logo, logo.Bounds(), draw.Over, nil)

	return resized, nil
}

func convertReceiptToBitmap(content []byte, paperSize string) ([]byte, error) {
	// ใช้ template ตามขนาดกระดาษ
	template := service.Templates[paperSize]
	if template.Width == 0 {
		template = service.Templates["80"] // ใช้ค่าเริ่มต้นถ้าไม่พบ template
	}

	// โหลดฟอนต์
	fontBytes, err := os.ReadFile("THSarabunNew.ttf")
	if err != nil {
		return nil, fmt.Errorf("error loading font: %v", err)
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return nil, fmt.Errorf("error parsing font: %v", err)
	}

	// ปรับขนาดฟอนต์ตามขนาดกระดาษ
	face := truetype.NewFace(f, &truetype.Options{
		Size:    16,
		DPI:     204.0, // ถ้าใช้ 300 dot กับขนาดตัว 14 มันต้องใช้ 9.47
		Hinting: font.HintingFull,
	})
	defer face.Close()

	d := &font.Drawer{
		Dst:  nil,
		Src:  image.Black,
		Face: face,
	}

	// เตรียม drawer สำหรับวัดความกว้าง
	tempImg := image.NewRGBA(image.Rect(0, 0, template.Width, 1000))
	d.Dst = tempImg

	// แยกข้อความเป็นบรรทัด
	var allLines []string
	scanner := bufio.NewScanner(bytes.NewReader(content))
	for scanner.Scan() {
		text := scanner.Text()

		// ตรวจสอบรูปแบบเส้นคั่น
		if strings.Contains(text, "----") || strings.Contains(text, "====") {
			// สร้างเส้นคั่นใหม่ที่เต็มความกว้าง
			var dividerChar string
			if strings.Contains(text, "====") {
				dividerChar = "="
			} else {
				dividerChar = "-"
			}
			// สร้างเส้นคั่นที่เต็มความกว้างโดยเว้นขอบ 2 ตัวอักษร
			fullWidthDivider := strings.Repeat(dividerChar, template.Width/8-4)
			text = "  " + fullWidthDivider + "  "
		}

		// ใช้ formatter เพื่อจัดรูปแบบข้อความให้พอดีกับความกว้าง
		formatter := service.NewPrintFormatter(paperSize)
		wrappedLines := formatter.WrapText(text)
		allLines = append(allLines, wrappedLines...)
	}

	// คำนวณความสูงของภาพ
	totalLines := len(allLines)
	var logoHeight int

	// โหลดโลโก้เพื่อคำนวณความสูงรวม (แต่ยังไม่วาด)
	if logo, err := loadAndResizeLogo(template); err == nil {
		logoHeight = logo.Bounds().Dy() + int(template.FontSize*2.0) + 20 // เพิ่ม padding ด้านล่างโลโก้
	}

	// คำนวณความสูงรวมทั้งหมด
	height := logoHeight + int(float64(totalLines)*template.FontSize*template.LineSpacing) + 100

	// สร้างภาพใหม่
	img := image.NewRGBA(image.Rect(0, 0, template.Width, height))
	draw.Draw(img, img.Bounds(), image.White, image.Point{}, draw.Src)
	d.Dst = img

	// กำหนดตำแหน่ง y เริ่มต้น
	var y int

	// โหลดและวาดโลโก้ (ถ้ามี)
	if logo, err := loadAndResizeLogo(template); err == nil {
		// คำนวณตำแหน่ง x เพื่อให้โลโก้อยู่กึ่งกลาง
		x := (template.Width - logo.Bounds().Dx()) / 2
		// วาดโลโก้ที่ตำแหน่งบนสุด
		draw.Draw(img, image.Rect(x, 10, x+logo.Bounds().Dx(), 10+logo.Bounds().Dy()),
			logo, image.Point{}, draw.Over)
		// ปรับ y เริ่มต้นของข้อความให้อยู่ใต้โลโก้
		y = 10 + logo.Bounds().Dy() + int(template.FontSize*2.0)
	} else {
		// ถ้าไม่มีโลโก้ ใช้ค่า y เดิม
		y = int(template.FontSize * 3.0)
	}

	// วาดข้อความ
	for _, text := range allLines {
		// ตรวจสอบประเภทของข้อความและกำหนดการจัดวาง
		var x int
		var processedText string = text

		// กรณีมี ~~ หรือ ** อยู่ในข้อความ (ไม่ใช่นำหน้า)
		if !strings.HasPrefix(text, "~") && !strings.HasPrefix(text, "*") {
			// แยกส่วนที่มี ~~ ก่อน
			if strings.Contains(text, "~~") {
				parts := strings.Split(text, "~~")
				if len(parts) == 2 {
					leftPart := parts[0]
					rightPart := parts[1]

					// ตรวจสอบว่าส่วนที่สองมี ** หรือไม่
					if strings.Contains(rightPart, "**") {
						subParts := strings.Split(rightPart, "**")
						if len(subParts) == 2 {
							// วาดส่วนแรกชิดซ้าย
							d.Dot = fixed.Point26_6{
								X: fixed.I(template.LeftPadding),
								Y: fixed.I(y),
							}
							d.DrawString(leftPart)

							// วาดส่วนกลาง
							width := d.MeasureString(subParts[0]).Ceil()
							x = (template.Width - width) / 2
							if x < template.LeftPadding {
								x = template.LeftPadding
							}
							d.Dot = fixed.Point26_6{
								X: fixed.I(x),
								Y: fixed.I(y),
							}
							d.DrawString(subParts[0])

							// วาดส่วนขวา
							width = d.MeasureString(subParts[1]).Ceil()
							x = template.Width - width - template.LeftPadding
							if x < template.LeftPadding {
								x = template.LeftPadding
							}
							d.Dot = fixed.Point26_6{
								X: fixed.I(x),
								Y: fixed.I(y),
							}
							d.DrawString(subParts[1])
							y += int(template.FontSize * template.LineSpacing)
							continue
						}
					}
				}
			}
		}

		// กรณีปกติ (เครื่องหมายนำหน้าหรือไม่มีเครื่องหมาย)
		switch {
		// จัดกึ่งกลาง
		case strings.HasPrefix(text, "~"):
			processedText = strings.TrimPrefix(text, "~")
			width := d.MeasureString(processedText).Ceil()
			x = (template.Width - width) / 2

		// จัดชิดขวา
		case strings.HasPrefix(text, "*"):
			processedText = strings.TrimPrefix(text, "*")
			width := d.MeasureString(processedText).Ceil()
			x = template.Width - width - template.LeftPadding

		// กรณีอื่นๆ จัดชิดซ้าย
		default:
			processedText = text
			x = template.LeftPadding
		}

		// ป้องกันไม่ให้ x น้อยกว่า padding ขั้นต่ำ
		if x < template.LeftPadding {
			x = template.LeftPadding
		}

		d.Dot = fixed.Point26_6{
			X: fixed.I(x),
			Y: fixed.I(y),
		}
		d.DrawString(processedText)
		y += int(template.FontSize * template.LineSpacing)
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
	widthBytes := (template.Width + 7) / 8
	buf.WriteByte(byte(widthBytes & 0xFF))
	buf.WriteByte(byte(widthBytes >> 8))
	buf.WriteByte(byte(height & 0xFF))
	buf.WriteByte(byte(height >> 8))

	// แปลงเป็น bitmap
	for y := 0; y < height; y++ {
		for x := 0; x < template.Width; x += 8 {
			var b byte
			for bit := 0; bit < 8; bit++ {
				if x+bit < template.Width {
					r, g, b_, _ := img.At(x+bit, y).RGBA()
					brightness := (r*299 + g*587 + b_*114) / 1000
					if brightness < 0xAFFF { // ลดค่า threshold ให้เท่ากับ convertToBitmap
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

func convertPNGToBitmap(pngData []byte, paperSize string) ([]byte, error) {
	template := service.Templates[paperSize]
	if template.Width == 0 {
		template = service.Templates["80"] // ใช้ค่าเริ่มต้นถ้าไม่พบ template
	}

	// อ่านรูป PNG
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %v", err)
	}

	// คำนวณความสูงให้ได้สัดส่วนที่ถูกต้อง
	height := img.Bounds().Dy() * template.Width / img.Bounds().Dx()

	// สร้างภาพ bitmap ใหม่
	bitmapImg := image.NewRGBA(image.Rect(0, 0, template.Width, height))

	// ใช้ nearest neighbor scaling สำหรับ QR Code
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
	widthBytes := (template.Width + 7) / 8
	buf.WriteByte(byte(widthBytes & 0xFF))
	buf.WriteByte(byte(widthBytes >> 8))
	buf.WriteByte(byte(height & 0xFF))
	buf.WriteByte(byte(height >> 8))

	// แปลงเป็น bitmap ด้วย threshold ที่เหมาะสมสำหรับ QR Code
	for y := 0; y < height; y++ {
		for x := 0; x < template.Width; x += 8 {
			var b byte
			for bit := 0; bit < 8; bit++ {
				if x+bit < template.Width {
					r, g, b_, _ := bitmapImg.At(x+bit, y).RGBA()
					brightness := (r*299 + g*587 + b_*114) / 1000
					if brightness < 0xAFFF { // ลดค่า threshold ให้เท่ากับ convertToBitmap
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

// isPNG ตรวจสอบว่าข้อมูลเป็นไฟล์ PNG หรือไม่
func isPNG(data []byte) bool {
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	return len(data) > 8 && bytes.Equal(data[:8], pngSignature)
}

// เพิ่มฟังก์ชันใหม่สำหรับตัดข้อความที่ยาวเกิน
func wrapItemName(name string, maxWidth int) []string {
	var lines []string

	// โหลดฟอนต์
	fontBytes, err := os.ReadFile("THSarabunNew.ttf")
	if err != nil {
		return []string{name}
	}

	f, err := truetype.Parse(fontBytes)
	if err != nil {
		return []string{name}
	}

	// สร้าง font face
	face := truetype.NewFace(f, &truetype.Options{
		Size:    15,
		DPI:     204.0,
		Hinting: font.HintingFull,
	})
	defer face.Close()

	d := &font.Drawer{
		Src:  image.Black,
		Face: face,
	}

	runes := []rune(name)
	currentLine := ""
	currentWidth := 0

	for i := 0; i < len(runes); i++ {
		r := runes[i]
		nextChar := string(r)
		charWidth := d.MeasureString(nextChar).Ceil()

		// ถ้าเพิ่มตัวอักษรแล้วเกินความกว้างที่กำหนด
		if currentWidth+charWidth > maxWidth*6 {
			if len(currentLine) > 0 {
				// หาตำแหน่งช่องว่างล่าสุด
				lastSpaceIndex := strings.LastIndex(currentLine, " ")
				if lastSpaceIndex > 0 {
					// ตัดที่ช่องว่าง
					lines = append(lines, strings.TrimSpace(currentLine[:lastSpaceIndex]))
					currentLine = currentLine[lastSpaceIndex+1:] + nextChar
					currentWidth = d.MeasureString(currentLine).Ceil()
				} else {
					// ถ้าไม่มีช่องว่าง ตัดตามความยาวปัจจุบัน
					lines = append(lines, strings.TrimSpace(currentLine))
					currentLine = nextChar
					currentWidth = charWidth
				}
			}
		} else {
			currentLine += nextChar
			currentWidth += charWidth
		}
	}

	// เพิ่มข้อความที่เหลือ
	if strings.TrimSpace(currentLine) != "" {
		lines = append(lines, strings.TrimSpace(currentLine))
	}

	return lines
}

func V2_prepareBillCheckPrintContent(orders []models.Order, tableIDs []string, discounts []PrintBillCheckDiscount, extraCharges []PrintBillCheckCharge, serviceChargePercent float64, subTotal, totalDiscount, totalExtraCharge, vatAmount, netTotal float64) ([]byte, error) {
	formatter := service.NewPrintFormatter("80")
	var content bytes.Buffer

	// ส่วนหัวของใบรายการอาหาร
	headerLines := []string{
		"",
		"~***** ใบรายการอาหาร *****",
		"",
		fmt.Sprintf("~โต๊ะ: %s", strings.Join(tableIDs, ", ")),
		fmt.Sprintf("~วันที่: %s", time.Now().Format("02-01-2006")),
		fmt.Sprintf("~เวลา: %s", time.Now().Format("15:04")),
		formatter.GetDivider(),
	}

	for _, line := range headerLines {
		content.WriteString(line + "\n")
	}

	// หัวข้อตาราง
	content.WriteString(fmt.Sprintf("%-35s ~~%5s **%12s\n", "รายการ", "จำนวน", "ราคา"))
	content.WriteString(formatter.GetDivider() + "\n")

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

	// คำนวณยอดรวม
	var totalItems int
	for _, order := range orders {
		for _, item := range order.Items {
			if item.Status != "cancelled" {
				totalItems += item.Quantity
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

				itemTotal := item.Price * float64(item.Quantity)
				// เพิ่มราคาตัวเลือกเสริม
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
					itemGroups[key] = struct {
						MenuItem models.MenuItem
						Quantity int
						Price    float64
						Options  []models.OrderItemOption
						Notes    string
					}{
						MenuItem: item.MenuItem,
						Quantity: item.Quantity,
						Price:    itemTotal,
						Options:  item.Options,
						Notes:    item.Notes,
					}
				}
			}
		}
	}

	// แปลง map เป็น slice และเรียงตามชื่อเมนู
	var groupedItems []struct {
		MenuItem models.MenuItem
		Quantity int
		Price    float64
		Options  []models.OrderItemOption
		Notes    string
	}
	for _, group := range itemGroups {
		groupedItems = append(groupedItems, group)
	}

	sort.Slice(groupedItems, func(i, j int) bool {
		return groupedItems[i].MenuItem.Name < groupedItems[j].MenuItem.Name
	})

	// พิมพ์รายการ
	for _, group := range groupedItems {
		itemName := group.MenuItem.Name
		itemLines := wrapItemName(itemName, 35)

		// พิมพ์บรรทัดแรกพร้อมจำนวนและราคา
		content.WriteString(fmt.Sprintf("%-35s ~~%5d **%12.2f\n",
			itemLines[0],
			group.Quantity,
			group.Price))

		// ถ้ามีบรรทัดต่อไป ให้พิมพ์โดยไม่มีจำนวนและราคา
		for i := 1; i < len(itemLines); i++ {
			content.WriteString(fmt.Sprintf("%-35s ~~%5s **%12s\n",
				itemLines[i],
				"",
				""))
		}

		// ตัวเลือกเพิ่มเติม
		for _, opt := range group.Options {
			optName := "  • " + opt.MenuOption.Name
			optPrice := opt.Price * float64(opt.Quantity)

			// ตัดข้อความตัวเลือกที่ยาวเกิน
			optLines := wrapItemName(optName, 35)

			// พิมพ์บรรทัดแรกพร้อมจำนวนและราคา
			content.WriteString(fmt.Sprintf("%-35s ~~%5d **%12.2f\n",
				optLines[0],
				opt.Quantity,
				optPrice))

			// ถ้ามีบรรทัดต่อไป ให้พิมพ์โดยไม่มีจำนวนและราคา
			for i := 1; i < len(optLines); i++ {
				content.WriteString(fmt.Sprintf("%-35s ~~%5s **%12s\n",
					optLines[i],
					"",
					""))
			}
		}

		if group.Notes != "" {
			content.WriteString(fmt.Sprintf("   [หมายเหตุ: %s]\n", group.Notes))
		}
	}

	content.WriteString(formatter.GetDivider() + "\n")

	// สรุปยอด
	summaryLines := []string{
		fmt.Sprintf("%-35s ~~%5d **%12.2f", "ยอดรวม", totalItems, subTotal),
	}

	// แสดงส่วนลด
	if totalDiscount > 0 {
		for _, discount := range discounts {
			var discountType models.DiscountType
			if err := db.DB.First(&discountType, discount.DiscountTypeID).Error; err != nil {
				continue
			}
			summaryLines = append(summaryLines,
				fmt.Sprintf("%-35s ~~%5s **%12.2f", discountType.Name, "", -totalDiscount))
		}
	}

	// แสดงค่าใช้จ่ายเพิ่มเติม
	if totalExtraCharge > 0 {
		for _, charge := range extraCharges {
			var chargeType models.AdditionalChargeType
			if err := db.DB.First(&chargeType, charge.ChargeTypeID).Error; err != nil {
				continue
			}
			summaryLines = append(summaryLines,
				fmt.Sprintf("%-35s ~~%5s **%12.2f", chargeType.Name, "", totalExtraCharge))
		}
	}

	// แสดง VAT
	summaryLines = append(summaryLines,
		fmt.Sprintf("%-35s ~~%5s **%12.2f", "ภาษีมูลค่าเพิ่ม 7%", "", vatAmount))

	// ยอดสุทธิ
	summaryLines = append(summaryLines,
		formatter.GetDivider(),
		fmt.Sprintf("%-35s ~~%5s **฿%11.2f", "ยอดรวมสุทธิ", "", netTotal),
		"",
		"~~ขอขอบพระคุณที่มาใช้บริการค่ะ",
	)

	for _, line := range summaryLines {
		content.WriteString(line + "\n")
	}

	return content.Bytes(), nil
}
