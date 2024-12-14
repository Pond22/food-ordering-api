package qr_service

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
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
	// ใช้ GORM บันทึกข้อมูล QRCode
	result := db.DB.Create(&qrCode) // `db` คือตัวแปรที่เชื่อมต่อ GORM
	if result.Error != nil {
		return result.Error
	}
	return nil
}

//qr handler

// @Summary เข้าถึง dynamic link โต๊ะนั้นๆ
// @Description เข้าสู่โต๊ะนั้นๆ ซึ่ง api เส้นนี้ไม่จำเป็นต้องถูกใช้งานโดยตรงเพราะ url ของแต่ละโต๊ะจะสามารถเข้าได้ผ่าน qr_code เท่านั้นจากฟังก์ชัน
// @Produce json
// @Param table path string true "เลขโต๊ะ"
// @Success 200 {object} models.QRCode "รายละเอียดของตาราง qr_code"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง"
// @Router /qr_code/{table} [get]
// @Tags tables
func HandleQRCodeRequest(c *fiber.Ctx) error {
	// สร้าง tableID ใหม่ด้วย UUID
	UUID := uuid.New().String()
	tableID := c.Params("table")
	num, err := strconv.Atoi(tableID)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Invalid ID: %s. ID must be a number.", tableID),
		})
	}
	expiryAt := time.Now().Add(2 * time.Hour)
	// สร้าง URL ที่จะใช้ใน QR Code
	url := fmt.Sprintf("http://localhost:8080/order?table=%v&uuid=%v", tableID, UUID)

	fmt.Println("tableID:", tableID)
	fmt.Println("Generated UUID:", UUID)
	// สร้าง QR Code สำหรับ tableID ใหม่
	// err := generateQRCode(tableID, url)
	// if err != nil {
	// 	return c.Status(500).SendString(err.Error())
	// }

	// ใช้ฟังก์ชันที่ส่งคืนข้อมูล []byte ของภาพ
	imageData, err := GenerateQRCodeAsBytes(url)
	if err != nil {
		return c.Status(500).SendString("Failed to generate QR Code")
	}

	qrCode := models.QRCode{
		// ID:        tableID,
		TableID:   num,
		UUID:      UUID,
		CreatedAt: time.Now(),
		ExpiryAt:  expiryAt,
		IsActive:  true,
		Qr_Image:  imageData,
	}

	err = SaveQRCode(qrCode)
	if err != nil {
		return c.Status(500).SendString("qr_code โต๊ะนี้มีอยู่ในระบบแล้ว")
	}

	// ส่งข้อมูลกลับไปยังผู้ใช้
	// return c.SendFile(fmt.Sprintf("qr_code_%s.png", tableID))
	c.Set("Content-Type", "image/png")
	c.Set("Content-Disposition", "inline")
	return c.Send(imageData)
	// return c.JSON(fiber.Map{
	// 	"message": "QR Code generated successfully",
	// 	"qr_url":  url,
	// 	"table":   tableID,
	// })
}

// func HandleQRCodeRequest(c *fiber.Ctx) error {
// 	tableID := c.Query("action")
// 	// url := fmt.Sprintf("https://example.com/order/%s", tableID)
// 	url := fmt.Sprintf("http://localhost:8080/order?table=%v", tableID)
// 	fmt.Println("Received tableID:", tableID)
// 	err := generateQRCode(tableID, url)
// 	if err != nil {
// 		return c.Status(500).SendString(err.Error())
// 	}

// 	// return c.SendFile(fmt.Sprintf("qr_code_%s.png", tableID))
// 	return c.JSON(fiber.Map{
// 		"message": "QR Code generated successfully",
// 		"qr_url":  url,
// 		"table":   tableID,
// 	})
// }

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
