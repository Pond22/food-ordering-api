package api_handlers

import (
	"bytes"
	"encoding/json"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestCreateCategoryHandler(t *testing.T) {
	app := fiber.New()
	setupTestDB(t)

	// Register the route
	app.Post("/api/categories", CreateCategoryHandler)

	// กรณีทดสอบที่ 1: สร้างหมวดหมู่สำเร็จ
	t.Run("Success - Create category", func(t *testing.T) {
		reqBody := models.Category{
			Name: "Drinks",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.Category
		json.NewDecoder(resp.Body).Decode(&response)

		assert.Equal(t, reqBody.Name, response.Name)
		assert.NotZero(t, response.ID)
	})

	// กรณีทดสอบที่ 2: ชื่อหมวดหมู่ซ้ำ
	t.Run("Failure - Duplicate category name", func(t *testing.T) {
		// สร้างหมวดหมู่เดิมในฐานข้อมูล
		db.DB.Create(&models.Category{Name: "Food"})

		reqBody := models.Category{
			Name: "Food",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Contains(t, response["error"], "Category name already exists")
	})

	// กรณีทดสอบที่ 3: ข้อมูลไม่ถูกต้อง
	t.Run("Failure - Invalid input", func(t *testing.T) {
		invalidJSON := []byte(`{invalid json}`)
		req := httptest.NewRequest("POST", "/api/categories", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Contains(t, response["error"], "Invalid input")
	})

	// กรณีทดสอบที่ 4: ไม่ระบุชื่อหมวดหมู่
	t.Run("Failure - Missing category name", func(t *testing.T) {
		reqBody := models.Category{
			Name: "",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/categories", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		resp, err := app.Test(req)

		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Contains(t, response["error"], "Name  are required")
	})
}
func TestCreateMenuItemHandler(t *testing.T) {
	// กรณีทดสอบที่ 1: สร้างเมนูสำเร็จ
	t.Run("Success - Create menu with options", func(t *testing.T) {
		// Setup
		app := fiber.New()
		setupTestDB(t)

		// สร้างข้อมูลทดสอบ
		reqBody := models.CreateMenuRequest{
			MenuItem: models.MenuItemRequest{
				Name:          "Test Menu",
				NameEn:        "Test Menu EN",
				NameCh:        "Test Menu CH",
				Description:   "Test Description",
				DescriptionEn: "Test Description EN",
				DescriptionCh: "Test Description CH",
				CategoryID:    1,
				Price:         100,
			},
			OptionGroups: []models.OptionGroupRequest{
				{
					Name:          "Size",
					NameEn:        "Size EN",
					NameCh:        "Size CH",
					MaxSelections: 1,
					IsRequired:    true,
					Options: []models.OptionRequest{
						{
							Name:   "Small",
							NameEn: "Small EN",
							NameCh: "Small CH",
							Price:  0,
						},
						{
							Name:   "Large",
							NameEn: "Large EN",
							NameCh: "Large CH",
							Price:  20,
						},
					},
				},
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/menu", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Test
		app.Post("/api/menu", CreateMenuItemHandler)
		resp, err := app.Test(req)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.MenuItem
		json.NewDecoder(resp.Body).Decode(&response)

		assert.Equal(t, reqBody.MenuItem.Name, response.Name)
		assert.Equal(t, reqBody.MenuItem.Price, response.Price)
		assert.Equal(t, 1, len(response.OptionGroups))
		assert.Equal(t, 2, len(response.OptionGroups[0].Options))
	})

	// กรณีทดสอบที่ 2: ชื่อเมนูซ้ำ
	t.Run("Failure - Duplicate menu name", func(t *testing.T) {
		// Setup
		app := fiber.New()
		setupTestDB(t)

		// สร้างเมนูที่มีอยู่แล้ว
		existingMenu := models.MenuItem{
			Name:  "Existing Menu",
			Price: 100,
		}
		db.DB.Create(&existingMenu)

		// พยายามสร้างเมนูซ้ำ
		reqBody := models.CreateMenuRequest{
			MenuItem: models.MenuItemRequest{
				Name:  "Existing Menu",
				Price: 150,
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/menu", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Test
		app.Post("/api/menu", CreateMenuItemHandler)
		resp, err := app.Test(req)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, http.StatusConflict, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Contains(t, response["error"], "ชื่อเมนูนี้มีอยู่แล้ว")
	})

	// กรณีทดสอบที่ 3: ข้อมูลไม่ถูกต้อง
	t.Run("Failure - Invalid input", func(t *testing.T) {
		// Setup
		app := fiber.New()
		setupTestDB(t)

		// ส่งข้อมูล JSON ที่ไม่ถูกต้อง
		invalidJSON := []byte(`{invalid json}`)
		req := httptest.NewRequest("POST", "/api/menu", bytes.NewBuffer(invalidJSON))
		req.Header.Set("Content-Type", "application/json")

		// Test
		app.Post("/api/menu", CreateMenuItemHandler)
		resp, err := app.Test(req)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	// กรณีทดสอบที่ 4: สร้างเมนูโดยไม่มี option groups
	t.Run("Success - Create menu without options", func(t *testing.T) {
		// Setup
		app := fiber.New()
		setupTestDB(t)

		reqBody := models.CreateMenuRequest{
			MenuItem: models.MenuItemRequest{
				Name:  "Simple Menu",
				Price: 50,
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/menu", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Test
		app.Post("/api/menu", CreateMenuItemHandler)
		resp, err := app.Test(req)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.MenuItem
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, reqBody.MenuItem.Name, response.Name)
		assert.Equal(t, 0, len(response.OptionGroups))
	})

	// กรณีทดสอบที่ 5: สร้างเมนูที่มี option group แต่ไม่มี options
	t.Run("Success - Create menu with empty options group", func(t *testing.T) {
		// Setup
		app := fiber.New()
		setupTestDB(t)

		reqBody := models.CreateMenuRequest{
			MenuItem: models.MenuItemRequest{
				Name:  "Menu with Empty Group",
				Price: 150,
			},
			OptionGroups: []models.OptionGroupRequest{
				{
					Name:          "Empty Group",
					MaxSelections: 1,
					IsRequired:    false,
					Options:       []models.OptionRequest{},
				},
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/menu", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Test
		app.Post("/api/menu", CreateMenuItemHandler)
		resp, err := app.Test(req)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response models.MenuItem
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Equal(t, reqBody.MenuItem.Name, response.Name)
		assert.Equal(t, 1, len(response.OptionGroups))
		assert.Equal(t, 0, len(response.OptionGroups[0].Options))
	})

	// กรณีทดสอบที่ 6: สร้างเมนูที่ไม่มีชื่อ
	t.Run("Failure - Missing menu name", func(t *testing.T) {
		// Setup
		app := fiber.New()
		setupTestDB(t)

		reqBody := models.CreateMenuRequest{
			MenuItem: models.MenuItemRequest{
				Price: 50,
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/menu", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Test
		app.Post("/api/menu", CreateMenuItemHandler)
		resp, err := app.Test(req)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Contains(t, response["error"], "ชื่อเมนูจำเป็น")
	})

	// กรณีทดสอบที่ 7: สร้างเมนูด้วยราคาที่ไม่ถูกต้อง
	t.Run("Failure - Invalid price", func(t *testing.T) {
		// Setup
		app := fiber.New()
		setupTestDB(t)

		reqBody := models.CreateMenuRequest{
			MenuItem: models.MenuItemRequest{
				Name:  "Menu with Invalid Price",
				Price: -10,
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/menu", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Test
		app.Post("/api/menu", CreateMenuItemHandler)
		resp, err := app.Test(req)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Contains(t, response["error"], "ราคาต้องไม่ติดลบ")
	})

	// กรณีทดสอบที่ 8: ตัวเลือกใน Option Group เกินจำนวนที่กำหนด
	t.Run("Failure - Exceeding max selections", func(t *testing.T) {
		// Setup
		app := fiber.New()
		setupTestDB(t)

		reqBody := models.CreateMenuRequest{
			MenuItem: models.MenuItemRequest{
				Name:  "Menu with Invalid Options",
				Price: 100,
			},
			OptionGroups: []models.OptionGroupRequest{
				{
					Name:          "Size",
					MaxSelections: 1,
					IsRequired:    true,
					Options: []models.OptionRequest{
						{Name: "Option 1", Price: 10},
						{Name: "Option 2", Price: 20},
						{Name: "Option 3", Price: 30},
					},
				},
			},
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest("POST", "/api/menu", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")

		// Test
		app.Post("/api/menu", CreateMenuItemHandler)
		resp, err := app.Test(req)

		// Assert
		assert.Nil(t, err)
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

		var response map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&response)
		assert.Contains(t, response["error"], "จำนวนตัวเลือกเกินที่กำหนด")
	})
}

func setupTestDB(t *testing.T) {
	var err error
	db.DB, err = gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Migrate the schema
	err = db.DB.AutoMigrate(&models.MenuItem{}, &models.OptionGroup{}, &models.MenuOption{}, &models.Category{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	// Create a default category
	db.DB.Create(&models.Category{
		Name: "Default Category",
	})
}
