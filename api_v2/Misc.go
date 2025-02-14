package api_v2

import (
	"bufio"
	"bytes"
	"fmt"
	"food-ordering-api/models"
	service "food-ordering-api/services"
	"image"
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
func prepareOrderPrintContent(job models.PrintJob) ([]byte, error) {
	// สร้าง formatter ตามขนาดกระดาษ
	formatter := service.NewPrintFormatter(job.Printer.PaperSize)
	var content bytes.Buffer

	if job.Order != nil {
		headerLines := []string{
			formatter.CenterText("Order #" + fmt.Sprintf("%d", job.Order.ID)),
			formatter.FormatLine("โต๊ะ: " + fmt.Sprintf("%d", job.Order.TableID)),
			formatter.GetDivider(),
			formatter.FormatLine(fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05"))),
			formatter.GetDivider(),
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
			content.WriteString("\n" + formatter.FormatLine("[รายการใหม่]") + "\n")
			content.WriteString(formatter.GetDivider() + "\n")

			for i, item := range newItems {
				// หมายเลขรายการและชื่ออาหาร
				itemLine := fmt.Sprintf("%d. %s", i+1, item.MenuItem.Name)
				if item.Quantity > 1 {
					itemLine += fmt.Sprintf(" x%d", item.Quantity)
				}
				content.WriteString(formatter.FormatLine(itemLine) + "\n")

				// ตัวเลือกเพิ่มเติม
				for _, opt := range item.Options {
					optionLine := fmt.Sprintf("   • %s: %s",
						opt.MenuOption.OptionGroup.Name,
						opt.MenuOption.Name)
					content.WriteString(formatter.FormatLine(optionLine) + "\n")
				}

				// หมายเหตุพิเศษ
				if item.Notes != "" {
					content.WriteString(formatter.FormatLine(fmt.Sprintf("   [หมายเหตุ: %s]", item.Notes)) + "\n")
				}

				// เว้นบรรทัดระหว่างรายการ
				if i < len(newItems)-1 {
					content.WriteString(formatter.FormatLine("----------------") + "\n")
				}
			}
		}

		footerLines := []string{
			formatter.GetDoubleDivider(),
			formatter.CenterText("** กรุณาตรวจสอบรายการให้ครบถ้วน **"),
			formatter.GetDoubleDivider(),
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
	formatter := service.NewPrintFormatter(job.Printer.PaperSize)
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
			formatter.CenterText("***** ใบเสร็จรับเงิน *****"),
			formatter.FormatLine("Receipt #" + fmt.Sprintf("%d", job.Receipt.ID)),
			formatter.FormatLine("โต๊ะ: " + tableIDDisplay),
			formatter.GetDivider(),
			formatter.FormatLine(fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05"))),
			formatter.GetDivider(),
			formatter.FormatLine("[ รายการอาหาร ]"),
			formatter.GetDivider(),
		}

		for _, line := range headerLines {
			content.WriteString(line + "\n")
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

		// รวมรายการที่เหมือนกัน
		for _, order := range job.Receipt.Orders {
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
		for i, group := range groupedItems {
			itemLine := fmt.Sprintf("%d. %s", i+1, group.MenuItem.Name)
			if group.Quantity > 1 {
				itemLine += fmt.Sprintf(" x%d", group.Quantity)
			}
			content.WriteString(formatter.FormatPrice(itemLine, group.Price) + "\n")

			for _, opt := range group.Options {
				optionLine := fmt.Sprintf("   • %s", opt.MenuOption.Name)
				content.WriteString(formatter.FormatPrice(optionLine, opt.Price) + "\n")
			}

			if group.Notes != "" {
				content.WriteString(formatter.FormatLine(fmt.Sprintf("   [หมายเหตุ: %s]", group.Notes)) + "\n")
			}
		}

		content.WriteString(formatter.GetDivider() + "\n")

		// พิมพ์สรุปยอด
		summaryLines := []string{
			formatter.FormatPrice("ยอดรวม:", job.Receipt.SubTotal),
			formatter.FormatPrice("ส่วนลด:", job.Receipt.DiscountTotal),
			formatter.FormatPrice("ค่าใช้จ่ายเพิ่มเติม:", job.Receipt.ChargeTotal),
			formatter.FormatPrice("ค่าบริการ:", job.Receipt.ServiceCharge),
			formatter.GetDivider(),
			formatter.FormatPrice("ยอดสุทธิ:", job.Receipt.Total),
			formatter.GetDivider(),
		}

		for _, line := range summaryLines {
			content.WriteString(line + "\n")
		}

		content.WriteString(formatter.FormatLine(fmt.Sprintf("ชำระโดย: %s", job.Receipt.PaymentMethod)) + "\n")
		content.WriteString(formatter.FormatLine(fmt.Sprintf("พนักงาน: %d", job.Receipt.StaffID)) + "\n")

		footerLines := []string{
			formatter.GetDoubleDivider(),
			formatter.CenterText("ขอบคุณที่ใช้บริการ"),
			formatter.GetDoubleDivider(),
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
		d.Dot = fixed.Point26_6{
			X: fixed.I(template.LeftPadding),
			Y: fixed.I(y),
		}
		d.DrawString(text)
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
					if brightness < 0x7FFF {
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
					if brightness < 0x5FFF { // ใช้ threshold ต่ำกว่าสำหรับ QR Code
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

func segmentText(text string, maxWidth int, d *font.Drawer) []string {
	var segments []string
	currentSegment := ""

	// แยกข้อความออกเป็นคำๆ
	words := strings.Fields(text)
	if len(words) == 0 {
		return segments
	}

	currentSegment = words[0]
	for _, word := range words[1:] {
		// ทดสอบความยาวของประโยคปัจจุบัน
		testSegment := currentSegment + " " + word

		// วัดความกว้างของข้อความ
		testWidth := d.MeasureString(testSegment).Ceil()

		if testWidth > maxWidth {
			// ถ้าความยาวเกินที่กำหนด ให้บันทึกส่วนปัจจุบัน
			segments = append(segments, strings.TrimSpace(currentSegment))
			currentSegment = word
		} else {
			currentSegment = testSegment
		}
	}

	// เพิ่มส่วนสุดท้าย
	if currentSegment != "" {
		segments = append(segments, strings.TrimSpace(currentSegment))
	}

	return segments
}
