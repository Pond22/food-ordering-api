package api_handlers

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

// @Summary เรียกรายการเมนูทั้งหมด
// @Description ฟังก์ชันนี้ใช้สำหรับเรียกรายการเมนูทั้งหมดที่มีอยู่ในระบบ
// @Produce json
// @Success 200 {array} models.MenuItem "รายการเมนูทั้งหมด"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูลเมนู"
// @Router /menu [get]
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
// @Param ID query string true "ID ของเมนู"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูที่ค้นพบตาม ID"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจาก ID ที่ไม่ถูกต้อง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการดึงข้อมูลเมนู"
// @Router /getmenu [get]
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
// @Param action query string true "รูปแบบการค้นหาเมนู: getByID, getByCategory หรือ getAll" Enums(getByID, getByCategory, getAll)
// @Param id query integer false "ID ของเมนู (ใช้กับ action=getByID)"
// @Param category_id query integer false "ID ของหมวดหมู่ (ใช้กับ action=getByCategory)"
// @Success 200 {array} models.MenuItem "รายการเมนูที่ค้นพบ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากการระบุพารามิเตอร์"
// @Router /menu [get]
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
// @Param action query string true "รูปแบบการค้นหาเมนู: getByID, getByCategory หรือ getAll" Enums(getByID, getByCategory, getAll)
// @Param id query integer false "ID ของเมนู (ใช้กับ action=getByID)"
// @Param category_id query integer false "ID ของหมวดหมู่ (ใช้กับ action=getByCategory)"
// @Success 200 {array} models.MenuItem "รายการเมนูที่ค้นพบ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากการระบุพารามิเตอร์"
// @Router /getmenu [get]
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
// @Param request body models.CreateMenuRequest true "ข้อมูลเมนูและ options"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูและ options ที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 409 {object} map[string]interface{} "ชื่อเมนูซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างเมนูหรือ options"
// @Router /add_menu [post]
// @Tags menu
func CreateMenuItemHandler(c *fiber.Ctx) error {
	var req models.CreateMenuRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid input",
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
// @Param menu_id query integer true "ID ของเมนูที่ต้องการเพิ่ม group option"
// @Param request body models.OptionGroupRequest true "ข้อมูลเมนูและ group options"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูและ group options ที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 409 {object} map[string]interface{} "ชื่อเมนูซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างเมนูหรือ options"
// @Router /add_group_option [post]
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
// @Param OptionGroupID query integer true "ID ของกลุ่มตัวเลือกที่ต้องการเพิ่ม option"
// @Param request body models.OptionRequest true "ข้อมูลเมนูและ options"
// @Success 200 {object} models.MenuItem "รายละเอียดของเมนูและ options ที่สร้างเสร็จแล้ว"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้องหรือไม่ครบถ้วน"
// @Failure 409 {object} map[string]interface{} "ชื่อเมนูซ้ำกับที่มีอยู่แล้ว"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้างเมนูหรือ options"
// @Router /add_more_option [post]
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
// @Param id path integer true "ID ของอาหารนั้นๆ"
// @Success 200 "ลบเมนูสำเร็จ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Router /softDelete_Menu/{id} [delete]
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
// @Param id path integer true "ID ของตัวเลือกนั้นๆ"
// @Success 200 "ลบตัตัวเลือกสำเร็จ"
// @Failure 400 {object} map[string]interface{} "เกิดข้อผิดพลาดจากข้อมูลที่ไม่ถูกต้อง"
// @Router /softDelete_Option/{id} [delete]
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