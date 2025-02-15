package service

import (
	"bytes"
	"fmt"
	"strings"
	"unicode/utf8"
)

// PrinterTemplate เก็บค่าการตั้งค่าสำหรับแต่ละขนาดเครื่องพิมพ์
type PrinterTemplate struct {
	Width          int     // ความกว้างเป็น pixels (384 หรือ 576)
	MaxChars       int     // จำนวนตัวอักษรสูงสุดต่อบรรทัด
	Divider        string  // เส้นขั้น (----)
	DoubleDivider  string  // เส้นขั้นคู่ (====)
	FontSize       float64 // ขนาดฟอนต์ที่เหมาะสม
	LineSpacing    float64 // ระยะห่างระหว่างบรรทัด
	LeftPadding    int     // ระยะห่างจากขอบซ้าย
	RightPadding   int     // ระยะห่างจากขอบขวา
	DefaultPadding int     // ระยะห่างมาตรฐาน
}

// Templates เก็บ template สำหรับแต่ละขนาดเครื่องพิมพ์
var Templates = map[string]PrinterTemplate{
	"58": {
		Width:          384,
		MaxChars:       32,
		Divider:        "--------------------------------", // เปลี่ยนเป็นเส้นประที่เต็มความกว้าง
		DoubleDivider:  "================================", // เปลี่ยนเป็นเส้นคู่ที่เต็มความกว้าง
		FontSize:       16.0,
		LineSpacing:    2.5,
		LeftPadding:    2,
		RightPadding:   2,
		DefaultPadding: 2,
	},
	"80": {
		Width:          576,
		MaxChars:       48,
		Divider:        "------------------------------------------------", // เปลี่ยนเป็นเส้นประที่เต็มความกว้าง
		DoubleDivider:  "================================================", // เปลี่ยนเป็นเส้นคู่ที่เต็มความกว้าง
		FontSize:       18.0,
		LineSpacing:    3.0,
		LeftPadding:    3,
		RightPadding:   3,
		DefaultPadding: 3,
	},
}

// PrintFormatter ใช้สำหรับจัดรูปแบบข้อความสำหรับการพิมพ์
type PrintFormatter struct {
	Template PrinterTemplate
}

// NewPrintFormatter สร้าง PrintFormatter ใหม่
func NewPrintFormatter(paperSize string) *PrintFormatter {
	template, exists := Templates[paperSize]
	if !exists {
		// ถ้าไม่พบ template ให้ใช้ 80 เป็นค่าเริ่มต้น
		template = Templates["80"]
	}
	return &PrintFormatter{Template: template}
}

// FormatLine จัดรูปแบบข้อความให้พอดีกับความกว้างที่กำหนด
func (pf *PrintFormatter) FormatLine(text string) string {
	if text == "" {
		return ""
	}

	textLength := utf8.RuneCountInString(text)
	if textLength > pf.Template.MaxChars {
		runes := []rune(text)
		if textLength >= 3 {
			maxLen := pf.Template.MaxChars - 3
			if maxLen < 0 {
				maxLen = 0
			}
			if maxLen > len(runes) {
				maxLen = len(runes)
			}
			return string(runes[:maxLen]) + "..."
		}
		maxLen := pf.Template.MaxChars
		if maxLen > len(runes) {
			maxLen = len(runes)
		}
		return string(runes[:maxLen])
	}
	return text
}

// FormatPrice จัดรูปแบบข้อความและราคาให้อยู่ซ้ายและขวาตามลำดับ
func (pf *PrintFormatter) FormatPrice(text string, price float64) string {
	if text == "" {
		return fmt.Sprintf("฿%.2f", price)
	}

	priceStr := fmt.Sprintf("฿%.2f", price)
	textLength := utf8.RuneCountInString(text)
	priceLength := utf8.RuneCountInString(priceStr)

	// คำนวณจำนวนช่องว่างที่ต้องการ
	spacesNeeded := pf.Template.MaxChars - textLength - priceLength
	if spacesNeeded < 1 {
		spacesNeeded = 1
	}

	return text + strings.Repeat(" ", spacesNeeded) + priceStr
}

// CenterText จัดข้อความให้อยู่กึ่งกลาง
func (pf *PrintFormatter) CenterText(text string) string {
	if text == "" {
		return ""
	}

	textLength := utf8.RuneCountInString(text)
	if textLength >= pf.Template.MaxChars {
		return pf.FormatLine(text)
	}

	padding := (pf.Template.MaxChars - textLength) / 2
	if padding < 0 {
		padding = 0
	}
	rightPadding := pf.Template.MaxChars - textLength - padding
	if rightPadding < 0 {
		rightPadding = 0
	}

	return strings.Repeat(" ", padding) + text + strings.Repeat(" ", rightPadding)
}

// GetDivider คืนค่าเส้นขั้นปกติ
func (pf *PrintFormatter) GetDivider() string {
	// สร้างเส้นคั่นที่มีความยาวเท่ากับ MaxChars
	divider := strings.Repeat("-", pf.Template.MaxChars)
	return pf.FormatLine(divider)
}

// GetDoubleDivider คืนค่าเส้นขั้นคู่
func (pf *PrintFormatter) GetDoubleDivider() string {
	// สร้างเส้นคั่นคู่ที่มีความยาวเท่ากับ MaxChars
	doubleDivider := strings.Repeat("=", pf.Template.MaxChars)
	return pf.FormatLine(doubleDivider)
}

// WrapText แบ่งข้อความเป็นหลายบรรทัดตามความกว้างที่กำหนด
func (pf *PrintFormatter) WrapText(text string) []string {
	var lines []string
	words := strings.Fields(text)
	if len(words) == 0 {
		return lines
	}

	currentLine := words[0]
	for _, word := range words[1:] {
		// ทดลองเพิ่มคำถัดไป
		testLine := currentLine + " " + word
		if utf8.RuneCountInString(testLine) <= pf.Template.MaxChars {
			currentLine = testLine
		} else {
			lines = append(lines, currentLine)
			currentLine = word
		}
	}
	lines = append(lines, currentLine)
	return lines
}

// FormatTableRow จัดรูปแบบแถวตาราง
func (pf *PrintFormatter) FormatTableRow(columns []string, widths []int) string {
	if len(columns) == 0 || len(widths) == 0 {
		return ""
	}

	var result string
	for i, col := range columns {
		if i >= len(widths) {
			break
		}

		// ตัดข้อความถ้ายาวเกินไป
		colLength := utf8.RuneCountInString(col)
		if colLength > widths[i] {
			runes := []rune(col)
			if widths[i] >= 3 {
				maxLen := widths[i] - 3
				if maxLen > len(runes) {
					maxLen = len(runes)
				}
				col = string(runes[:maxLen]) + "..."
			} else {
				maxLen := widths[i]
				if maxLen > len(runes) {
					maxLen = len(runes)
				}
				col = string(runes[:maxLen])
			}
		}

		// เพิ่มช่องว่างให้พอดีกับความกว้างที่กำหนด
		format := fmt.Sprintf("%%-%ds", widths[i])
		result += fmt.Sprintf(format, col)
	}

	return pf.FormatLine(result)
}

// convertToBitmap แปลงเนื้อหาเป็น bitmap
func convertToBitmap(content []byte, paperSize string) ([]byte, error) {
	// ใช้ template ตามขนาดกระดาษ
	template := Templates[paperSize]
	if template.Width == 0 {
		template = Templates["80"] // ใช้ค่าเริ่มต้นถ้าไม่พบ template
	}

	var buf bytes.Buffer

	// Initialize printer
	buf.Write([]byte{0x1B, 0x40})       // Initialize
	buf.Write([]byte{0x1D, 0x21, 0x00}) // Normal size
	buf.Write([]byte{0x1B, 0x4D, 0x00}) // Font A
	buf.Write([]byte{0x1B, 0x33, 60})   // Line spacing

	// แปลงข้อความเป็นบรรทัด
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		buf.WriteString(line)
		buf.Write([]byte{0x0A}) // Line feed
	}

	// Feed and cut
	buf.Write([]byte{0x0A, 0x0A, 0x0A, 0x0A})
	buf.Write([]byte{0x1D, 0x56, 0x41, 0x03}) // Partial cut

	return buf.Bytes(), nil
}

// convertPNGToBitmap แปลง PNG เป็น bitmap
func convertPNGToBitmap(content []byte, paperSize string) ([]byte, error) {
	// ใช้ template ตามขนาดกระดาษ
	template := Templates[paperSize]
	if template.Width == 0 {
		template = Templates["80"] // ใช้ค่าเริ่มต้นถ้าไม่พบ template
	}

	var buf bytes.Buffer

	// Initialize printer
	buf.Write([]byte{0x1B, 0x40})       // Initialize
	buf.Write([]byte{0x1D, 0x21, 0x00}) // Normal size
	buf.Write([]byte{0x1B, 0x4D, 0x00}) // Font A
	buf.Write([]byte{0x1B, 0x33, 60})   // Line spacing

	// Set bitmap mode for QR Code
	buf.Write([]byte{0x1D, 0x2A})
	widthBytes := (template.Width + 7) / 8
	buf.WriteByte(byte(widthBytes & 0xFF))
	buf.WriteByte(byte(widthBytes >> 8))

	// Copy QR Code content directly
	buf.Write(content)

	// Feed and cut
	buf.Write([]byte{0x0A, 0x0A, 0x0A, 0x0A})
	buf.Write([]byte{0x1D, 0x56, 0x41, 0x03}) // Partial cut

	return buf.Bytes(), nil
}
