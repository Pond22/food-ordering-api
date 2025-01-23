package api_handlers

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"

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
// @Success 200 {object} models.Category "ลบหมวดหมู่สำเร็จ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต (Unauthorized)"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง (Forbidden)"
// @Failure 404 {object} map[string]interface{} "ไม่พบหมวดหมู่ที่ต้องการแก้ไข"
// @Failure 409 {object} map[string]interface{} "ชื่อหมวดหมู่ซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการอัพเดตหมวดหมู่"
// @Router /api/categories/{id} [delete]
// @Tags categories
func Delete_categoryHandler(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Category ID is required",
		})
	}

	tx := db.DB.Begin()

	// ตรวจสอบว่ามีหมวดหมู่อยู่จริง
	var existingCategory models.Category
	if err := tx.First(&existingCategory, id).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error": "Category not found",
		})
	}

	// ลบเมนูในหมวดหมู่ (soft delete)
	// สังเกตว่าไม่ต้องไปยุ่งกับ options เพราะถูก cascade จาก menu item อยู่แล้ว
	if err := tx.Where("category_id = ?", id).Delete(&models.MenuItem{}).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to delete menu items",
		})
	}

	// ลบหมวดหมู่ (soft delete)
	if err := tx.Delete(&existingCategory).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to delete category",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to commit transaction",
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"message": "ลบหมวดหมู่และเมนูที่เกี่ยวข้องสำเร็จ",
	})
}

type updateCat struct {
	Name   string `json:"name,omitempty"`
	NameEn string `json:"nameEn,omitempty"`
	NameCh string `json:"nameCh,omitempty"`
}

// @Summary อัพเดตชื่อหมวดหมู่
// @Description ฟังก์ชันนี้ใช้สำหรับแก้ไขชื่อของหมวดหมู่ที่มีอยู่แล้ว โดยระบุ ID และชื่อใหม่
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของหมวดหมู่"
// @Param category body updateCat true "ข้อมูลหมวดหมู่ที่ต้องการอัพเดต"
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
	var updatedCategory updateCat
	if err := c.BodyParser(&updatedCategory); err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Invalid input",
		})
	}

	// ตรวจสอบว่ามีการส่งข้อมูลที่จะอัปเดตมาอย่างน้อย 1 ฟิลด์
	if updatedCategory.Name == "" && updatedCategory.NameEn == "" && updatedCategory.NameCh == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "At least one field (name, nameEn, or nameCh) is required for update",
		})
	}

	// ค้นหาหมวดหมู่เดิม
	var existingCategory models.Category
	if err := db.DB.First(&existingCategory, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error": "Category not found",
		})
	}

	// ตรวจสอบความซ้ำซ้อนของชื่อแต่ละภาษา
	if updatedCategory.Name != "" {
		var duplicateCheck models.Category
		if err := db.DB.Where("name = ? AND id != ?", updatedCategory.Name, id).First(&duplicateCheck).Error; err == nil {
			return c.Status(http.StatusConflict).JSON(map[string]interface{}{
				"error": "Category name (Thai) already exists",
			})
		}
		existingCategory.Name = updatedCategory.Name
	}

	if updatedCategory.NameEn != "" {
		var duplicateCheck models.Category
		if err := db.DB.Where("name_en = ? AND id != ?", updatedCategory.NameEn, id).First(&duplicateCheck).Error; err == nil {
			return c.Status(http.StatusConflict).JSON(map[string]interface{}{
				"error": "Category name (English) already exists",
			})
		}
		existingCategory.NameEn = updatedCategory.NameEn
	}

	if updatedCategory.NameCh != "" {
		var duplicateCheck models.Category
		if err := db.DB.Where("name_ch = ? AND id != ?", updatedCategory.NameCh, id).First(&duplicateCheck).Error; err == nil {
			return c.Status(http.StatusConflict).JSON(map[string]interface{}{
				"error": "Category name (Chinese) already exists",
			})
		}
		existingCategory.NameCh = updatedCategory.NameCh
	}

	// อัปเดตข้อมูลในฐานข้อมูล
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

type respon_cat struct {
	ID     uint              `json:"id"`
	Name   string            `json:"Name"`
	NameEn string            `json:"NameEn"`
	NameCh string            `json:"NameCh"`
	Menus  []models.MenuItem `json:"menus"`
}

// @Summary กู้คืนมวดหมู่และรายการอาหารทั้งหมดในนั้น
// @Description ฟังก์ชันนี้ใช้สำหรับลบหมวดหมู่และถ้ามีรายการอาหารจะถูกลบด้วย
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของหมวดหมู่"
// @Success 200 {array}  models.Category "ลบสำเร็จ"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต (Unauthorized)"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง (Forbidden)"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการลบข้อมูลหมวดหมู่"
// @Router /api/categories/restore_categories/{id} [post]
// @Tags categories
func Restore_categoryHandler(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Category ID is required",
		})
	}

	tx := db.DB.Begin()

	// ตรวจสอบว่ามีหมวดหมู่อยู่จริง
	var existingCategory models.Category
	if err := tx.Unscoped().Where("id = ? AND deleted_at IS NOT NULL", id).First(&existingCategory, id).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error": "Category not found",
		})
	}

	var duplicateCategory models.Category
	if err := tx.Where("name = ? AND deleted_at IS NULL", existingCategory.Name).First(&duplicateCategory).Error; err == nil {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(map[string]interface{}{
			"error": "Category name already exists",
		})
	}

	// Restore เมนูที่ถูก soft delete ในหมวดหมู่นั้น
	if err := tx.Unscoped().Model(&models.MenuItem{}).Where("category_id = ? AND deleted_at IS NOT NULL", id).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to restore menu items",
		})
	}
	// กู้หมวดหมู่
	if err := tx.Unscoped().Model(&existingCategory).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to restore category",
		})
	}

	var restoredCategory models.Category
	if err := tx.First(&restoredCategory, id).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to load restored data",
		})
	}

	var menu []models.MenuItem
	if err := tx.Preload("OptionGroups.Options").Where("category_id = ?", id).Find(&menu).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to load restored data",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to commit transaction",
		})
	}

	response := respon_cat{
		ID:     restoredCategory.ID,
		Name:   restoredCategory.Name,
		NameEn: restoredCategory.NameEn,
		NameCh: restoredCategory.NameCh,
		Menus:  menu,
	}

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"message":  "กู้คืนหมวดหมู่และเมนูที่เกี่ยวข้องสำเร็จ",
		"category": response,
	})
}

// @Summary เรียกดูหมวดหมู่ที่ถูกลบพร้อมแสดงรายการอาหารในนั้นถ้ามี
// @Description ฟังก์ชันนี้ใช้สำหรับเรียกดูหมวดหมู่ที่ถูกลบรวมถึงอาหารในนั้นถ้าถูกลบพร้อมหมวดหมู่
// @Produce json
// @Security BearerAuth
// @Success 200 {array}  models.Category "สำเร็จ"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต (Unauthorized)"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง (Forbidden)"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการเรียกดูข้อมูลหมวดหมู่ที่ถูกลบ"
// @Router /api/categories/get_delete_categories [get]
// @Tags categories
func Get_Delete_Cat(c *fiber.Ctx) error {
	var deleteCat []models.Category

	if err := db.DB.Unscoped().Where("deleted_at IS NOT NULL").Find(&deleteCat).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to load Delete data",
		})
	}

	return c.JSON(deleteCat)
}