package api_handlers

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"

	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
)

// @Summary สร้างหมวดหมู่ใหม่
// @Description ฟังก์ชันนี้ใช้สำหรับสร้างหมวดหมู่ใหม่ โดยต้องระบุข้อมูลชื่อหมวดหมู่
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param category body models.CreateCategoryRequest true "ข้อมูลหมวดหมู่ใหม่"
// @Success 200 {object} models.Category "รายละเอียดของหมวดหมู่ที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต (Unauthorized)"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง (Forbidden)"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างหมวดหมู่ใหม่"
// @Router /api/categories [post]
// @Tags categories
func CreateCategoryHandler(c *fiber.Ctx) error {
	var category models.Category

	if err := c.BodyParser(&category); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Invalid input",
		})
	}

	if category.Name == "" { //if menuItem.Name == "" || menuItem.CategoryID == 0 {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Name  are required",
		})
	}

	// db.DB.First(&category, "Name = ?", category.Name)
	if err := db.DB.First(&category, "Name = ?", category.Name).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "Category name already exists",
		})
	}
	if err := db.DB.Create(&category).Error; err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Error creating category: %v", err),
		})
	}

	return c.Status(http.StatusOK).JSON(category)
}

type DeleteCategoryOption struct {
	ForceDelete bool `json:"force_delete"` // true = ถ้าลบหมวดหมู่จะลบเมนูต่างๆ ที่เชื่อมอยู่ด้วย
}

// @Summary ลบหมวดหมู่
// @Description ฟังก์ชันนี้ใช้สำหรับลบหมวดหมู่ถ้าต้องการลบทั้งหมดรวมถึงอาหารในหมวดหมู่ให้ใช้ true ลบแค่หมวดหมู่ false แต่ต้องระวังถ้าระบุ false แล้วมีเมนูในหมวดหมู่จะ error
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของหมวดหมู่"
// @Param order body DeleteCategoryOption true "ture ถ้ามีเมนูอยู่ในหมวดหมู่จะลบเมนูไปด้วย"
// @Success 200 {object} models.Category "ลบหมวดหมู่สำเร็จ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต (Unauthorized)"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง (Forbidden)"
// @Failure 404 {object} map[string]interface{} "ไม่พบหมวดหมู่ที่ต้องการแก้ไข"
// @Failure 409 {object} map[string]interface{} "ชื่อหมวดหมู่ซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการอัพเดตหมวดหมู่"
// @Router /api/categories/{id} [delete]
// @Tags categories
func Delete_categoryHandler(c *fiber.Ctx) error { //เหลือทำให้ลบเมนูไม่กระทบออเดอร์
	id := c.Params("id")
	fmt.Print(id)
	var option DeleteCategoryOption
	if err := c.BodyParser(&option); err != nil {
		option.ForceDelete = false
	}

	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Category ID is required",
		})
	}

	var existingCategory models.Category
	if err := db.DB.First(&existingCategory, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error": "Category not found",
		})
	}
	if option.ForceDelete {
		if err := db.DB.Where("category_id = ?", id).Delete(&models.MenuItem{}).Error; err != nil {
			return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
				"error": "Failed to delete menu items",
			})
		}
	}
	if err := db.DB.Delete(&existingCategory).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to delete category",
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"Massage": "ลบสำเร็จ",
	})

}

// @Summary อัพเดตชื่อหมวดหมู่
// @Description ฟังก์ชันนี้ใช้สำหรับแก้ไขชื่อของหมวดหมู่ที่มีอยู่แล้ว โดยระบุ ID และชื่อใหม่
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของหมวดหมู่"
// @Param category body models.Category true "ข้อมูลหมวดหมู่ที่ต้องการอัพเดต"
// @Success 200 {object} models.Category "รายละเอียดของหมวดหมู่ที่อัพเดตแล้ว"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Failure 404 {object} map[string]interface{} "ไม่พบหมวดหมู่ที่ต้องการแก้ไข"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต (Unauthorized)"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง (Forbidden)"
// @Failure 409 {object} map[string]interface{} "ชื่อหมวดหมู่ซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการอัพเดตหมวดหมู่"
// @Router /api/categories/{id} [put]
// @Tags categories
func UpdateCategoryHandler(c *fiber.Ctx) error {
	// รับ ID จาก parameter
	id := c.Params("id")
	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Category ID is required",
		})
	}

	// รับข้อมูลใหม่จาก request body
	var updatedCategory models.Category
	if err := c.BodyParser(&updatedCategory); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Invalid input",
		})
	}

	fmt.Print(id, updatedCategory.Name)

	// ตรวจสอบว่ามีชื่อที่จะอัพเดตหรือไม่
	if updatedCategory.Name == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Category name is required",
		})
	}

	// ค้นหาหมวดหมู่เดิม
	var existingCategory models.Category
	if err := db.DB.First(&existingCategory, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error": "Category not found",
		})
	}

	// ตรวจสอบว่าชื่อใหม่ซ้ำกับที่มีอยู่หรือไม่ (ยกเว้นตัวมันเอง)
	var duplicateCheck models.Category
	if err := db.DB.Where("name = ? AND id != ?", updatedCategory.Name, id).First(&duplicateCheck).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(map[string]interface{}{
			"error": "Category name already exists",
		})
	}

	// อัพเดตชื่อหมวดหมู่
	existingCategory.Name = updatedCategory.Name
	if err := db.DB.Save(&existingCategory).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Error updating category: %v", err),
		})
	}

	return c.JSON(existingCategory)
}

// @Summary เรียกรายการหมวดหมู่ทั้งหมด
// @Description ฟังก์ชันนี้ใช้สำหรับเรียกข้อมูลหมวดหมู่ทั้งหมดที่มีอยู่ในระบบ
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.Category "รายการหมวดหมู่ทั้งหมด"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต (Unauthorized)"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง (Forbidden)"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูลหมวดหมู่"
// @Router /api/categories [get]
// @Tags categories
func GetCategoriesHandler(c *fiber.Ctx) error {
	var categories []models.Category
	// ค้นหาทุก Category
	if err := db.DB.Find(&categories).Error; err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Error fetching categories: %v", err),
		})
	}
	// ส่งรายการ categories กลับในรูปแบบ JSON
	return c.JSON(categories)
}

// ---------------------------------------------------------------------
type OrderRequest struct {
	TableID uint               `json:"table_id"`
	Items   []OrderItemRequest `json:"items"`
	UUID    string             `json:"uuid"`
}

type OrderItemRequest struct {
	MenuItemID uint   `json:"menu_item_id"`
	Quantity   int    `json:"quantity"`
	Notes      string `json:"notes"`
}

// @Summary สั่งอาหาร......
// @Description สั่งอาหารสำหรับโต๊ะที่ระบุ
// @Accept json
// @Produce json
// @Param order body OrderRequest true "ข้อมูลการสั่งอาหาร"
// @Success 200 {object} models.Order "รายละเอียดออเดอร์ที่สร้าง"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการประมวลผล"
// @Router /order [post]
// @Tags orders
func Order_test(c *fiber.Ctx) error {

	var orderReq OrderRequest
	if err := c.BodyParser(&orderReq); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// ตรวจสอบ UUID
	if orderReq.UUID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "UUID is required",
		})
	}
	fmt.Printf(orderReq.UUID)
	// ตรวจสอบว่าโต๊ะมีอยู่จริง
	// var table models.Table
	// if err := db.DB.First(&table, orderReq.TableID).Error; err != nil {
	// 	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
	// 		"error": "Table not found",
	// 	})
	// }

	// ตรวจสอบว่าโต๊ะนี้มี QR code ที่ active อยู่หรือไม่
	var qrCode models.QRCode
	if err := db.DB.Where("table_id = ? AND uuid = ? AND is_active = true AND expiry_at > ?",
		orderReq.TableID, orderReq.UUID, time.Now()).First(&qrCode).Error; err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
			"error": "Invalid table access",
		})
	}

	// ตรวจสอบความถูกต้องของ UUID กับ table
	// if err := db.DB.Where("table_id = ? AND uuid = ? AND is_active = true AND expiry_at > ?",
	// 	orderReq.TableID, orderReq.UUID, time.Now()).First(&qrCode).Error; err != nil {
	// 	return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
	// 		"error": "Invalid or expired table access",
	// 	})
	// }

	// Validate order
	if len(orderReq.Items) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order must contain at least one item",
		})
	}

	// Start transaction
	tx := db.DB.Begin()

	// Create new order
	order := models.Order{
		UUID:   orderReq.UUID,
		Status: "pending",
		Total:  0, // Will calculate later
		Items:  []models.OrderItem{},
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order",
		})
	}

	// Process each order item
	var totalAmount float64
	for _, item := range orderReq.Items {
		// Get menu item to get current price
		var menuItem models.MenuItem
		if err := tx.First(&menuItem, item.MenuItemID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Menu item %d not found", item.MenuItemID),
			})
		}

		// Create order item
		orderItem := models.OrderItem{
			OrderID:    order.ID,
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			Price:      float64(menuItem.Price), // ราคาในฐานข้อมูล
			Notes:      item.Notes,
			Status:     "pending",
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create order item",
			})
		}

		// Calculate item total and add to order total
		totalAmount += float64(menuItem.Price) * float64(item.Quantity)
	}

	// Update order with total amount
	order.Total = totalAmount
	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order total",
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	var completeOrder models.Order
	if err := db.DB.Preload("Items.MenuItem.Category").
		Preload("Items.Order").
		Preload("Items").
		First(&completeOrder, order.ID).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load complete order",
		})
	}

	return c.JSON(completeOrder)
}