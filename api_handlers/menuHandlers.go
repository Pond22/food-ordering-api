package api_handlers

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// @Summary เรียกรายการเมนูทั้งหมด
// @Description ฟังก์ชันนี้ใช้สำหรับเรียกรายการเมนูทั้งหมดที่มีอยู่ในระบบ
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.MenuItem "รายการเมนูทั้งหมด"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูลเมนู"
// @Router /api/menu [get]
// @Tags menu
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
// @Security BearerAuth
// @Param ID query string true "ID ของเมนู"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่ค้นพบตาม ID"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก ID ที่ไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูลเมนู"
// @Router /api/menu/{id} [get]
// @Tags menu
func GetMenuByID(c *fiber.Ctx) error {
	var menuItem models.MenuItem

	// รับพารามิเตอร์ ID จาก URL
	idParam := c.Query("id")

	num, err := strconv.Atoi(idParam)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": fmt.Sprintf("Invalid ID: %s. ID must be a number.", idParam),
		})
	}

	if err := db.DB.Preload("Category").Preload("OptionGroups").Preload("OptionGroups.Options").First(&menuItem, "id = ?", num).Error; err != nil {
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
// @Security BearerAuth
// @Param action query string true "รูปแบบการค้นหาเมนู: getByID, getByCategory หรือ getAll" Enums(getByID, getByCategory, getAll)
// @Param id query integer false "ID ของเมนู (ใช้กับ action=getByID)"
// @Param category_id query integer false "ID ของหมวดหมู่ (ใช้กับ action=getByCategory)"
// @Success 200 {array} models.MenuItem "รายการเมนูที่ค้นพบ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากการระบุพารามิเตอร์"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Router /api/menu/{id} [get]
// @Tags menu
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
// @Security BearerAuth
// @Param action query string true "รูปแบบการค้นหาเมนู: getByID, getByCategory หรือ getAll" Enums(getByID, getByCategory, getAll)
// @Param id query integer false "ID ของเมนู (ใช้กับ action=getByID)"
// @Param category_id query integer false "ID ของหมวดหมู่ (ใช้กับ action=getByCategory)"
// @Success 200 {array} models.MenuItem "รายการเมนูที่ค้นพบ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากการระบุพารามิเตอร์"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Router /api/menu [get]
// @Tags menu
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

// @Summary สร้างเมนูใหม่พร้อม options
// @Description สร้างเมนูอาหารใหม่พร้อมกับตัวเลือกเพิ่มเติม (options) ของเมนูนั้นๆ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateMenuRequest true "ข้อมูลเมนูและ options"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูและ options ที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 409 {object} map[string]interface{} "ชื่อเมนูซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างเมนูหรือ options"
// @Router /api/menu [post]
// @Tags menu
func CreateMenuItemHandler(c *fiber.Ctx) error {
	var req models.CreateMenuRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	if req.MenuItem.Name == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ชื่อเมนูจำเป็น",
		})
	}

	if req.MenuItem.Price < 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ราคาต้องไม่ติดลบ",
		})
	}

	tx := db.DB.Begin()

	var completeMenuItem models.MenuItem
	if err := tx.Where("name = ?", req.MenuItem.Name).First(&completeMenuItem).Error; err == nil {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "ชื่อเมนูนี้มีอยู่แล้ว",
		})
	}

	// if err := tx.Unscoped().Where("name = ?", req.MenuItem.Name).First(&completeMenuItem).Error; err == nil {
	// 	db.DB.Unscoped().Model(&completeMenuItem).Update("deleted_at", nil)
	// 	tx.Rollback()
	// 	return c.Status(http.StatusConflict).JSON(fiber.Map{
	// 		"massage": "เมนูนี้ถูก Softdelete เดี้ยวกู้ให้นะ :)",
	// 	})
	// }

	var deletedMenuItem models.MenuItem
	if err := tx.Unscoped().Where("name = ?", req.MenuItem.Name).First(&deletedMenuItem).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"status":  "deleted",
			"message": "เมนูนี้ถูก Soft Delete ไปแล้ว คุณต้องการกู้คืนหรือไม่?",
		})
	}

	// สร้าง MenuItem
	menuItem := models.MenuItem{
		Name:          req.MenuItem.Name,
		NameEn:        req.MenuItem.NameEn,
		NameCh:        req.MenuItem.NameCh,
		Description:   req.MenuItem.Description,
		DescriptionEn: req.MenuItem.DescriptionEn,
		DescriptionCh: req.MenuItem.DescriptionCh,
		CategoryID:    req.MenuItem.CategoryID,
		Price:         req.MenuItem.Price,
	}

	if err := tx.Create(&menuItem).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Error creating menu item"})
	}

	if len(req.OptionGroups) > 0 {
		for _, group := range req.OptionGroups {
			if group.MaxSelections == 0 {
				continue
			}

			optionGroup := models.OptionGroup{
				MenuItemID:    menuItem.ID,
				Name:          group.Name,
				NameEn:        group.NameEn,
				NameCh:        group.NameCh,
				MaxSelections: group.MaxSelections,
				IsRequired:    group.IsRequired,
			}

			if err := tx.Create(&optionGroup).Error; err != nil {
				tx.Rollback()
				return c.Status(500).JSON(fiber.Map{"error": "Error creating option group"})
			}

			if len(group.Options) > 0 {
				for _, opt := range group.Options {
					groupOption := models.MenuOption{
						GroupID: optionGroup.ID,
						Name:    opt.Name,
						NameEn:  opt.NameEn,
						NameCh:  opt.NameCh,
						Price:   opt.Price,
					}

					if err := tx.Create(&groupOption).Error; err != nil {
						tx.Rollback()
						return c.Status(500).JSON(fiber.Map{"error": "Error creating group option"})
					}
				}
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error committing transaction"})
	}

	if err := db.DB.Preload("Category").Preload("OptionGroups.Options").First(&completeMenuItem, menuItem.ID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading complete menu item"})
	}
	return c.JSON(completeMenuItem)
}

// @Summary เพิ่ม group options แยกถ้าสร้างเมนูไปแล้ว
// @Description สร้าง (group options) ของเมนูนั้นๆ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param menu_id query integer true "ID ของเมนูที่ต้องการเพิ่ม group option"
// @Param request body models.OptionGroupRequest true "ข้อมูลเมนูและ group options"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูและ group options ที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 409 {object} map[string]interface{} "ชื่อเมนูซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างเมนูหรือ options"
// @Router /api/menu/option-groups [post]
// @Tags menu
func AddMoreGroup(c *fiber.Ctx) error {
	id := c.Query("menu_id")
	menuId, err := strconv.Atoi(id)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error": "MenuID is required or หาไม่เจอ",
		})
	}
	var req models.OptionGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	if req.MaxSelections <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "MaxSelections ต้องมากกว่า 0",
		})
	}

	tx := db.DB.Begin()

	var opt models.OptionGroup
	if err := tx.Where("name = ? AND menu_item_id = ?", req.Name, menuId).First(&opt).Error; err == nil {
		tx.Rollback()
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "ชื่อ optgroup ซ้ำกับที่มีอยู่ในเมนูนี้",
		})
	}

	var jjjjjj models.MenuItem
	menu := db.DB.First(&jjjjjj, menuId)
	if menu.Error != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "หา id อาหารไม่เจอ",
		})
	}

	optionGroup := models.OptionGroup{
		MenuItemID:    jjjjjj.ID,
		Name:          req.Name,
		NameEn:        req.NameEn,
		NameCh:        req.NameCh,
		MaxSelections: req.MaxSelections,
		IsRequired:    req.IsRequired,
	}

	if err := tx.Create(&optionGroup).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{"error": "Error creating option group"})
	}

	if len(req.Options) > 0 {
		for _, opt := range req.Options {
			groupOption := models.MenuOption{
				GroupID: optionGroup.ID,
				Name:    opt.Name,
				NameEn:  opt.NameEn,
				NameCh:  opt.NameCh,
				Price:   opt.Price,
			}

			if err := tx.Create(&groupOption).Error; err != nil {
				tx.Rollback()
				return c.Status(500).JSON(fiber.Map{"error": "Error creating group option"})
			}
		}
	}
	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error committing transaction"})
	}

	var completeMenuItem models.MenuItem
	if err := db.DB.Preload("OptionGroups.Options").First(&completeMenuItem, jjjjjj.ID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading complete menu item"})
	}
	return c.JSON(completeMenuItem)
}

// @Summary เพิ่ม options แยกจากกลุ่มที่มีอยู่
// @Description สร้าง (options) ของเมนูนั้นๆ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param OptionGroupID query integer true "ID ของกลุ่มตัวเลือกที่ต้องการเพิ่ม option"
// @Param request body models.OptionRequest true "ข้อมูลเมนูและ options"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูและ options ที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 409 {object} map[string]interface{} "ชื่อเมนูซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างเมนูหรือ options"
// @Router /api/menu/options [post]
// @Tags menu
func AddMoreMenuOption(c *fiber.Ctx) error {
	id := c.Query("OptionGroupID")
	optgroupID, err := strconv.Atoi(id)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(map[string]interface{}{
			"error": "OptionGroupID is required",
		})
	}
	var req models.OptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
		})
	}

	var existingGroup models.OptionGroup
	if err := db.DB.First(&existingGroup, optgroupID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error": "Option Group not found",
		})
	}
	var existingOpt models.MenuOption
	if err := db.DB.Where("Name = ? AND group_id = ?", req.Name, optgroupID).First(&existingOpt).Error; err == nil {
		return c.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error": "Option ซ่ำชื่อนี้มีอยู่แล้ว",
		})
	}

	if err := db.DB.Unscoped().Where("Name = ? AND group_id = ?", req.Name, optgroupID).First(&existingOpt).Error; err == nil {
		db.DB.Unscoped().Model(&existingOpt).Update("deleted_at", nil)
		return c.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error": "Option ซ่ำชื่อกับ softdelete แต่ไม่ต้องห่วงเรากู้คืนให้แล้ว :)",
		})
	}

	groupOption := models.MenuOption{
		GroupID: existingGroup.ID,
		Name:    req.Name,
		NameEn:  req.NameEn,
		NameCh:  req.NameCh,
		Price:   req.Price,
	}

	if err := db.DB.Create(&groupOption).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order",
		})
	}

	if err := db.DB.Preload("Options").First(&existingGroup, optgroupID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "Error loading complete option group"})
	}
	return c.JSON(existingGroup)
}

// @Summary ลบเมนู/รายการอาหาร
// @Description ฟังก์ชันนี้ใช้สำหรับลบเมนูโดยจะเป็นการ soft delete งิงิ
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของอาหารนั้นๆ"
// @Success 200 "ลบเมนูสำเร็จ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Router /api/menu/{id} [delete]
// @Tags menu
func SoftDelete_Menu(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Menu ID is required",
		})
	}

	var existingMenu models.MenuItem
	if err := db.DB.First(&existingMenu, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error": "Menu not found",
		})
	}

	if err := db.DB.Delete(&existingMenu).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to delete menu",
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"Massage": "ลบสำเร็จ",
	})

}

// @Summary ลบตัวเลือกอาหาร
// @Description ฟังก์ชันนี้ใช้สำหรับลบตัวเลือกโดยจะเป็นการ soft delete
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของตัวเลือกนั้นๆ"
// @Success 200 "ลบตัตัวเลือกสำเร็จ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Router /api/menu/options/{id} [delete]
// @Tags menu
func SoftDelete_Option(c *fiber.Ctx) error {
	id := c.Params("id")

	if id == "" {
		return c.Status(http.StatusBadRequest).JSON(map[string]interface{}{
			"error": "Option ID is required",
		})
	}

	var existingMenuOption models.MenuOption
	if err := db.DB.First(&existingMenuOption, id).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(map[string]interface{}{
			"error": "Option not found",
		})
	}

	if err := db.DB.Delete(&existingMenuOption).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(map[string]interface{}{
			"error": "Failed to delete Option",
		})
	}

	return c.Status(http.StatusOK).JSON(map[string]interface{}{
		"Massage": "ลบสำเร็จ",
	})

}

// @Summary อัพเดทข้อมูลเมนูพื้นฐาน
// @Description อัพเดทข้อมูลพื้นฐานของเมนู เช่น ชื่อ, ราคา, คำอธิบาย
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของเมนู"
// @Param request body models.MenuItemRequest true "ข้อมูลเมนูที่ต้องการอัพเดท"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่อัพเดทแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 404 {object} map[string]interface{} "ไม่พบเมนูที่ต้องการอัพเดท"
// @Router /api/menu/{id} [put]
// @Tags menu
func UpdateMenuItem(c *fiber.Ctx) error {
	id := c.Params("id")
	menuID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu ID format",
		})
	}

	var req models.MenuItemRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input format",
		})
	}

	// Validation
	if req.Price < 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Price cannot be negative",
		})
	}

	// ตรวจสอบว่ามีเมนูอยู่จริง
	var existingMenu models.MenuItem
	if err := db.DB.First(&existingMenu, menuID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Menu not found",
		})
	}

	// ตรวจสอบว่ามีชื่อซ้ำกับเมนูอื่นหรือไม่ (ยกเว้นตัวเอง)
	var duplicateMenu models.MenuItem
	if err := db.DB.Where("name = ? AND id != ?", req.Name, menuID).First(&duplicateMenu).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "Menu name already exists",
		})
	}

	// อัพเดทข้อมูล
	updates := models.MenuItem{
		Name:          req.Name,
		NameEn:        req.NameEn,
		NameCh:        req.NameCh,
		Description:   req.Description,
		DescriptionEn: req.DescriptionEn,
		DescriptionCh: req.DescriptionCh,
		CategoryID:    req.CategoryID,
		Price:         req.Price,
	}

	if err := db.DB.Model(&existingMenu).Updates(updates).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update menu",
		})
	}

	// ดึงข้อมูลที่อัพเดทแล้วมาแสดง
	var updatedMenu models.MenuItem
	if err := db.DB.Preload("Category").Preload("OptionGroups.Options").First(&updatedMenu, menuID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading updated menu",
		})
	}

	return c.JSON(updatedMenu)
}

// @Summary อัพเดทข้อมูล Option Group
// @Description อัพเดทข้อมูลของ Option Group เช่น ชื่อ, จำนวนที่เลือกได้, การบังคับเลือก
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของ Option Group"
// @Param request body models.OptionGroupRequest true "ข้อมูล Option Group ที่ต้องการอัพเดท"
// @Success 200 {object} models.OptionGroup "รายละเอียดของ Option Group ที่อัพเดทแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 404 {object} map[string]interface{} "ไม่พบ Option Group ที่ต้องการอัพเดท"
// @Router /api/menu/option-groups/{id} [put]
// @Tags menu
func UpdateOptionGroup(c *fiber.Ctx) error {
	id := c.Params("id")
	groupID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid option group ID format",
		})
	}

	var req models.OptionGroupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input format",
		})
	}

	// Validation
	if req.MaxSelections <= 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "MaxSelections must be greater than 0",
		})
	}

	// ตรวจสอบว่ามี Option Group อยู่จริง
	var existingGroup models.OptionGroup
	if err := db.DB.First(&existingGroup, groupID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Option group not found",
		})
	}

	// ตรวจสอบว่าชื่อไม่ซ้ำกับ Option Group อื่นในเมนูเดียวกัน
	var duplicateGroup models.OptionGroup
	if err := db.DB.Where("name = ? AND menu_item_id = ? AND id != ?",
		req.Name, existingGroup.MenuItemID, groupID).First(&duplicateGroup).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "Option group name already exists in this menu",
		})
	}

	// อัพเดทข้อมูล
	updates := models.OptionGroup{
		Name:          req.Name,
		NameEn:        req.NameEn,
		NameCh:        req.NameCh,
		MaxSelections: req.MaxSelections,
		IsRequired:    req.IsRequired,
	}

	if err := db.DB.Model(&existingGroup).Updates(updates).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update option group",
		})
	}

	// ดึงข้อมูลที่อัพเดทแล้วมาแสดง
	var updatedGroup models.OptionGroup
	if err := db.DB.Preload("Options").First(&updatedGroup, groupID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading updated option group",
		})
	}

	return c.JSON(updatedGroup)
}

// @Summary อัพเดทข้อมูล Option
// @Description อัพเดทข้อมูลของ Option เช่น ชื่อ, ราคา
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของ Option"
// @Param request body models.OptionRequest true "ข้อมูล Option ที่ต้องการอัพเดท"
// @Success 200 {object} models.MenuOption "รายละเอียดของ Option ที่อัพเดทแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 404 {object} map[string]interface{} "ไม่พบ Option ที่ต้องการอัพเดท"
// @Router /api/menu/options/{id} [put]
// @Tags menu
func UpdateOption(c *fiber.Ctx) error {
	id := c.Params("id")
	optionID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid option ID format",
		})
	}

	var req models.OptionRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input format",
		})
	}

	// Validation
	if req.Price < 0 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Price cannot be negative",
		})
	}

	// ตรวจสอบว่ามี Option อยู่จริง
	var existingOption models.MenuOption
	if err := db.DB.First(&existingOption, optionID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Option not found",
		})
	}

	// ตรวจสอบว่าชื่อไม่ซ้ำกับ Option อื่นในกลุ่มเดียวกัน
	var duplicateOption models.MenuOption
	if err := db.DB.Where("name = ? AND group_id = ? AND id != ?",
		req.Name, existingOption.GroupID, optionID).First(&duplicateOption).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "Option name already exists in this group",
		})
	}

	// อัพเดทข้อมูล
	updates := models.MenuOption{
		Name:   req.Name,
		NameEn: req.NameEn,
		NameCh: req.NameCh,
		Price:  req.Price,
	}

	if err := db.DB.Model(&existingOption).Updates(updates).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update option",
		})
	}

	// ดึงข้อมูลที่อัพเดทแล้วมาแสดง
	var updatedOption models.MenuOption
	if err := db.DB.First(&updatedOption, optionID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading updated option",
		})
	}

	return c.JSON(updatedOption)
}

// @Summary อัพเดทรูปภาพเมนู
// @Description อัพเดทรูปภาพของเมนู
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของเมนู"
// @Param image formData file true "รูปภาพเมนู"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่อัพเดทรูปภาพแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 404 {object} map[string]interface{} "ไม่พบเมนูที่ต้องการอัพเดท"
// @Router /api/menu/image/{id} [put]
// @Tags menu
func UpdateMenuImage(c *fiber.Ctx) error {
	id := c.Params("id")
	menuID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu ID format",
		})
	}

	// ตรวจสอบว่ามีเมนูอยู่จริง
	var existingMenu models.MenuItem
	if err := db.DB.First(&existingMenu, menuID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Menu not found",
		})
	}

	// รับไฟล์รูปภาพ
	file, err := c.FormFile("image")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Error getting image file",
		})
	}

	// ตรวจสอบขนาดไฟล์ (เช่น จำกัดขนาดไม่เกิน 5MB)
	if file.Size > 5*1024*1024 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Image file size must be less than 5MB",
		})
	}

	// ตรวจสอบนามสกุลไฟล์
	filename := file.Filename
	ext := filepath.Ext(filename)
	validExts := map[string]bool{
		".jpg":  true,
		".jpeg": true,
		".png":  true,
	}
	if !validExts[strings.ToLower(ext)] {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Only JPG and PNG files are allowed",
		})
	}

	// เปิดไฟล์
	fileContent, err := file.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error opening image file",
		})
	}
	defer fileContent.Close()

	// อ่านข้อมูลไฟล์
	buffer := make([]byte, file.Size)
	if _, err := fileContent.Read(buffer); err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error reading image file",
		})
	}

	// อัพเดทรูปภาพในฐานข้อมูล
	if err := db.DB.Model(&existingMenu).Update("image", buffer).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update menu image",
		})
	}

	// ดึงข้อมูลที่อัพเดทแล้วมาแสดง
	var updatedMenu models.MenuItem
	if err := db.DB.Preload("Category").Preload("OptionGroups.Options").First(&updatedMenu, menuID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading updated menu",
		})
	}

	return c.JSON(updatedMenu)
}

// @Summary ดูรายการเมนูที่ถูกลบ
// @Description ดูรายการเมนูทั้งหมดที่ถูก soft delete
// @Produce json
// @Security BearerAuth
// @Success 200 {array} models.MenuItem "รายการเมนูที่ถูกลบ"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูล"
// @Router /api/menu/deleted [get]
// @Tags menu-restore
func GetDeletedMenus(c *fiber.Ctx) error {
	var deletedMenus []models.MenuItem

	if err := db.DB.Unscoped().Where("deleted_at IS NOT NULL").
		Preload("Category").
		Preload("OptionGroups", "deleted_at IS NOT NULL").
		Preload("OptionGroups.Options", "deleted_at IS NOT NULL").
		Find(&deletedMenus).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch deleted menus",
		})
	}

	return c.JSON(deletedMenus)
}

// @Summary กู้คืนเมนู
// @Description กู้คืนเมนูที่ถูก soft delete พร้อมกับ option groups และ options ที่เกี่ยวข้อง
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของเมนูที่ต้องการกู้คืน"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่กู้คืนแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 404 {object} map[string]interface{} "ไม่พบเมนูที่ต้องการกู้คืน"
// @Failure 409 {object} map[string]interface{} "มีเมนูที่ใช้ชื่อนี้อยู่แล้ว"
// @Router /api/menu/restore/{id} [post]
// @Tags menu-restore
func RestoreMenu(c *fiber.Ctx) error {
	id := c.Params("id")
	menuID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid menu ID format",
		})
	}

	// เริ่ม transaction
	tx := db.DB.Begin()

	// ตรวจสอบว่าเมนูที่จะกู้คืนมีอยู่จริง
	var deletedMenu models.MenuItem
	if err := tx.Unscoped().First(&deletedMenu, menuID).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Deleted menu not found",
		})
	}

	// ตรวจสอบว่าชื่อเมนูไม่ซ้ำกับเมนูที่มีอยู่
	var existingMenu models.MenuItem
	if err := tx.Where("name = ? AND deleted_at IS NULL", deletedMenu.Name).First(&existingMenu).Error; err == nil {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "A menu with this name already exists",
		})
	}

	// กู้คืนเมนู
	if err := tx.Unscoped().Model(&deletedMenu).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to restore menu",
		})
	}

	// กู้คืน option groups ที่เกี่ยวข้อง
	var deletedGroups []models.OptionGroup
	if err := tx.Unscoped().Where("menu_item_id = ? AND deleted_at IS NOT NULL", menuID).
		Find(&deletedGroups).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch deleted option groups",
		})
	}

	for _, group := range deletedGroups {
		// กู้คืน option group
		if err := tx.Unscoped().Model(&group).Update("deleted_at", nil).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to restore option group",
			})
		}

		// กู้คืน options ที่เกี่ยวข้อง
		if err := tx.Unscoped().Model(&models.MenuOption{}).
			Where("group_id = ? AND deleted_at IS NOT NULL", group.ID).
			Update("deleted_at", nil).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to restore options",
			})
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	// ดึงข้อมูลที่กู้คืนแล้วมาแสดง
	var restoredMenu models.MenuItem
	if err := db.DB.Preload("Category").
		Preload("OptionGroups.Options").
		First(&restoredMenu, menuID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading restored menu",
		})
	}

	return c.JSON(restoredMenu)
}

// @Summary กู้คืน Option Group
// @Description กู้คืน Option Group ที่ถูก soft delete พร้อมกับ options ที่เกี่ยวข้อง
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของ Option Group ที่ต้องการกู้คืน"
// @Success 200 {object} models.OptionGroup "รายละเอียดของ Option Group ที่กู้คืนแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 404 {object} map[string]interface{} "ไม่พบ Option Group ที่ต้องการกู้คืน"
// @Failure 409 {object} map[string]interface{} "มี Option Group ที่ใช้ชื่อนี้อยู่แล้วในเมนูเดียวกัน"
// @Router /api/menu/restore-group/{id} [post]
// @Tags menu-restore
func RestoreOptionGroup(c *fiber.Ctx) error {
	id := c.Params("id")
	groupID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid option group ID format",
		})
	}

	// เริ่ม transaction
	tx := db.DB.Begin()

	// ตรวจสอบว่า option group ที่จะกู้คืนมีอยู่จริง
	var deletedGroup models.OptionGroup
	if err := tx.Unscoped().First(&deletedGroup, groupID).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Deleted option group not found",
		})
	}

	// ตรวจสอบว่าชื่อไม่ซ้ำกับ option group อื่นในเมนูเดียวกัน
	var existingGroup models.OptionGroup
	if err := tx.Where("menu_item_id = ? AND name = ? AND deleted_at IS NULL",
		deletedGroup.MenuItemID, deletedGroup.Name).First(&existingGroup).Error; err == nil {
		tx.Rollback()
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "An option group with this name already exists in this menu",
		})
	}

	// กู้คืน option group
	if err := tx.Unscoped().Model(&deletedGroup).Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to restore option group",
		})
	}

	// กู้คืน options ที่เกี่ยวข้อง
	if err := tx.Unscoped().Model(&models.MenuOption{}).
		Where("group_id = ? AND deleted_at IS NOT NULL", groupID).
		Update("deleted_at", nil).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to restore options",
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	// ดึงข้อมูลที่กู้คืนแล้วมาแสดง
	var restoredGroup models.OptionGroup
	if err := db.DB.Preload("Options").First(&restoredGroup, groupID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading restored option group",
		})
	}

	return c.JSON(restoredGroup)
}

// @Summary กู้คืน Option
// @Description กู้คืน Option เดี่ยวที่ถูก soft delete
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path integer true "ID ของ Option ที่ต้องการกู้คืน"
// @Success 200 {object} models.MenuOption "รายละเอียดของ Option ที่กู้คืนแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 404 {object} map[string]interface{} "ไม่พบ Option ที่ต้องการกู้คืน"
// @Failure 409 {object} map[string]interface{} "มี Option ที่ใช้ชื่อนี้อยู่แล้วใน Option Group เดียวกัน"
// @Router /api/menu/restore-option/{id} [post]
// @Tags menu-restore
func RestoreOption(c *fiber.Ctx) error {
	id := c.Params("id")
	optionID, err := strconv.Atoi(id)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid option ID format",
		})
	}

	// ตรวจสอบว่า option ที่จะกู้คืนมีอยู่จริง
	var deletedOption models.MenuOption
	if err := db.DB.Unscoped().First(&deletedOption, optionID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Deleted option not found",
		})
	}

	// ตรวจสอบว่าชื่อไม่ซ้ำกับ option อื่นใน group เดียวกัน
	var existingOption models.MenuOption
	if err := db.DB.Where("group_id = ? AND name = ? AND deleted_at IS NULL",
		deletedOption.GroupID, deletedOption.Name).First(&existingOption).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "An option with this name already exists in this group",
		})
	}

	// กู้คืน option
	if err := db.DB.Unscoped().Model(&deletedOption).Update("deleted_at", nil).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to restore option",
		})
	}

	// ดึงข้อมูลที่กู้คืนแล้วมาแสดง
	var restoredOption models.MenuOption
	if err := db.DB.First(&restoredOption, optionID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error loading restored option",
		})
	}

	return c.JSON(restoredOption)
}