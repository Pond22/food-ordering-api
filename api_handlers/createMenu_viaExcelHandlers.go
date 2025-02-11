package api_handlers

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/xuri/excelize/v2"
)

type MenuImportRow struct {
	Name          string              `json:"name"`
	NameEn        string              `json:"name_en"`
	NameCh        string              `json:"name_ch"`
	Description   string              `json:"description"`
	DescriptionEn string              `json:"description_en"`
	DescriptionCh string              `json:"description_ch"`
	CategoryName  string              `json:"category_name"`
	Price         float64             `json:"price"`
	OptionGroups  []OptionGroupImport `json:"option_groups,omitempty"`
}

type OptionGroupImport struct {
	Name          string         `json:"name"`
	NameEn        string         `json:"name_en"`
	NameCh        string         `json:"name_ch"`
	MaxSelections int            `json:"max_selections"`
	IsRequired    bool           `json:"is_required"`
	Options       []OptionImport `json:"options"`
}

type OptionImport struct {
	Name   string  `json:"name"`
	NameEn string  `json:"name_en"`
	NameCh string  `json:"name_ch"`
	Price  float64 `json:"price"`
}

type MenuImportResponse struct {
	Success []MenuImportRow   `json:"success"`
	Errors  []MenuImportError `json:"errors"`
}

type MenuImportError struct {
	Row     int           `json:"row"`
	Message string        `json:"message"`
	Data    MenuImportRow `json:"data"`
}

type ImportError struct {
	Row        int    `json:"row"`
	Error      string `json:"error"`
	InputData  string `json:"input_data"`
	Suggestion string `json:"suggestion"`
}

type ImportResponse struct {
	Success      int             `json:"success"`
	Failed       int             `json:"failed"`
	FailedItems  []ImportError   `json:"failed_items"`
	SuccessItems []MenuImportRow `json:"success_items"`
}

// parseOptionString แปลงข้อความที่คั่นด้วย | เป็นโครงสร้างข้อมูลกลุ่มตัวเลือก
func parseOptionString(optionStr string) ([]OptionGroupImport, error) {
	var groups []OptionGroupImport
	// รูปแบบ: ชื่อกลุ่ม|ชื่อกลุ่มEN|ชื่อกลุ่มCH|จำนวนเลือก|บังคับ?|ตัวเลือก1:ราคา,ตัวเลือก2:ราคา
	// ตัวอย่าง: ความเผ็ด|Spiciness|辣度|1|true|ไม่เผ็ด:0,เผ็ดน้อย:0,เผ็ดมาก:0

	groupStrings := strings.Split(optionStr, "||")
	for _, groupStr := range groupStrings {
		if strings.TrimSpace(groupStr) == "" {
			continue
		}

		parts := strings.Split(groupStr, "|")
		if len(parts) < 6 {
			continue
		}

		maxSel, err := strconv.Atoi(strings.TrimSpace(parts[3]))
		if err != nil {
			return nil, err
		}

		isReq := strings.TrimSpace(strings.ToLower(parts[4])) == "true"

		group := OptionGroupImport{
			Name:          strings.TrimSpace(parts[0]),
			NameEn:        strings.TrimSpace(parts[1]),
			NameCh:        strings.TrimSpace(parts[2]),
			MaxSelections: maxSel,
			IsRequired:    isReq,
		}

		// แยกตัวเลือก
		optionParts := strings.Split(parts[5], ",")
		for _, optPart := range optionParts {
			if strings.TrimSpace(optPart) == "" {
				continue
			}

			// แยกชื่อและราคา
			optDetails := strings.Split(optPart, ":")
			if len(optDetails) != 4 {
				continue
			}

			price, err := strconv.ParseFloat(strings.TrimSpace(optDetails[3]), 64)
			if err != nil {
				return nil, err
			}

			option := OptionImport{
				Name:   strings.TrimSpace(optDetails[0]),
				NameEn: strings.TrimSpace(optDetails[1]),
				NameCh: strings.TrimSpace(optDetails[2]),
				Price:  price,
			}
			group.Options = append(group.Options, option)
		}

		groups = append(groups, group)
	}

	return groups, nil
}

// เพิ่มค่าคงที่สำหรับขนาดไฟล์สูงสุด
const maxFileSize = 5 * 1024 * 1024 // 5MB

// @Summary นำเข้าเมนูจากไฟล์ Excel
// @Description นำเข้าเมนูพร้อมกลุ่มตัวเลือกและตัวเลือกเสริมจากไฟล์ Excel (.xlsx, .xls)
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "ไฟล์ Excel สำหรับนำเข้าเมนู"
// @Success 200 {object} ImportResponse "ผลการนำเข้าเมนู"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการนำเข้า"
// @Router /api/menu/import [post]
// @Tags menu
func ImportMenuFromExcel(c *fiber.Ctx) error {
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "No file uploaded",
		})
	}

	// ตรวจสอบขนาดไฟล์
	if file.Size > maxFileSize {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "File size exceeds maximum limit of 5MB",
		})
	}

	// ตรวจสอบนามสกุลไฟล์
	fileName := strings.ToLower(file.Filename)
	isCSV := strings.HasSuffix(fileName, ".csv")
	isExcel := strings.HasSuffix(fileName, ".xlsx") || strings.HasSuffix(fileName, ".xls")

	if !isCSV && !isExcel {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid file format. Only .xlsx, .xls, and .csv files are allowed",
		})
	}

	// เปิดไฟล์
	src, err := file.Open()
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to open file",
		})
	}
	defer src.Close()

	var rows [][]string

	if isCSV {
		// อ่านไฟล์ CSV
		reader := csv.NewReader(src)
		rows, err = reader.ReadAll()
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse CSV file",
			})
		}
	} else {
		// อ่านไฟล์ Excel
		fileBytes := make([]byte, file.Size)
		if _, err := src.Read(fileBytes); err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read file",
			})
		}

		workbook, err := excelize.OpenReader(bytes.NewReader(fileBytes))
		if err != nil {
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Failed to parse Excel file",
			})
		}

		sheetName := workbook.GetSheetName(0)
		rows, err = workbook.GetRows(sheetName)
		if err != nil {
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to read sheet",
			})
		}
	}

	// ฟังก์ชันสำหรับสร้างข้อความแนะนำ
	getErrorSuggestion := func(errMsg string, inputData string) string {
		switch {
		case strings.Contains(errMsg, "Menu item already exists"):
			return "เมนูนี้มีอยู่ในระบบแล้ว กรุณาตรวจสอบชื่อเมนู"

		case strings.Contains(errMsg, "Invalid option groups format"):
			if strings.Contains(errMsg, "parsing") {
				return fmt.Sprintf("รูปแบบตัวเลือกไม่ถูกต้อง ต้องใช้รูปแบบ: ชื่อกลุ่ม|ชื่อEN|ชื่อCN|จำนวนที่เลือกได้|บังคับ?|ตัวเลือก1:ตัวเลือกEN:ตัวเลือกCN:ราคา\nข้อมูลที่รับมา: %s", inputData)
			}
			return "รูปแบบตัวเลือกไม่ถูกต้อง กรุณาตรวจสอบรูปแบบในเทมเพลต"

		default:
			return "กรุณาตรวจสอบข้อมูลให้ถูกต้องตามรูปแบบในเทมเพลต"
		}
	}

	failedItems := []ImportError{}
	successItems := []MenuImportRow{}
	successCount := 0

	// ในลูปการนำเข้าข้อมูล
	for i, row := range rows[1:] {
		// เริ่ม transaction ใหม่สำหรับแต่ละรายการ
		tx := db.DB.Begin()

		if len(row) < 9 {
			failedItems = append(failedItems, ImportError{
				Row:        i + 2,
				Error:      "Required fields missing or invalid",
				InputData:  row[8],
				Suggestion: getErrorSuggestion("Required fields missing or invalid", row[8]),
			})

			tx.Rollback()
			continue
		}

		menuRow := MenuImportRow{
			Name:          strings.TrimSpace(row[0]),
			NameEn:        strings.TrimSpace(row[1]),
			NameCh:        strings.TrimSpace(row[2]),
			Description:   strings.TrimSpace(row[3]),
			DescriptionEn: strings.TrimSpace(row[4]),
			DescriptionCh: strings.TrimSpace(row[5]),
			CategoryName:  strings.TrimSpace(row[6]),
		}

		// แปลงราคาเป็นตัวเลข
		price, err := strconv.ParseFloat(strings.TrimSpace(row[7]), 64)
		if err != nil {
			failedItems = append(failedItems, ImportError{
				Row:        i + 2,
				Error:      "Invalid price format",
				InputData:  row[7],
				Suggestion: getErrorSuggestion("Invalid price format", row[7]),
			})
			tx.Rollback()
			continue
		}
		menuRow.Price = price

		// แปลงข้อมูลกลุ่มตัวเลือก
		if row[8] != "" {
			optionGroups, err := parseOptionString(row[8])
			if err != nil {
				failedItems = append(failedItems, ImportError{
					Row:        i + 2,
					Error:      "Invalid option groups format: " + err.Error(),
					InputData:  row[8],
					Suggestion: getErrorSuggestion("Invalid option groups format", row[8]),
				})
				tx.Rollback()
				continue
			}
			menuRow.OptionGroups = optionGroups
		}

		// ตรวจสอบข้อมูลที่จำเป็น
		if menuRow.Name == "" || menuRow.CategoryName == "" || menuRow.Price <= 0 {
			failedItems = append(failedItems, ImportError{
				Row:        i + 2,
				Error:      "Required fields missing or invalid",
				InputData:  row[8],
				Suggestion: getErrorSuggestion("Required fields missing or invalid", row[8]),
			})
			tx.Rollback()
			continue
		}

		// ค้นหาหมวดหมู่ (จะไม่สร้างใหม่ถ้าไม่มีอยู่)
		var category models.Category
		if err := tx.Where("name = ?", menuRow.CategoryName).First(&category).Error; err != nil {
			failedItems = append(failedItems, ImportError{
				Row:        i + 2,
				Error:      "Category does not exist: " + menuRow.CategoryName,
				InputData:  menuRow.CategoryName,
				Suggestion: getErrorSuggestion("Category does not exist", menuRow.CategoryName),
			})
			tx.Rollback()
			continue // ข้ามเมนูนี้เพราะหมวดหมู่ไม่มีอยู่จริง
		}

		// แก้ไขส่วนการตรวจสอบเมนูซ้ำ
		var existingMenu models.MenuItem
		// ตรวจสอบเฉพาะเมนูที่ไม่ได้ถูก soft delete
		if err := db.DB.Where("name = ? AND deleted_at IS NULL", menuRow.Name).First(&existingMenu).Error; err == nil {
			failedItems = append(failedItems, ImportError{
				Row:        i + 2,
				Error:      "Menu item already exists",
				InputData:  menuRow.Name,
				Suggestion: getErrorSuggestion("Menu item already exists", menuRow.Name),
			})
			tx.Rollback()
			continue
		}

		// สร้างเมนูใหม่เสมอถ้าไม่มีชื่อซ้ำในระบบ (ไม่สนใจเมนูที่ถูก soft delete)
		menuItem := models.MenuItem{
			Name:          menuRow.Name,
			NameEn:        menuRow.NameEn,
			NameCh:        menuRow.NameCh,
			Description:   menuRow.Description,
			DescriptionEn: menuRow.DescriptionEn,
			DescriptionCh: menuRow.DescriptionCh,
			CategoryID:    category.ID,
			Price:         int16(menuRow.Price),
			Is_available:  true,
		}

		if err := tx.Create(&menuItem).Error; err != nil {
			failedItems = append(failedItems, ImportError{
				Row:        i + 2,
				Error:      "Failed to create menu item",
				InputData:  row[8],
				Suggestion: getErrorSuggestion("Failed to create menu item", row[8]),
			})
			tx.Rollback()
			continue
		}

		// สร้างกลุ่มตัวเลือกและตัวเลือก
		for _, group := range menuRow.OptionGroups {
			optionGroup := models.OptionGroup{
				MenuItemID:    menuItem.ID,
				Name:          group.Name,
				NameEn:        group.NameEn,
				NameCh:        group.NameCh,
				MaxSelections: group.MaxSelections,
				IsRequired:    group.IsRequired,
			}

			if err := tx.Create(&optionGroup).Error; err != nil {
				failedItems = append(failedItems, ImportError{
					Row:        i + 2,
					Error:      "Failed to create option group",
					InputData:  row[8],
					Suggestion: getErrorSuggestion("Failed to create option group", row[8]),
				})
				tx.Rollback()
				continue
			}

			for _, opt := range group.Options {
				option := models.MenuOption{
					GroupID: optionGroup.ID,
					Name:    opt.Name,
					NameEn:  opt.NameEn,
					NameCh:  opt.NameCh,
					Price:   opt.Price,
				}

				if err := tx.Create(&option).Error; err != nil {
					failedItems = append(failedItems, ImportError{
						Row:        i + 2,
						Error:      "Failed to create option",
						InputData:  row[8],
						Suggestion: getErrorSuggestion("Failed to create option", row[8]),
					})
					tx.Rollback()
					continue
				}
			}
		}

		// ถ้าสำเร็จให้ commit transaction ของรายการนี้
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			failedItems = append(failedItems, ImportError{
				Row:        i + 2,
				Error:      "Failed to commit transaction",
				InputData:  row[8],
				Suggestion: "เกิดข้อผิดพลาดในการบันทึกข้อมูล กรุณาลองใหม่อีกครั้ง",
			})
			continue
		}

		// เพิ่มรายการที่สำเร็จ
		successItems = append(successItems, menuRow)
		successCount++
	}

	response := ImportResponse{
		Success:      successCount,
		Failed:       len(failedItems),
		FailedItems:  failedItems,
		SuccessItems: successItems,
	}

	return c.JSON(response)
}
