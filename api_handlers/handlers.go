package api_handlers

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// @Summary สร้างเมนูใหม่
// @Description ฟังก์ชันนี้ใช้สำหรับสร้างเมนูใหม่ โดยต้องระบุข้อมูลที่จำเป็นในการสร้าง เช่น ชื่อเมนูและ ID ของหมวดหมู่ที่เกี่ยวข้อง
// @Accept json
// @Produce json
// @Param menuItem body models.MenuItem true "ข้อมูลเมนูใหม่"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างเมนูใหม่"
// @Router /menu [post]
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
	if err := db.DB.First(&category, menuItem.CategoryID).Error; err != nil {
		// ถ้าไม่พบ CategoryID ในตาราง categories
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Invalid CategoryID, category not found",
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
// @Param category body models.Category true "ข้อมูลหมวดหมู่ใหม่"
// @Success 200 {object} models.Category "รายละเอียดของหมวดหมู่ที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างหมวดหมู่ใหม่"
// @Router /categories [post]
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

// @Summary ดึงรายการเมนูตามหมวดหมู่
// @Description ฟังก์ชันนี้ใช้สำหรับดึงข้อมูลเมนูทั้งหมดที่อยู่ในหมวดหมู่ที่ระบุ โดยระบุ ID ของหมวดหมู่นั้น
// @Produce json
// @Param category_id query string true "ID ของหมวดหมู่"
// @Success 200 {array} models.MenuItem "รายการเมนูที่อยู่ในหมวดหมู่ที่ระบุ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก ID ของหมวดหมู่ที่ไม่ถูกต้อง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูลเมนูตามหมวดหมู่"
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

	// ค้นหา MenuItems ตาม CategoryID ที่ระบุ
	if err := db.DB.Where("category_id = ?", categoryID).Find(&menuItems).Error; err != nil {
		return c.Status(500).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Error fetching menu items: %v", err),
		})
	}

	// ส่งรายการ MenuItems ที่ค้นพบกลับในรูปแบบ JSON
	return c.JSON(menuItems)
}

// @Summary เลือกการดำเนินการกับข้อมูลเมนู
// @Description ฟังก์ชันนี้ใช้สำหรับเรียกข้อมูลเมนู โดยสามารถระบุ action ได้ 3 แบบ คือ getByID, getByCategory, และ getAll เพื่อกำหนดรูปแบบการค้นหาเมนูตาม action ที่ระบุเนื่องจากใช้หลักการ Switchcase ทำให้ใน Document มีตัวอย่างเพัยงแค่การเลือก Actions
// @Produce json
// @Param action query string true "รูปแบบการค้นหาเมนู: getByID, getByCategory หรือ getAll"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่ค้นพบจากการค้นหา"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก action ที่ไม่ถูกต้อง"
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
