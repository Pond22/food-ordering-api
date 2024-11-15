package qr_service

import (
	"fmt"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

func generateQRCode(tableID string, url string) error {
	png, err := qrcode.Encode(url, qrcode.Medium, 256)
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("qr_code_%s.png", tableID), png, 0644)
	if err != nil {
		return err
	}

	return nil
}

//qr handler

// @Summary เข้าถึง dynamic link โต๊ะนั้นๆ
// @Description เข้าสู่โต๊ะนั้นๆ ซึ่ง api เส้นนี้ไม่จำเป็นต้องถูกใช้งานโดยตรงเพราะ url ของแต่ละโต๊ะจะสามารถเข้าได้ผ่าน qr_code เท่านั้นจากฟังก์ชัน
// @Produce json
// Param action query string true "เลขโต๊ะ"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่ค้นพบจากการค้นหา"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง"
// @Router /qr_code [get]
func HandleQRCodeRequest(c *fiber.Ctx) error {
	// สร้าง tableID ใหม่ด้วย UUID
	tableID := uuid.New().String()

	// สร้าง URL ที่จะใช้ใน QR Code
	url := fmt.Sprintf("http://localhost:8080/order?table=%v", tableID)

	fmt.Println("Generated tableID:", tableID)

	// สร้าง QR Code สำหรับ tableID ใหม่
	err := generateQRCode(tableID, url)
	if err != nil {
		return c.Status(500).SendString(err.Error())
	}

	// ส่งข้อมูลกลับไปยังผู้ใช้
	return c.SendFile(fmt.Sprintf("qr_code_%s.png", tableID))
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

// @Summary เข้าถึง dynamic link โต๊ะนั้นๆ
// @Description เข้าสู่โต๊ะนั้นๆ ซึ่ง api เส้นนี้ไม่จำเป็นต้องถูกใช้งานโดยตรงเพราะ url ของแต่ละโต๊ะจะสามารถเข้าได้ผ่าน qr_code เท่านั้นจากฟังก์ชัน
// @Produce json
// @Param action query string true "เลขโต๊ะ"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่ค้นพบจากการค้นหา"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง"
// @Router /order [get]
func Table(c *fiber.Ctx) error {
	tableID := c.Query("table") // ดึง Table ID จาก query string
	if tableID == "" {
		return c.Status(400).JSON(fiber.Map{"error": "Table ID is required"})
	}

	// โหลดข้อมูลโต๊ะ (หรือสร้าง session)
	// เช่น ตรวจสอบว่าโต๊ะนี้มีการเปิด session แล้วหรือยัง
	return c.JSON(fiber.Map{
		"message": "Welcome to the ordering page",
		"tableID": tableID,
	})
}

// @Summary สั่งอาหาร
// @Description สั่งอาหารของโต๊ะ
// @Produce json
// @Param action query string true "เลขโต๊ะ"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่ค้นพบจากการค้นหา"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง"
// @Router /order [post]
func Order(c *fiber.Ctx) error {
	var order struct {
		TableID string   `json:"tableID"`
		Items   []string `json:"items"` // รายการอาหาร
	}
	if err := c.BodyParser(&order); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Invalid input"})
	}

	// บันทึกคำสั่งซื้อในฐานข้อมูล
	fmt.Printf("Order received: %+v\n", order)
	return c.JSON(fiber.Map{
		"message": "Order received successfully",
		"order":   order,
	})
}
