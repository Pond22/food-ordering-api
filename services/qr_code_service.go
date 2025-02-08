package service

import (
	"bytes"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"image"
	"image/png"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/disintegration/imaging"
	"github.com/fogleman/gg"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

// func generateQRCode(tableID string, url string) error {
// 	png, err := qrcode.Encode(url, qrcode.Medium, 256)
// 	if err != nil {
// 		return err
// 	}

// 	err = os.WriteFile(fmt.Sprintf("qr_code_%s.png", tableID), png, 0644)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

func GenerateQRCodeAsBytes(url string) ([]byte, error) {
	qrCode, err := qrcode.Encode(url, qrcode.Medium, 256) // ขนาด 256x256
	if err != nil {
		return nil, err
	}
	return qrCode, nil
}

func SaveQRCode(qrCode models.QRCode) error {
	result := db.DB.Create(&qrCode)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func hasGroupID(groupID *string) bool {
	return groupID != nil && *groupID != ""
}

func updateTableStatus(tx *gorm.DB, table models.Table, tableID int) error {
	// ถ้าโต๊ะมี group_id ให้อัปเดตทุกโต๊ะในกลุ่ม
	if table.GroupID != nil {
		return tx.Model(&models.Table{}).
			Where("group_id = ?", table.GroupID).
			Updates(map[string]interface{}{
				"status":     "occupied",
				"updated_at": time.Now(),
			}).Error
	}

	// ถ้าไม่มี group_id ให้อัปเดตเฉพาะโต๊ะนั้น
	return tx.Model(&models.Table{}).
		Where("id = ?", tableID).
		Updates(map[string]interface{}{
			"status":     "occupied",
			"updated_at": time.Now(),
		}).Error
}

// @Summary เข้าถึง dynamic link โต๊ะนั้นๆ
// @Description เข้าสู่โต๊ะนั้นๆ ซึ่ง api เส้นนี้ไม่จำเป็นต้องถูกใช้งานโดยตรงเพราะ url ของแต่ละโต๊ะจะสามารถเข้าได้ผ่าน qr_code เท่านั้นจากฟังก์ชัน
// @Produce json
// @Param id path integer true "ID โต๊ะ"
// @Success 200 {object} models.QRCode "รายละเอียดของตาราง qr_code"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง"
// @Router /api/qr/{id} [get]
// @Tags Qr_code
func HandleQRCodeRequest(c *fiber.Ctx) error {
	UUID := uuid.New().String()
	tableID := c.Params("id")
	num, err := strconv.Atoi(tableID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid ID: %s. ID must be a number.", tableID),
		})
	}

	// ตรวจสอบว่าโต๊ะมีอยู่จริง
	var table models.Table
	if err := db.DB.First(&table, num).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Table not found",
		})
	}

	actualTableID := num
	if hasGroupID(table.GroupID) && table.ParentID != nil {
		actualTableID = int(*table.ParentID)
	}

	if table.Status != "available" {
		return c.Status(500).JSON(fiber.Map{
			"error": "โต๊ะต้องพร้อมใช้งาน",
		})
	}

	var existingQR models.QRCode
	result := db.DB.Where("tableID = ? AND is_active = ?", actualTableID, true).First(&existingQR)
	if result.Error == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "table_id นี้มีคิวอาร์กำลังใช้งานอยู่",
		})
	}

	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// อัพเดทสถานะโต๊ะ
	if err := updateTableStatus(tx, table, actualTableID); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update table status",
		})
	}

	// สร้าง QR Code
	expiryAt := time.Now().Add(2 * time.Hour)
	url := fmt.Sprintf("http://localhost:5173/menu?tableID=%v&uuid=%v", tableID, UUID)

	// สร้าง QR Code image
	qrCode, err := qrcode.New(url, qrcode.Medium)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate QR Code",
		})
	}

	// แปลงเป็น PNG bytes สำหรับแสดงผล
	displayImageData, err := qrCode.PNG(256)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to generate QR Code PNG",
			"details": err.Error(),
		})
	}

	// สร้าง QR Code พร้อมข้อความ
	finalQRCodeImage, err := createQRCodeWithText(displayImageData, table.Name, "./logo.jpg")
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create QR Code with text",
			"details": err.Error(),
		})
	}

	// บันทึก QR Code ในฐานข้อมูล
	qrCodeRecord := models.QRCode{
		TableID:   actualTableID,
		UUID:      UUID,
		CreatedAt: time.Now(),
		ExpiryAt:  expiryAt,
		IsActive:  true,
		Qr_Image:  finalQRCodeImage,
	}

	if err := tx.Create(&qrCodeRecord).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to save QR Code",
		})
	}

	// หา main printer
	var mainPrinter models.Printer
	if err := tx.Where("name = ?", "main").First(&mainPrinter).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Main printer not found",
		})
	}

	// เตรียมเนื้อหาสำหรับการพิมพ์
	// var printContent bytes.Buffer

	// headerLines := []string{
	// 	"===== QR Code สำหรับโต๊ะ =====",
	// 	fmt.Sprintf("วันที่-เวลา: %s", time.Now().Format("02/01/2006 15:04:05")),
	// 	"========================",
	// 	"\n", // เว้นบรรทัดก่อน QR Code
	// }

	// for _, line := range headerLines {
	// 	printContent.WriteString(line + "\n")
	// }

	// printContent.Write(finalQRCodeImage)

	// footerLines := []string{
	// 	"\n", // เว้นบรรทัดหลัง QR Code
	// 	"========================",
	// 	"โปรดสแกน QR Code เพื่อเข้าสู่ระบบ",
	// 	"========================",
	// }

	// for _, line := range footerLines {
	// 	printContent.WriteString("\n" + line)
	// }

	// สร้าง print job
	printJob := models.PrintJob{
		PrinterID: mainPrinter.ID,
		Content:   finalQRCodeImage,
		Status:    "pending",
		JobType:   "qr_code",
		CreatedAt: time.Now(),
	}

	if err := tx.Create(&printJob).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create print job",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	// ส่ง QR Code image กลับไปแสดงผล
	c.Set("Content-Type", "image/png")
	c.Set("Content-Disposition", "inline")
	return c.Send(finalQRCodeImage)
}

// table_service
// query หรือ path
// @Summary เข้าถึง dynamic link โต๊ะนั้นๆ
// @Description เข้าสู่โต๊ะนั้นๆ ซึ่ง api เส้นนี้ไม่จำเป็นต้องถูกใช้งานโดยตรงเพราะ url ของแต่ละโต๊ะจะสามารถเข้าได้ผ่าน qr_code เท่านั้นจากฟังก์ชัน
// @Produce json
// @Param table query string true "เลขโต๊ะ"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่ค้นพบจากการค้นหา"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง"
// @Router /order [get]
// @Tags tables
func Table(c *fiber.Ctx) error {
	tableID := c.Query("table") // ดึง Table ID จาก query string
	uuid := c.Query("uuid")
	fmt.Printf("TableIDDDDDD = %v \n", tableID)
	fmt.Printf("uuid = %v \n", uuid)

	if tableID == "" || uuid == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Both table ID and UUID are required",
		})
	}

	// โหลดข้อมูลโต๊ะ (หรือสร้าง session)
	// เช่น ตรวจสอบว่าโต๊ะนี้มีการเปิด session แล้วหรือยัง
	return c.JSON(fiber.Map{
		"message": "Welcome to the ordering page",
		"tableID": tableID,
		"uuid":    uuid,
	})
}

func isPNG(data []byte) bool {
	// PNG signature
	pngSignature := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}

	// เพิ่ม log เพื่อตรวจสอบ
	if len(data) > 8 {
		log.Printf("First 8 bytes: %v", data[:8])
		log.Printf("Matches PNG Signature: %v", bytes.Equal(data[:8], pngSignature))
	}

	return len(data) > 8 && bytes.Equal(data[:8], pngSignature)
}

func createQRCodeWithText(displayImageData []byte, actualTableID string, logoImagePath string) ([]byte, error) {
	// สร้าง reader จาก QR Code
	imgReader := bytes.NewReader(displayImageData)
	qrCodeImage, _, err := image.Decode(imgReader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode QR Code image: %v", err)
	}

	// ดึงขนาดของ QR Code
	qrWidth := qrCodeImage.Bounds().Dx()
	qrHeight := qrCodeImage.Bounds().Dy()

	// โหลดโลโก้ร้าน
	logoFile, err := os.Open(logoImagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open logo image: %v", err)
	}
	defer logoFile.Close()

	logoImg, _, err := image.Decode(logoFile)
	if err != nil {
		return nil, fmt.Errorf("failed to decode logo image: %v", err)
	}

	// ปรับขนาดโลโก้ให้ใหญ่ขึ้น
	logoWidth := qrWidth    // ใช้ความกว้างเต็ม QR
	logoHeight := logoWidth // รักษาสัดส่วน 1:1

	// เพิ่มพื้นที่ด้านบนให้มากขึ้น
	paddingTop := logoHeight + 80 // เพิ่มพื้นที่ตามขนาดโลโก้
	newHeight := qrHeight + paddingTop

	// สร้าง context และเติมพื้นหลัง
	dc := gg.NewContext(qrWidth, newHeight)
	dc.SetRGB(1, 1, 1)
	dc.Clear()

	// วาดโลโก้
	resizedLogo := imaging.Resize(logoImg, logoWidth, logoHeight, imaging.Lanczos)
	logoX := (float64(qrWidth) - float64(logoWidth)) / 2
	dc.DrawImageAnchored(resizedLogo, int(logoX), 10, 0, 0)

	// โหลดฟอนต์
	face, err := gg.LoadFontFace("THSarabunNew.ttf", 52)
	if err != nil {
		return nil, fmt.Errorf("failed to load font: %v", err)
	}
	dc.SetFontFace(face)

	// ข้อความที่ต้องการวาด
	label := fmt.Sprintf("โต๊ะ %s", actualTableID)
	textWidth, _ := dc.MeasureString(label)

	// คำนวณตำแหน่งข้อความ
	x := (float64(qrWidth) - textWidth) / 2
	y := float64(logoHeight + 50)

	// วาดข้อความ
	dc.SetRGB(0, 0, 0)
	dc.DrawString(label, x, y)

	// วาดเส้นตกแต่งด้านซ้ายและขวาของข้อความ
	lineY := y - 8            // ปรับตำแหน่ง Y ของเส้นให้อยู่ใกล้ข้อความ
	lineLength := float64(40) // ความยาวของเส้นแต่ละข้าง

	// เส้นด้านซ้าย
	dc.DrawLine(x-lineLength-10, lineY, x-10, lineY)
	dc.SetLineWidth(3)
	dc.Stroke()

	// เส้นด้านขวา
	dc.DrawLine(x+textWidth+10, lineY, x+textWidth+lineLength+10, lineY)
	dc.SetLineWidth(3)
	dc.Stroke()

	// วาด QR Code
	dc.DrawImage(qrCodeImage, 0, paddingTop)

	// สร้าง PNG ใหม่
	var buf bytes.Buffer
	if err := png.Encode(&buf, dc.Image()); err != nil {
		return nil, fmt.Errorf("failed to encode final image: %v", err)
	}

	return buf.Bytes(), nil
}
