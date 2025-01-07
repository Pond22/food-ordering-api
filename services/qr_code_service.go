package service

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
	// ถ้าโต๊ะอยู่ในกลุ่มและไม่ใช่ parent
	if hasGroupID(table.GroupID) && table.ParentID != nil {
		var parentTable models.Table
		if err := db.DB.First(&parentTable, table.ParentID).Error; err != nil {
			return c.Status(500).JSON(fiber.Map{
				"error": "Parent table not found",
			})
		}

		actualTableID = int(*table.ParentID)
		// // ส่ง error แจ้งให้ใช้ parent table แทน
		// return c.Status(400).JSON(fiber.Map{
		// 	"error": fmt.Sprintf("This table is part of a group. Please use parent table (ID: %d) instead", *table.ParentID),
		// })
	}

	if table.Status != "available" {
		return c.Status(500).JSON(fiber.Map{
			"error": "โต๊ะต้องพร้อมใช้งาน",
		})
	}

	var existingQR models.QRCode
	result := db.DB.Where("table_id = ? AND is_active = ?", actualTableID, true).First(&existingQR)

	if result.Error == nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "table_id นี้มีคิวอาร์กำลังใช้งานอยู่",
		})
	}

	tx := db.DB.Begin()

	// ในฟังก์ชัน HandleQRCodeRequest
	if err := updateTableStatus(tx, table, actualTableID); err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update table status",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	expiryAt := time.Now().Add(2 * time.Hour)
	url := fmt.Sprintf("http://localhost:8080/order?table=%v&uuid=%v", tableID, UUID)

	imageData, err := GenerateQRCodeAsBytes(url)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate QR Code",
		})
	}

	// ถ้าเป็น parent table หรือโต๊ะเดี่ยว
	qrCode := models.QRCode{
		TableID:   actualTableID,
		UUID:      UUID,
		CreatedAt: time.Now(),
		ExpiryAt:  expiryAt,
		IsActive:  true,
		Qr_Image:  imageData,
	}

	err = SaveQRCode(qrCode)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": err,
		})
	}

	c.Set("Content-Type", "image/png")
	c.Set("Content-Disposition", "inline")
	return c.Send(imageData)
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
