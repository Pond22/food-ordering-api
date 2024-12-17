package api_handlers

import (
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"

	"time"

	"github.com/gofiber/fiber/v2"
)

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