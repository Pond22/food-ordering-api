package api_handlers

import (
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
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

	// เริ่ม transaction
	tx := db.DB.Begin()

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

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	// ดึงข้อมูล Order ที่สมบูรณ์
	var completeOrder models.Order
	if err := db.DB.Preload("Items.MenuItem").
		Preload("Items.Options").
		First(&completeOrder, order.ID).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to load complete order",
		})
	}

	return c.JSON(completeOrder)
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

	// สร้าง Response ที่เรียบง่าย
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

// @Summary ดึงรายการออเดอร์ตามโต๊ะ
// @Description ดึงรายการออเดอร์ทั้งหมดของโต๊ะที่ระบุ
// @Produce json
// @Param table_id path int true "Table ID"
// @Success 200 {array} models.Order
// @Router /api/orders/table/{table_id} [get]
func GetTableOrders(c *fiber.Ctx) error {
	tableID := c.Params("table_id")

	var orders []models.Order
	if err := db.DB.Preload("Items").
		Preload("Items.MenuItem").
		Preload("Items.Options").
		Where("table_id = ?", tableID).
		Find(&orders).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch table orders",
		})
	}

	return c.JSON(orders)
}
