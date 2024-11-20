package api_handlers

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
)

// @Summary สร้างเมนูใหม่
// @Description ฟังก์ชันนี้ใช้สำหรับสร้างเมนูใหม่ โดยต้องระบุข้อมูลที่จำเป็นในการสร้าง เช่น ชื่อเมนูและ ID ของหมวดหมู่ที่เกี่ยวข้อง
// @Accept json
// @Produce json
// @Param request body models.CreateMenuRequest true "ข้อมูลเมนูใหม่"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างเมนูใหม่"
// @Router /add_menu [post]
func CreateMenuItemHandler(c *fiber.Ctx) error {
	var menuItem models.MenuItem

	if err := c.BodyParser(&menuItem); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Invalid input",
		})
	}

	if menuItem.Name == "" { //if menuItem.Name == "" || menuItem.CategoryID == 0 {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Name and CategoryID are required",
		})
	}

	var category models.Category
	if err := db.DB.First(&category, "ID = ?", menuItem.CategoryID).Error; err != nil {
		// ถ้าไม่พบ CategoryID ในตาราง categories
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Invalid CategoryID, category not found",
		})
	}

	if err := db.DB.First(&menuItem, "Name = ?", menuItem.Name).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "Name name already exists",
		})
	}

	if err := db.DB.Create(&menuItem).Error; err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Error creating menu item: %v", err),
		})
	}

	return c.Status(http.StatusOK).JSON(menuItem)
}

// @Summary สร้างหมวดหมู่ใหม่
// @Description ฟังก์ชันนี้ใช้สำหรับสร้างหมวดหมู่ใหม่ โดยต้องระบุข้อมูลชื่อหมวดหมู่
// @Accept json
// @Produce json
// @Param category body models.CreateCategoryRequest true "ข้อมูลหมวดหมู่ใหม่"
// @Success 200 {object} models.Category "รายละเอียดของหมวดหมู่ที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างหมวดหมู่ใหม่"
// @Router /add_category [post]
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

// @Summary เรียกรายการหมวดหมู่ทั้งหมด
// @Description ฟังก์ชันนี้ใช้สำหรับเรียกข้อมูลหมวดหมู่ทั้งหมดที่มีอยู่ในระบบ
// @Produce json
// @Success 200 {array} models.Category "รายการหมวดหมู่ทั้งหมด"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูลหมวดหมู่"
// @Router /getCategory [get]
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

// @Summary เรียกรายการเมนูทั้งหมด
// @Description ฟังก์ชันนี้ใช้สำหรับเรียกรายการเมนูทั้งหมดที่มีอยู่ในระบบ
// @Produce json
// @Success 200 {array} models.MenuItem "รายการเมนูทั้งหมด"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูลเมนู"
// @Router /menu [get]
func GetMenuAll(c *fiber.Ctx) error {
	var menuItem []models.MenuItem
	// ค้นหาทุก Category
	if err := db.DB.Find(&menuItem).Error; err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Error fetching menuItem: %v", err),
		})
	}
	// ส่งรายการ categories กลับในรูปแบบ JSON
	return c.JSON(menuItem)
}

// @Summary ดึงข้อมูลเมนูตาม ID
// @Description ฟังก์ชันนี้ใช้สำหรับดึงข้อมูลของเมนูโดยการระบุ ID ของเมนูนั้น
// @Produce json
// @Param ID query string true "ID ของเมนู"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่ค้นพบตาม ID"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก ID ที่ไม่ถูกต้อง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูลเมนู"
// @Router /menu [get]
func GetMenuByID(c *fiber.Ctx) error {
	var menuItem models.MenuItem

	// รับพารามิเตอร์ ID จาก URL
	idParam := c.Query("ID")
	num, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Invalid ID: %s. ID must be a number.", idParam),
		})
	}

	if err := db.DB.First(&menuItem, "id = ?", num).Error; err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Error fetching menuItem: %v", err),
		})
	}
	// ส่งรายการ categories กลับในรูปแบบ JSON
	return c.JSON(menuItem)
}

// @Summary เลือกการดำเนินการกับข้อมูลเมนู
// @Description ฟังก์ชันนี้ใช้สำหรับเรียกข้อมูลเมนู โดยสามารถระบุ action ได้ 3 แบบ
// @Produce json
// @Param action query string true "รูปแบบการค้นหาเมนู: getByID, getByCategory หรือ getAll" Enums(getByID, getByCategory, getAll)
// @Param id query integer false "ID ของเมนู (ใช้กับ action=getByID)"
// @Param category_id query integer false "ID ของหมวดหมู่ (ใช้กับ action=getByCategory)"
// @Success 200 {array} models.MenuItem "รายการเมนูที่ค้นพบ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากการระบุพารามิเตอร์"
// @Router /menu [get]
func GetMenuByCategory(c *fiber.Ctx) error {
	var menuItems []models.MenuItem

	// รับ CategoryID จาก query parameter
	categoryID := c.Query("category_id")

	// ตรวจสอบว่า CategoryID มีค่าอยู่หรือไม่
	if categoryID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error": "CategoryID is required",
		})
	}

	if err := db.DB.Preload("Category").Where("category_id = ?", categoryID).Find(&menuItems).Error; err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Error fetching menu items: %v", err),
		})
	}

	// ส่งรายการ MenuItems ที่ค้นพบกลับในรูปแบบ JSON
	return c.JSON(menuItems)
}

// @Summary เลือกการดำเนินการกับข้อมูลเมนู
// @Description ฟังก์ชันนี้ใช้สำหรับเรียกข้อมูลเมนู โดยสามารถระบุ action ได้ 3 แบบ
// @Produce json
// @Param action query string true "รูปแบบการค้นหาเมนู: getByID, getByCategory หรือ getAll" Enums(getByID, getByCategory, getAll)
// @Param id query integer false "ID ของเมนู (ใช้กับ action=getByID)"
// @Param category_id query integer false "ID ของหมวดหมู่ (ใช้กับ action=getByCategory)"
// @Success 200 {array} models.MenuItem "รายการเมนูที่ค้นพบ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากการระบุพารามิเตอร์"
// @Router /menu [get]
func GetMenu(c *fiber.Ctx) error {
	action := c.Query("action") // รับ action จาก query parameter

	switch action {
	case "getByID":
		// เรียกฟังก์ชันสำหรับสร้างข้อมูล
		return GetMenuByID(c)
	case "getByCategory":
		// เรียกฟังก์ชันสำหรับอัปเดตข้อมูล
		return GetMenuByCategory(c)
	case "getAll":
		// เรียกฟังก์ชันสำหรับลบข้อมูล
		return GetMenuAll(c)
	default:
		// กรณีไม่มี action ที่ตรงกัน
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Invalid action",
		})
	}
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
		TableID: orderReq.TableID,
		Status:  "pending",
		Total:   0, // Will calculate later
		Items:   []models.OrderItem{},
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
