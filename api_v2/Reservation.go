package api_v2

import (
	"bytes"
	"errors"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"image"
	"image/png"
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

// @Summary ยกเลิกการจองโดยตรง
// @Description ยกเลิกการจองโดยอ้างอิงจาก ID การจอง
// @Accept json
// @Produce json
// @Param id path int true "ID ของการจอง"
// @Success 200 {object} map[string]interface{} "ยกเลิกการจองสำเร็จ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 404 {object} map[string]interface{} "ไม่พบการจอง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดภายในระบบ"
// @Router /api/v2/reservation/cancel/{id} [post]
// @Tags Reservation_V2
func CancelReservation(c *fiber.Ctx) error {
	reservationID := c.Params("id")
	if reservationID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "กรุณาระบุ ID การจอง",
		})
	}

	tx := db.DB.Begin()

	// ค้นหาการจอง
	var reservation models.TableReservation
	if err := tx.Where("id = ? AND status = ?", reservationID, "active").First(&reservation).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(404).JSON(fiber.Map{
				"error": "ไม่พบการจองที่ระบุ หรือการจองถูกยกเลิกไปแล้ว",
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการค้นหาการจอง",
		})
	}

	// อัพเดทสถานะการจอง
	if err := tx.Model(&reservation).Update("status", "cancelled").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถยกเลิกการจองได้",
		})
	}

	// ถ้าโต๊ะถูกกั้นไว้แล้ว (status = reserved) ให้คืนสถานะเป็น available
	if err := tx.Model(&models.Table{}).
		Where("id = ? AND status = ?", reservation.TableID, "reserved").
		Update("status", "available").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอัพเดทสถานะโต๊ะได้",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "เกิดข้อผิดพลาดในการบันทึกข้อมูล",
		})
	}

	return c.JSON(fiber.Map{
		"message": "ยกเลิกการจองสำเร็จ",
	})
}

// ใน tableHandlers.go เพิ่ม function ใหม่

// @Summary เช็คอินโต๊ะที่จองไว้
// @Description เปิดโต๊ะที่จองไว้เมื่อลูกค้ามาถึง
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID โต๊ะ"
// @Success 200 {object} map[string]interface{} "เช็คอินสำเร็จ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 404 {object} map[string]interface{} "ไม่พบการจอง"
// @Router /api/v2/reservation/checkin/{id} [post]
// @Tags Reservation_V2
func CheckinReservedTable(c *fiber.Ctx) error {
	tableID := c.Params("id")
	if tableID == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "กรุณาระบุ ID โต๊ะ",
		})
	}

	num, err := strconv.Atoi(tableID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": fmt.Sprintf("Invalid ID: %s. ID must be a number.", tableID),
		})
	}

	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// ตรวจสอบว่ามีการจองที่ active อยู่หรือไม่
	var reservation models.TableReservation
	if err := tx.Where("table_id = ? AND status = ?", num, "active").First(&reservation).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{
			"error": "ไม่พบการจองที่ยังใช้งานอยู่สำหรับโต๊ะนี้",
		})
	}

	// ตรวจสอบว่าโต๊ะมีอยู่จริง
	var table models.Table
	if err := tx.First(&table, num).Error; err != nil {
		tx.Rollback()
		return c.Status(404).JSON(fiber.Map{
			"error": "Table not found",
		})
	}

	// เช็คว่าโต๊ะอยู่ในกลุ่มหรือไม่
	actualTableID := num
	if hasGroupID(table.GroupID) && table.ParentID != nil {
		actualTableID = int(*table.ParentID)
	}

	// สร้าง QR Code
	UUID := uuid.New().String()
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

	// แปลงเป็น PNG bytes
	displayImageData, err := qrCode.PNG(256)
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate QR Code PNG",
		})
	}

	// สร้าง QR Code พร้อมข้อความ
	finalQRCodeImage, err := createQRCodeWithText(displayImageData, table.Name, "./logo.jpg")
	if err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create QR Code with text",
		})
	}

	// บันทึก QR Code
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

	// อัพเดทสถานะการจอง
	if err := tx.Model(&reservation).Update("status", "completed").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอัพเดทสถานะการจองได้",
		})
	}

	// อัพเดทสถานะโต๊ะ
	if err := tx.Model(&models.Table{}).Where("id = ?", tableID).Update("status", "occupied").Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "ไม่สามารถอัพเดทสถานะโต๊ะได้",
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
