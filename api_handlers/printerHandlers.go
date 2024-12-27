package api_handlers

import (
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"

	"github.com/gofiber/fiber/v2"
)

// Request structs
type CreateCategoryRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type UpdateCategoryRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type AssignCategoryRequest struct {
	CategoryIDs []uint `json:"category_ids" binding:"required"`
}

type PrinterResponse struct {
	ID         uint              `json:"id"`
	Name       string            `json:"name"`
	IPAddress  string            `json:"ip_address"`
	Port       int               `json:"port"`
	Department string            `json:"department"`
	Status     string            `json:"status"`
	Categories []models.Category `json:"categories"`
}

// // @Summary ดึงรายการหมวดหมู่ทั้งหมด
// // @Description ดึงรายการหมวดหมู่เครื่องพิมพ์ทั้งหมด
// // @Produce json
// // @Success 200 {array} models.Category
// // @Router /api/printers/printer-categories [get]
// // @Tags Printer
// func GetPrinterCategories(c *fiber.Ctx) error {
// 	var categories []models.Category
// 	if err := db.DB.Find(&categories).Error; err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch categories",
// 		})
// 	}
// 	return c.JSON(categories)
// }

// // @Summary อัพเดทข้อมูลหมวดหมู่
// // @Description อัพเดทข้อมูลของหมวดหมู่เครื่องพิมพ์
// // @Accept json
// // @Produce json
// // @Param id path int true "Category ID"
// // @Param category body UpdateCategoryRequest true "ข้อมูลที่ต้องการอัพเดท"
// // @Success 200 {object} models.Category
// // @Router /api/printers/printer-categories/{id} [put]
// // @Tags Printer
// func UpdatePrinterCategory(c *fiber.Ctx) error {
// 	id := c.Params("id")
// 	var req UpdateCategoryRequest
// 	if err := c.BodyParser(&req); err != nil {
// 		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
// 			"error": "Invalid request format",
// 		})
// 	}

// 	var category models.Category
// 	if err := db.DB.First(&category, id).Error; err != nil {
// 		return c.Status(http.StatusNotFound).JSON(fiber.Map{
// 			"error": "Category not found",
// 		})
// 	}

// 	updates := map[string]interface{}{}
// 	if req.Name != "" {
// 		updates["name"] = req.Name
// 	}
// 	if req.Description != "" {
// 		updates["description"] = req.Description
// 	}

// 	if err := db.DB.Model(&category).Updates(updates).Error; err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to update category",
// 		})
// 	}

// 	return c.JSON(category)
// }

// // @Summary ลบหมวดหมู่
// // @Description ลบหมวดหมู่เครื่องพิมพ์
// // @Produce json
// // @Param id path int true "Category ID"
// // @Success 200 {object} map[string]interface{}
// // @Router /api/printers/printer-categories/{id} [delete]
// // @Tags Printer
// func DeletePrinterCategory(c *fiber.Ctx) error {
// 	id := c.Params("id")

// 	// ลบความสัมพันธ์กับเครื่องพิมพ์ก่อน
// 	if err := db.DB.Exec("DELETE FROM printer_categories WHERE category_id = ?", id).Error; err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to remove printer associations",
// 		})
// 	}

// 	if err := db.DB.Delete(&models.Category{}, id).Error; err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to delete category",
// 		})
// 	}

// 	return c.JSON(fiber.Map{
// 		"message": "Category deleted successfully",
// 	})
// }

// @Summary กำหนดหมวดหมู่ให้เครื่องพิมพ์
// @Description กำหนดหมวดหมู่ที่เครื่องพิมพ์สามารถพิมพ์ได้
// @Accept json
// @Produce json
// @Param id path int true "Printer ID"
// @Param categories body AssignCategoryRequest true "รายการ ID ของหมวดหมู่"
// @Success 200 {object} models.Printer
// @Router /api/printers/categories/{id} [post]
// @Tags Printer
func AssignPrinterCategories(c *fiber.Ctx) error {
	printerID := c.Params("id")
	var req AssignCategoryRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	tx := db.DB.Begin()

	var printer models.Printer
	if err := tx.First(&printer, printerID).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Printer not found",
		})
	}

	var categories []models.Category
	if err := tx.Find(&categories, req.CategoryIDs).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid category IDs",
		})
	}

	if len(categories) != len(req.CategoryIDs) {
		tx.Rollback()
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Some categories not found",
		})
	}

	if err := tx.Model(&printer).Association("Categories").Replace(&categories); err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update printer categories",
		})
	}

	if err := tx.Preload("Categories").First(&printer, printerID).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch updated printer data",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	return c.JSON(printer)
}

// @Summary ดึงรายการหมวดหมู่ของเครื่องพิมพ์
// @Description ดึงรายการหมวดหมู่ที่เครื่องพิมพ์สามารถพิมพ์ได้
// @Produce json
// @Param id path int true "Printer ID"
// @Success 200 {array} models.Category
// @Router /api/printers/categories/{id} [get]
// @Tags Printer
func GetPrinterCategoriesById(c *fiber.Ctx) error {
	printerID := c.Params("id")

	var printer models.Printer
	if err := db.DB.Preload("Categories").First(&printer, printerID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Printer not found",
		})
	}

	return c.JSON(printer.Categories)
}

// @Summary ดึงข้อมูลเครื่องพิมพ์ทั้งหมด
// @Description ดึงรายการเครื่องพิมพ์ทั้งหมดพร้อมหมวดหมู่
// @Produce json
// @Success 200 {array} PrinterResponse
// @Router /api/printers [get]
// @Tags Printer
func GetAllPrinters(c *fiber.Ctx) error {
	var printers []models.Printer
	if err := db.DB.Preload("Categories").Find(&printers).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch printers",
		})
	}

	return c.JSON(printers)
}

// @Summary ดึงข้อมูลเครื่องพิมพ์ตาม ID
// @Description ดึงข้อมูลเครื่องพิมพ์และหมวดหมู่ตาม ID ที่ระบุ
// @Produce json
// @Param id path int true "Printer ID"
// @Success 200 {object} PrinterResponse
// @Router /api/printers/{id} [get]
// @Tags Printer
// func GetPrinterByID(c *fiber.Ctx) error {
// 	id := c.Params("id")

// 	var printer models.Printer
// 	if err := db.DB.Preload("Categories").First(&printer, id).Error; err != nil {
// 		return c.Status(http.StatusNotFound).JSON(fiber.Map{
// 			"error": "Printer not found",
// 		})
// 	}

// 	return c.JSON(printer)
// }

// printerHandlers.go

// @Summary รับงานพิมพ์ที่รอดำเนินการ
// @Description ดึงรายการงานพิมพ์ที่ยังไม่ได้พิมพ์สำหรับ IP ที่ระบุ
// @Produce json
// @Param printer_ip query string true "Printer IP Address"
// @Success 200 {array} models.PrintJob
// @Router /api/printers/pending-jobs [get]
// @Tags Printer
func GetPendingPrintJobs(c *fiber.Ctx) error {
	printerIP := c.Query("printer_ip")
	if printerIP == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Printer IP is required",
		})
	}
	s := "pending"
	var jobs []models.PrintJob
	if err := db.DB.Where("printer_ip = ? AND status = ?", printerIP, s).
		Order("created_at ASC").
		Find(&jobs).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch print jobs",
		})
	}

	return c.JSON(jobs)
}

// @Summary อัพเดทสถานะงานพิมพ์
// @Description อัพเดทสถานะของงานพิมพ์ (completed/failed)
// @Accept json
// @Produce json
// @Param id path int true "Print Job ID"
// @Param status body string true "New Status"
// @Success 200 {object} models.PrintJob
// @Router /api/printers/status/{id} [put]
// @Tags Printer
func UpdatePrintJobStatus(c *fiber.Ctx) error {
	jobID := c.Params("id")
	var req struct {
		Status string `json:"status"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if req.Status != "completed" && req.Status != "failed" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid status. Must be 'completed' or 'failed'",
		})
	}

	var job models.PrintJob
	if err := db.DB.Where("id = ?", jobID).First(&job).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Print job not found",
		})
	}

	job.Status = req.Status
	if err := db.DB.Save(&job).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update job status",
		})
	}

	return c.JSON(job)
}
