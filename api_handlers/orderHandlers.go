package api_handlers

import (
	"bytes"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// สำหรับรับข้อมูลการสั่งอาหาร
type CreateOrderRequest struct {
	UUID     string             `json:"uuid" binding:"required"`     // UUID ของ QR Code
	TableID  uint               `json:"table_id" binding:"required"` // ID ของโต๊ะ
	Items    []orderItemRequest `json:"items" binding:"required"`    // รายการอาหารที่สั่ง
	UsePromo []UsePromoRequest  `json:"use_promo,omitempty"`         // โปรโมชั่นที่ใช้ (ถ้ามี)
}

type orderItemRequest struct {
	MenuItemID uint                     `json:"menu_item_id" binding:"required"`
	Quantity   int                      `json:"quantity" binding:"required,min=1"`
	Options    []OrderItemOptionRequest `json:"options,omitempty"`
	Notes      string                   `json:"notes,omitempty"`
}

type OrderItemOptionRequest struct {
	MenuOptionID uint `json:"menu_option_id" binding:"required"`
}

type UsePromoRequest struct {
	PromotionID uint   `json:"promotion_id" binding:"required"`
	MenuItemIDs []uint `json:"menu_item_ids" binding:"required"` // รายการอาหารที่ใช้โปรโมชั่น
}

type updateStatusRequest struct {
	Status string `json:"status" binding:"required,oneof=pending preparing ready served cancelled"`
}

// @Summary สร้างออเดอร์ใหม่
// @Description สร้างออเดอร์ใหม่พร้อมรายการอาหารและโปรโมชั่น (ถ้ามี)
// @Accept json
// @Produce json
// @Param order body CreateOrderRequest true "ข้อมูลออเดอร์"
// @Success 200 {object} models.Order
// @Router /api/orders [post]
// @Tags Order_ใหม่
func CreateOrder(c *fiber.Ctx) error {
	var req CreateOrderRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	tx := db.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 1. ตรวจสอบ QR Code และโต๊ะ
	var qrCode models.QRCode
	if err := tx.Where("uuid = ? AND table_id = ? AND is_active = ? AND expiry_at > ?",
		req.UUID, req.TableID, true, time.Now()).First(&qrCode).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusForbidden).JSON(fiber.Map{
			"error": "Invalid or expired QR code",
		})
	}

	// 2. สร้าง Order
	order := models.Order{
		UUID:    req.UUID,
		TableID: int(req.TableID),
		Status:  "pending",
		Total:   0, // จะคำนวณในภายหลัง
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order",
		})
	}

	// 3. สร้าง OrderItems
	var totalAmount float64 = 0
	for _, item := range req.Items {
		// ดึงข้อมูลเมนู
		var menuItem models.MenuItem
		if err := tx.First(&menuItem, item.MenuItemID).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": "Menu item not found",
			})
		}

		// สร้าง OrderItem
		orderItem := models.OrderItem{
			OrderID:    order.ID,
			MenuItemID: item.MenuItemID,
			Quantity:   item.Quantity,
			Price:      float64(menuItem.Price),
			Notes:      item.Notes,
			Status:     "pending",
		}

		if err := tx.Create(&orderItem).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to create order item",
			})
		}

		// เพิ่ม Options (ถ้ามี)
		for _, opt := range item.Options {
			var menuOption models.MenuOption
			if err := tx.First(&menuOption, opt.MenuOptionID).Error; err != nil {
				tx.Rollback()
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "Menu option not found",
				})
			}

			orderItemOption := models.OrderItemOption{
				OrderItemID:  orderItem.ID,
				MenuOptionID: opt.MenuOptionID,
				Price:        menuOption.Price,
				Quantity:     item.Quantity,
			}

			if err := tx.Create(&orderItemOption).Error; err != nil {
				tx.Rollback()
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to create order item option",
				})
			}

			totalAmount += menuOption.Price * float64(item.Quantity)
		}

		totalAmount += float64(menuItem.Price) * float64(item.Quantity)
	}

	// 4. จัดการโปรโมชั่น (ถ้ามี)
	if len(req.UsePromo) > 0 {
		for _, promo := range req.UsePromo {
			var promotion models.Promotion
			if err := tx.Preload("Items.MenuItem").First(&promotion, promo.PromotionID).Error; err != nil {
				tx.Rollback()
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "Promotion not found",
				})
			}

			// ตรวจสอบว่าโปรโมชั่นยังใช้งานได้
			now := time.Now()
			if !promotion.IsActive || now.Before(promotion.StartDate) || now.After(promotion.EndDate) {
				tx.Rollback()
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "Promotion is not active",
				})
			}

			// สร้าง OrderItem จากรายการในโปรโมชั่น
			for _, promoItem := range promotion.Items {
				orderItem := models.OrderItem{
					OrderID:    order.ID,
					MenuItemID: promoItem.MenuItemID,
					Quantity:   promoItem.Quantity,
					Price:      promotion.Price, // ใช้ราคาโปรโมชั่น
					Status:     "pending",
				}

				if err := tx.Create(&orderItem).Error; err != nil {
					tx.Rollback()
					return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to create order item",
					})
				}
			}

			totalAmount += promotion.Price

			promoUsage := models.PromotionUsage{
				PromotionID: promotion.ID,
				OrderID:     order.ID,
			}

			if err := tx.Create(&promoUsage).Error; err != nil {
				tx.Rollback()
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to record promotion usage",
				})
			}
		}
	}

	// 5. อัพเดทยอดรวมของ Order
	if err := tx.Model(&order).Update("total", totalAmount).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order total",
		})
	}

	// 6. ดึงข้อมูล Order ที่สมบูรณ์
	var completeOrder models.Order
	if err := tx.Preload("Items.MenuItem.Category").
		Preload("Items.Options.MenuOption").
		First(&completeOrder, order.ID).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load complete order",
		})
	}

	// 7. จัดกลุ่มรายการอาหารตาม category ID
	categoryItems := make(map[uint][]models.OrderItem)
	for _, item := range completeOrder.Items {
		categoryID := item.MenuItem.CategoryID
		categoryItems[categoryID] = append(categoryItems[categoryID], item)
	}

	// 8. สร้าง print jobs สำหรับแต่ละหมวดหมู่
	for categoryID, items := range categoryItems {
		// ดึงเครื่องพิมพ์ทั้งหมดที่ถูกกำหนดให้พิมพ์หมวดหมู่นี้โดยตรง
		var categoryPrinters []models.Printer
		if err := tx.Table("printers").
			Joins("JOIN printer_categories ON printers.id = printer_categories.printer_id").
			Where("printer_categories.category_id = ?", categoryID).
			Find(&categoryPrinters).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch printers for category",
			})
		}

		// ถ้าไม่มีเครื่องพิมพ์สำหรับหมวดหมู่นี้ ใช้เครื่องพิมพ์หลัก
		if len(categoryPrinters) == 0 {
			var mainPrinter models.Printer
			if err := tx.Where("name = ?", "main").First(&mainPrinter).Error; err != nil {
				tx.Rollback()
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "No printer available for category ID: " + fmt.Sprint(categoryID),
				})
			}
			categoryPrinters = []models.Printer{mainPrinter}
		}

		// ดึงข้อมูลหมวดหมู่เพื่อใช้ในการสร้างเนื้อหา
		var category models.Category
		if err := tx.First(&category, categoryID).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch category info",
			})
		}

		// สร้างเนื้อหาที่จะพิมพ์
		content := createOrderPrintContent(completeOrder, category.Name, items)

		// สร้าง print job สำหรับทุกเครื่องพิมพ์ที่กำหนดให้พิมพ์หมวดหมู่นี้
		for _, printer := range categoryPrinters {
			printJob := models.PrintJob{
				PrinterID: printer.ID,
				OrderID:   &completeOrder.ID,
				Content:   content,
				Status:    "pending",
				JobType:   "order",
			}

			if err := tx.Create(&printJob).Error; err != nil {
				tx.Rollback()
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("Failed to create print job for printer %s", printer.Name),
				})
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	return c.JSON(completeOrder)
}

func createOrderPrintContent(order models.Order, categoryName string, items []models.OrderItem) []byte {
	var buf bytes.Buffer

	// 1. Initialize printer
	buf.Write([]byte{0x1B, 0x40}) // ESC @ - Initialize printer

	// 2. Select Thai Code Page (WPC1255)
	buf.Write([]byte{0x1B, 0x74, 0x49}) // OEM864

	// 3. Set Character Size - using standard size as shown in manual
	buf.Write([]byte{0x1D, 0x21, 0x00}) // GS ! n - Character size x1

	// 4. Set print density (optional - for clearer text)
	buf.Write([]byte{0x1B, 0x7C, 0x04}) // ESC | n - Set print density

	// Header
	buf.Write([]byte{0x1B, 0x61, 0x01}) // Center align
	buf.WriteString(fmt.Sprintf("Order #%d\n", order.ID))
	buf.Write([]byte{0x1B, 0x61, 0x00}) // Left align

	buf.WriteString(fmt.Sprintf("Table: %d\n", order.TableID))
	buf.WriteString(fmt.Sprintf("Category: %s\n", categoryName))
	buf.WriteString("-------------------------\n")

	// Menu items
	for _, item := range items {
		// Bold for item name
		buf.Write([]byte{0x1B, 0x45, 0x01}) // Bold on
		buf.WriteString(fmt.Sprintf("%s", item.MenuItem.Name))
		buf.Write([]byte{0x1B, 0x45, 0x00}) // Bold off
		buf.WriteString(fmt.Sprintf(" x%d\n", item.Quantity))

		// Options
		for _, opt := range item.Options {
			buf.WriteString(fmt.Sprintf("  + %s\n", opt.MenuOption.Name))
		}

		// Notes
		if item.Notes != "" {
			buf.WriteString(fmt.Sprintf("  Note: %s\n", item.Notes))
		}
		buf.WriteString("\n")
	}

	// Footer
	buf.WriteString("-------------------------\n")
	buf.WriteString(fmt.Sprintf("Printed: %s\n", time.Now().Format("15:04:05")))

	// Cut paper
	buf.Write([]byte{0x1D, 0x56, 0x41, 0x03}) // GS V A 3 - Full cut

	return buf.Bytes()
}

// @Summary อัพเดทสถานะออเดอร์
// @Description อัพเดทสถานะของออเดอร์ (pending, preparing, ready, served, cancelled)
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Param status body updateStatusRequest true "สถานะใหม่"
// @Success 200 {object} models.Order
// @Router /api/orders/status/{id} [put]
// @Tags Order_ใหม่
func UpdateOrderStatus(c *fiber.Ctx) error {
	orderID := c.Params("id")
	var req updateStatusRequest

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	tx := db.DB.Begin()

	// ดึงข้อมูล Order
	var order models.Order
	if err := tx.First(&order, orderID).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// อัพเดทสถานะ
	if err := tx.Model(&order).Update("status", req.Status).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order status",
		})
	}

	// อัพเดทสถานะของทุก items ในออเดอร์
	if err := tx.Model(&models.OrderItem{}).
		Where("order_id = ?", orderID).
		Update("status", req.Status).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order items status",
		})
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	return c.JSON(order)
}

// @Summary ยืนยันการเสิร์ฟอาหาร
// @Description ยืนยันการเสิร์ฟอาหารสำหรับรายการอาหารในออเดอร์
// @Accept json
// @Produce json
// @Param id path int true "OrderItem ID"
// @Success 200 {object} models.OrderItem
// @Router /api/orders/items/serve/{id} [post]
// @Tags Order_ใหม่
func ServeOrderItem(c *fiber.Ctx) error {
	itemID := c.Params("id")

	tx := db.DB.Begin()

	var orderItem models.OrderItem
	if err := tx.First(&orderItem, itemID).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Order item not found",
		})
	}

	now := time.Now()
	updates := map[string]interface{}{
		"status":    "served",
		"served_at": now,
	}

	if err := tx.Model(&orderItem).Updates(updates).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order item",
		})
	}

	// ตรวจสอบว่าทุกรายการในออเดอร์ถูกเสิร์ฟแล้วหรือไม่
	var unservedCount int64
	if err := tx.Model(&models.OrderItem{}).
		Where("order_id = ? AND status != ?", orderItem.OrderID, "served").
		Count(&unservedCount).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check order items status",
		})
	}

	// ถ้าทุกรายการถูกเสิร์ฟแล้ว อัพเดทสถานะของออเดอร์เป็น served
	if unservedCount == 0 {
		if err := tx.Model(&models.Order{}).
			Where("id = ?", orderItem.OrderID).
			Update("status", "served").Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update order status",
			})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	return c.JSON(orderItem)
}

// @Summary ยกเลิกออเดอร์
// @Description ยกเลิกออเดอร์และรายการอาหารทั้งหมดในออเดอร์
// @Accept json
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} models.Order
// @Router /api/orders/{id}/cancel [post]
func CancelOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")

	tx := db.DB.Begin()

	// ดึงข้อมูล Order พร้อม Items และ PromotionUsage
	var order models.Order
	if err := tx.Preload("Items").
		Preload("Items.Options").
		First(&order, orderID).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// ตรวจสอบว่าสามารถยกเลิกได้หรือไม่
	if order.Status == "served" || order.Status == "cancelled" {
		tx.Rollback()
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Cannot cancel order that has been served or already cancelled",
		})
	}

	// 1. ยกเลิกการใช้โปรโมชั่น (soft delete)
	if err := tx.Delete(&models.PromotionUsage{}, "order_id = ?", order.ID).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to cancel promotion usage",
		})
	}

	// 2. อัพเดทสถานะของทุก OrderItems เป็น cancelled
	if err := tx.Model(&models.OrderItem{}).
		Where("order_id = ?", order.ID).
		Updates(map[string]interface{}{
			"status": "cancelled",
		}).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to cancel order items",
		})
	}

	// 3. อัพเดทสถานะของ Order เป็น cancelled
	if err := tx.Model(&order).Updates(map[string]interface{}{
		"status": "cancelled",
	}).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to cancel order",
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	// ดึงข้อมูล Order ที่อัพเดทแล้ว
	var updatedOrder models.Order
	if err := db.DB.Preload("Items").
		Preload("Items.Options").
		First(&updatedOrder, order.ID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load updated order",
		})
	}

	return c.JSON(updatedOrder)
}

type OrderResponse struct {
	ID      uint                `json:"id"`
	UUID    string              `json:"uuid"`
	TableID int                 `json:"table_id"`
	Status  string              `json:"status"`
	Total   float64             `json:"total"`
	Items   []OrderItemResponse `json:"items"`
}

type OrderItemResponse struct {
	ID         uint                      `json:"id"`
	MenuItemID uint                      `json:"menu_item_id"`
	MenuItem   MenuItemResponseMin       `json:"menu_item"`
	Quantity   int                       `json:"quantity"`
	Price      float64                   `json:"price"`
	Notes      string                    `json:"notes"`
	Status     string                    `json:"status"`
	Options    []OrderItemOptionResponse `json:"options,omitempty"`
}

type MenuItemResponseMin struct {
	ID    uint   `json:"id"`
	Name  string `json:"name"`
	Price int16  `json:"price"`
}

type OrderItemOptionResponse struct {
	ID    uint    `json:"id"`
	Name  string  `json:"name"`
	Price float64 `json:"price"`
}

// @Summary ดึงรายการออเดอร์ที่กำลังดำเนินการ
// @Description ดึงรายการออเดอร์ที่ยังไม่เสร็จสิ้น (pending, preparing, ready)
// @Produce json
// @Success 200 {array} OrderResponse
// @Router /api/orders/active [get]
// @Tags Order_ใหม่
func GetActiveOrders(c *fiber.Ctx) error {
	var orders []models.Order

	if err := db.DB.Preload("Items").
		Preload("Items.MenuItem").
		Preload("Items.Options").
		Where("status IN ?", []string{"pending", "preparing", "ready"}).
		Find(&orders).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch active orders",
		})
	}

	var response []OrderResponse
	for _, order := range orders {
		orderResp := OrderResponse{
			ID:      order.ID,
			UUID:    order.UUID,
			TableID: order.TableID,
			Status:  order.Status,
			Total:   order.Total,
		}

		for _, item := range order.Items {
			orderItem := OrderItemResponse{
				ID:         item.ID,
				MenuItemID: item.MenuItemID,
				MenuItem: MenuItemResponseMin{
					ID:    item.MenuItem.ID,
					Name:  item.MenuItem.Name,
					Price: item.MenuItem.Price,
				},
				Quantity: item.Quantity,
				Price:    item.Price,
				Notes:    item.Notes,
				Status:   item.Status,
			}

			// เพิ่ม Options ถ้ามี
			for _, opt := range item.Options {
				orderItem.Options = append(orderItem.Options, OrderItemOptionResponse{
					ID:    opt.ID,
					Name:  opt.MenuOption.Name,
					Price: opt.Price,
				})
			}

			orderResp.Items = append(orderResp.Items, orderItem)
		}

		response = append(response, orderResp)
	}

	return c.JSON(response)
}

// @Summary ดึงรายละเอียดออเดอร์
// @Description ดึงรายละเอียดของออเดอร์ตาม ID
// @Produce json
// @Param id path int true "Order ID"
// @Success 200 {object} models.Order
// @Router /api/orders/{id} [get]
// @Tags Order_ใหม่
func GetOrder(c *fiber.Ctx) error {
	orderID := c.Params("id")

	var order models.Order
	if err := db.DB.Preload("Items").
		Preload("Items.MenuItem").
		Preload("Items.Options").
		First(&order, orderID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	return c.JSON(order)
}

type cancle_item_req struct {
	OrderUUID string `json:"order_uuid" binding:"required"`
	Table_id  uint   `json:"id"`
	Items     []struct {
		OrderItemID uint `json:"order_item_id" binding:"required"`
		Quantity    int  `json:"quantity" binding:"required,min=1"`
	} `json:"items" binding:"required"`
}

// @Summary ยกเลิกรายการอาหาร
// @Description ยกเลิกรายการอาหารในออเดอร์ตามจำนวนที่กำหนด
// @Accept json
// @Produce json
// @Param order body cancle_item_req true "ข้อมูลยกเลิก"
// @Success 200 {object} models.OrderItem
// @Router /api/orders/items/cancel [post]
// @Tags Order_ใหม่
func CancelOrderItem(c *fiber.Ctx) error {
	var req cancle_item_req

	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if req.OrderUUID == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Order UUID is required",
		})
	}

	tx := db.DB.Begin()

	var order models.Order

	if err := tx.Where("uuid = ? AND table_id = ?", req.OrderUUID, req.Table_id).Preload("Items.MenuItem").Preload("Items.Options").First(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// รายการสำหรับ PrintJob
	var printContents []string

	for _, reqItem := range req.Items {
		var orderItem models.OrderItem
		if err := tx.Where("id = ?", reqItem.OrderItemID).Preload("MenuItem").Preload("Options").First(&orderItem).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("Order item ID %d not found", reqItem.OrderItemID),
			})
		}

		if reqItem.Quantity > orderItem.Quantity {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Cancel quantity exceeds ordered quantity for item ID %d", reqItem.OrderItemID),
			})
		}

		if orderItem.Quantity == reqItem.Quantity {
			orderItem.Status = "cancelled"
			orderItem.Quantity = 0
			if err := tx.Where("order_item_id = ?", orderItem.ID).
				Delete(&models.OrderItemOption{}).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to delete item options",
				})
			}
		} else {
			orderItem.Quantity -= reqItem.Quantity
			for i := range orderItem.Options {
				orderItem.Options[i].Quantity -= reqItem.Quantity
				if err := tx.Save(&orderItem.Options[i]).Error; err != nil {
					tx.Rollback()
					return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
						"error": "Failed to update order item option quantity",
					})
				}
			}
		}

		if err := tx.Save(&orderItem).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update order item status",
			})
		}

		printContents = append(printContents, fmt.Sprintf("%s|%d", orderItem.MenuItem.Name, reqItem.Quantity))
	}

	// อัปเดตยอดรวมออเดอร์หลังจากยกเลิกรายการ
	if err := updateOrdersTotalByUUIDAndTableID(tx, req.OrderUUID, req.Table_id); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order total",
		})
	}

	// หาเครื่องพิมพ์ที่เกี่ยวข้องตามหมวดหมู่ของเมนูอาหารในออเดอร์
	var printers []models.Printer
	err := db.DB.Distinct("printers.*").
		Joins("JOIN printer_categories ON printers.id = printer_categories.printer_id").
		Joins("JOIN order_items ON order_items.order_id = ?", order.ID).
		Joins("JOIN menu_items ON menu_items.id = order_items.menu_item_id").
		Where("printer_categories.category_id = menu_items.category_id").
		Find(&printers).Error

	if err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": fmt.Sprintf("Failed to find printers: %v", err),
		})
	}

	// If no category-specific printers found, fall back to main printer
	if len(printers) == 0 {
		var mainPrinter models.Printer
		if err := tx.Where("name = ?", "main").First(&mainPrinter).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "No printers available and could not find main printer",
			})
		}
		printers = append(printers, mainPrinter)
	}

	printContentString := strings.Join(printContents, "\n")
	for _, printer := range printers {
		printJob := models.PrintJob{
			PrinterID:         printer.ID,
			OrderID:           &order.ID,
			Status:            "pending",
			JobType:           "cancelation",
			CancelledQuantity: len(req.Items),
			Content:           []byte(printContentString),
		}

		if err := tx.Create(&printJob).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": fmt.Sprintf("Failed to create print job for printer %s", printer.Name),
			})
		}
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	return c.JSON(fiber.Map{"message": "Order items cancelled successfully"})
}

func updateOrdersTotalByUUIDAndTableID(tx *gorm.DB, uuid string, tableID uint) error {
	// ดึงข้อมูล Orders ทั้งหมดที่ตรงกับ UUID และ Table ID
	var orders []models.Order
	err := tx.Where("uuid = ? AND table_id = ?", uuid, tableID).
		Preload("Items.MenuItem").
		Preload("Items.Options").
		Find(&orders).Error
	if err != nil {
		return fmt.Errorf("failed to find orders: %w", err)
	}

	// ถ้าไม่เจอ orders ใด ๆ
	if len(orders) == 0 {
		return fmt.Errorf("no orders found for uuid %s and table_id %d", uuid, tableID)
	}

	// คำนวณยอดรวมใหม่สำหรับแต่ละ order
	for _, order := range orders {
		var total float64

		// คำนวณราคารวมของ Items (ไม่รวมสถานะ 'cancelled')
		err := tx.Model(&models.OrderItem{}).
			Select("COALESCE(SUM(price * quantity), 0)").
			Where("order_id = ? AND status != ?", order.ID, "cancelled").
			Scan(&total).Error
		if err != nil {
			return fmt.Errorf("failed to calculate total for order ID %d: %w", order.ID, err)
		}

		// คำนวณราคารวมของ Options (ไม่รวมของที่ถูกลบและสถานะ 'cancelled')
		var optionsTotal float64
		err = tx.Model(&models.OrderItemOption{}).
			Joins("JOIN order_items ON order_item_options.order_item_id = order_items.id").
			Select("COALESCE(SUM(order_item_options.price * order_item_options.quantity), 0)").
			Where("order_items.order_id = ? AND order_items.status != ? AND order_item_options.deleted_at IS NULL",
				order.ID, "cancelled").
			Scan(&optionsTotal).Error
		if err != nil {
			return fmt.Errorf("failed to calculate options total for order ID %d: %w", order.ID, err)
		}

		// รวมยอดทั้งหมด
		finalTotal := total + optionsTotal
		fmt.Printf("Updating order ID %d with total: %.2f\n", order.ID, finalTotal)

		// อัปเดตยอดรวมในตาราง orders
		err = tx.Model(&models.Order{}).
			Where("id = ?", order.ID).
			Update("total", finalTotal).Error
		if err != nil {
			return fmt.Errorf("failed to update total for order ID %d: %w", order.ID, err)
		}
	}

	return nil
}

// // @Summary ดึงรายการออเดอร์ตามโต๊ะ
// // @Description ดึงรายการออเดอร์ทั้งหมดของโต๊ะที่ระบุ
// // @Produce json
// // @Param table_id path int true "Table ID"
// // @Success 200 {array} models.Order
// // @Router /api/orders/table/{table_id} [get]
// func GetTableOrders(c *fiber.Ctx) error {
// 	tableID := c.Params("table_id")

// 	var orders []models.Order
// 	if err := db.DB.Preload("Items").
// 		Preload("Items.MenuItem").
// 		Preload("Items.Options").
// 		Where("table_id = ?", tableID).
// 		Find(&orders).Error; err != nil {
// 		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
// 			"error": "Failed to fetch table orders",
// 		})
// 	}

// 	return c.JSON(orders)
// }