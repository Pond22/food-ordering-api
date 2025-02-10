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
		Total:   0,
	}

	if err := tx.Create(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create order",
		})
	}

	// 3. จัดการรายการสั่งอาหารปกติ
	var totalAmount float64 = 0
	for _, item := range req.Items {
		var menuItem models.MenuItem
		if err := tx.First(&menuItem, item.MenuItemID).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("Menu item ID %d not found", item.MenuItemID),
			})
		}

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

		// บันทึก Options
		for _, opt := range item.Options {
			var menuOption models.MenuOption
			if err := tx.First(&menuOption, opt.MenuOptionID).Error; err != nil {
				tx.Rollback()
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": fmt.Sprintf("Menu option ID %d not found", opt.MenuOptionID),
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

	// 4. จัดการโปรโมชั่น
	if len(req.UsePromo) > 0 {
		for _, promoReq := range req.UsePromo {
			var promotion models.Promotion
			if err := tx.Preload("Items.MenuItem").First(&promotion, promoReq.PromotionID).Error; err != nil {
				tx.Rollback()
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "Promotion not found",
				})
			}

			// ตรวจสอบการใช้งานโปรโมชั่น
			now := time.Now()
			if !promotion.IsActive || now.Before(promotion.StartDate) || now.After(promotion.EndDate) {
				tx.Rollback()
				return c.Status(http.StatusBadRequest).JSON(fiber.Map{
					"error": "Promotion is not active",
				})
			}

			// สร้าง PromotionUsage
			promoUsage := models.PromotionUsage{
				PromotionID: promotion.ID,
				OrderID:     order.ID,
				SaveAmount:  0, // จะคำนวณภายหลัง
			}

			if err := tx.Create(&promoUsage).Error; err != nil {
				tx.Rollback()
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to record promotion usage",
				})
			}

			// กรณีโปรโมชั่นแบบธรรมดา (fixed set)
			if promotion.MaxSelections == 0 && promotion.MinSelections == 0 {
				// ไม่ต้องส่ง MenuItemIDs
				if len(promoReq.MenuItemIDs) > 0 {
					tx.Rollback()
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{
						"error": "This promotion does not require item selection",
					})
				}

				// สร้าง OrderItem จากรายการในโปรโมชั่นทั้งหมด
				for _, promoItem := range promotion.Items {
					orderItem := models.OrderItem{
						OrderID:          order.ID,
						MenuItemID:       promoItem.MenuItemID,
						Quantity:         promoItem.Quantity,
						Price:            promotion.Price / float64(len(promotion.Items)),
						Status:           "pending",
						PromotionUsageID: &promoUsage.ID,
					}

					if err := tx.Create(&orderItem).Error; err != nil {
						tx.Rollback()
						return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
							"error": "Failed to create promotion order item",
						})
					}
				}

				totalAmount += promotion.Price
			} else {
				// กรณีโปรโมชั่นแบบเลือกได้
				if len(promoReq.MenuItemIDs) < promotion.MinSelections ||
					len(promoReq.MenuItemIDs) > promotion.MaxSelections {
					tx.Rollback()
					return c.Status(http.StatusBadRequest).JSON(fiber.Map{
						"error": fmt.Sprintf("Invalid number of selections. Must select between %d and %d items",
							promotion.MinSelections, promotion.MaxSelections),
					})
				}

				// ตรวจสอบว่าเมนูที่เลือกอยู่ในโปรโมชั่น
				for _, menuItemID := range promoReq.MenuItemIDs {
					found := false
					for _, promoItem := range promotion.Items {
						if promoItem.MenuItemID == menuItemID {
							found = true
							break
						}
					}

					if !found {
						tx.Rollback()
						return c.Status(http.StatusBadRequest).JSON(fiber.Map{
							"error": fmt.Sprintf("Menu item ID %d is not in promotion", menuItemID),
						})
					}
				}

				// คำนวณราคาต่อรายการ
				pricePerItem := promotion.Price / float64(len(promoReq.MenuItemIDs))

				// สร้าง OrderItem สำหรับรายการที่เลือก
				for _, menuItemID := range promoReq.MenuItemIDs {
					orderItem := models.OrderItem{
						OrderID:          order.ID,
						MenuItemID:       menuItemID,
						Quantity:         1,
						Price:            pricePerItem,
						Status:           "pending",
						PromotionUsageID: &promoUsage.ID,
					}

					if err := tx.Create(&orderItem).Error; err != nil {
						tx.Rollback()
						return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
							"error": "Failed to create promotion order item",
						})
					}
				}

				totalAmount += promotion.Price
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

	// 8. สร้าง print jobs โดยจัดกลุ่มตามเครื่องพิมพ์
	printerJobs := make(map[uint][]struct {
		CategoryID uint
		Items      []models.OrderItem
	})

	// จัดกลุ่มรายการตามเครื่องพิมพ์
	for categoryID, items := range categoryItems {
		// ดึงเครื่องพิมพ์ที่รองรับหมวดหมู่นี้
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

		// เพิ่มรายการลงในกลุ่มของแต่ละเครื่องพิมพ์
		for _, printer := range categoryPrinters {
			printerJobs[printer.ID] = append(printerJobs[printer.ID], struct {
				CategoryID uint
				Items      []models.OrderItem
			}{
				CategoryID: categoryID,
				Items:      items,
			})
		}
	}

	// สร้าง print job สำหรับแต่ละเครื่องพิมพ์
	for printerID, jobs := range printerJobs {
		var printer models.Printer
		if err := tx.First(&printer, printerID).Error; err != nil {
			tx.Rollback()
			return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to fetch printer info",
			})
		}

		// รวมรายการทั้งหมดสำหรับเครื่องพิมพ์นี้
		var allItems []models.OrderItem
		var categoryNames []string
		for _, job := range jobs {
			allItems = append(allItems, job.Items...)
			var category models.Category
			if err := tx.First(&category, job.CategoryID).Error; err != nil {
				tx.Rollback()
				return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
					"error": "Failed to fetch category info",
				})
			}
			categoryNames = append(categoryNames, category.Name)
		}

		// สร้างเนื้อหาที่จะพิมพ์
		content := createOrderPrintContent(completeOrder, strings.Join(categoryNames, ", "), allItems)

		// สร้าง print job
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

// @Summary ดึงรายการออเดอร์ที่กำลังดำเนินการ
// @Description ดึงรายการออเดอร์ที่ยังไม่เสร็จสิ้น (pending, preparing, ready)
// @Produce json
// @Param uuid path string true "uuid"
// @Success 200 {array} OrderResponse
// @Router /api/orders/table/{uuid} [get]
// @Tags Order_ใหม่
func GetOrdersByid(c *fiber.Ctx) error {
	var orders []models.Order
	uuid := c.Params("uuid")

	if uuid == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "uuid require",
		})
	}

	if err := db.DB.Preload("Items").
		Preload("Items.MenuItem").
		Preload("Items.Options").
		Where("status IN ? AND uuid = ?", []string{"pending", "preparing", "ready"}, uuid).
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

	tx := db.DB.Begin()

	var order models.Order
	if err := tx.Where("uuid = ? AND table_id = ?", req.OrderUUID, req.Table_id).
		Preload("Items.MenuItem").
		Preload("Items.Options").
		First(&order).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Order not found",
		})
	}

	// ดึงข้อมูล OrderItems ที่จะยกเลิกพร้อมข้อมูล PromotionUsage และ MenuItem
	var orderItems []models.OrderItem
	if err := tx.Where("id IN ?", getOrderItemIDs(req.Items)).
		Preload("PromotionUsage").
		Preload("MenuItem").
		Find(&orderItems).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Failed to fetch order items",
		})
	}

	// แยกรายการตาม PromotionUsage
	promoUsageMap := make(map[uint][]models.OrderItem)
	var normalItems []models.OrderItem
	var printContents []string // ย้ายมาประกาศตรงนี้เพื่อให้ใช้ได้ทั้งโปรโมชั่นและรายการปกติ

	for _, item := range orderItems {
		if item.PromotionUsageID != nil {
			promoUsageMap[*item.PromotionUsageID] = append(promoUsageMap[*item.PromotionUsageID], item)
		} else {
			normalItems = append(normalItems, item)
		}
	}

	// จัดการรายการที่เป็นโปรโมชั่น
	for promoUsageID, items := range promoUsageMap {
		// ตรวจสอบว่ามีการยกเลิกทุกรายการในโปรโมชั่นหรือไม่
		var totalPromoItems int64
		if err := tx.Model(&models.OrderItem{}).
			Where("promotion_usage_id = ?", promoUsageID).
			Count(&totalPromoItems).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to check promotion items",
			})
		}

		if int64(len(items)) < totalPromoItems {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "Must cancel all items in a promotion together",
			})
		}

		// สร้าง print content สำหรับรายการโปรโมชั่นก่อนลบ
		for _, item := range items {
			printContents = append(printContents, fmt.Sprintf("%s|%d (โปรโมชั่น)", item.MenuItem.Name, item.Quantity))
		}

		// Hard delete รายการในโปรโมชั่น
		if err := tx.Unscoped().Where("promotion_usage_id = ?", promoUsageID).
			Delete(&models.OrderItem{}).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete promotion items",
			})
		}

		// Hard delete PromotionUsage
		if err := tx.Unscoped().Delete(&models.PromotionUsage{}, promoUsageID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to delete promotion usage",
			})
		}
	}

	// จัดการรายการปกติ
	for _, reqItem := range req.Items {
		// ข้ามรายการที่เป็นโปรโมชั่น (จัดการไปแล้ว)
		isPromoItem := false
		for _, items := range promoUsageMap {
			for _, item := range items {
				if item.ID == reqItem.OrderItemID {
					isPromoItem = true
					break
				}
			}
			if isPromoItem {
				break
			}
		}
		if isPromoItem {
			continue
		}

		var orderItem models.OrderItem
		if err := tx.Where("id = ?", reqItem.OrderItemID).
			Preload("MenuItem").
			Preload("Options").
			First(&orderItem).Error; err != nil {
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
		} else {
			orderItem.Quantity -= reqItem.Quantity
		}

		if err := tx.Save(&orderItem).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update order item",
			})
		}

		printContents = append(printContents, fmt.Sprintf("%s|%d", orderItem.MenuItem.Name, reqItem.Quantity))
	}

	// อัพเดทยอดรวม
	if err := updateOrdersTotalByUUIDAndTableID(tx, req.OrderUUID, req.Table_id); err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update order total",
		})
	}

	// ตรวจสอบว่าทุกรายการในออเดอร์ถูกยกเลิกหรือไม่
	var nonCancelledCount int64
	if err := tx.Model(&models.OrderItem{}).
		Where("order_id = ? AND status != ?", order.ID, "cancelled").
		Count(&nonCancelledCount).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to check order items status",
		})
	}

	// ถ้าทุกรายการถูกยกเลิก ให้อัพเดทสถานะของออเดอร์เป็น cancelled
	if nonCancelledCount == 0 {
		if err := tx.Model(&order).Update("status", "cancelled").Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to update order status",
			})
		}
	}

	// สร้าง print jobs
	if len(printContents) > 0 { // ตอนนี้จะรวมทั้งรายการปกติและโปรโมชั่น
		var printers []models.Printer
		err := tx.Distinct("printers.*").
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
	}

	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	return c.JSON(fiber.Map{"message": "Items cancelled successfully"})
}

func getOrderItemIDs(items []struct {
	OrderItemID uint `json:"order_item_id" binding:"required"`
	Quantity    int  `json:"quantity" binding:"required,min=1"`
}) []uint {
	ids := make([]uint, len(items))
	for i, item := range items {
		ids[i] = item.OrderItemID
	}
	return ids
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

// Helper function สำหรับคำนวณส่วนลด
func calculatePromotionSaving(promotion *models.Promotion, selectedItemIDs []uint) float64 {
	var normalPrice float64 = 0
	for _, id := range selectedItemIDs {
		for _, item := range promotion.Items {
			if item.MenuItemID == id {
				normalPrice += float64(item.MenuItem.Price)
				break
			}
		}
	}
	return normalPrice - promotion.Price
}

// @Summary จัดการออเดอร์ก่อนชำระเงิน
// @Description ใช้สำหรับการยกเลิกรายการอาหารในกรณีพิเศษก่อนชำระเงิน เช่น อาหารที่เสิร์ฟแล้วแต่ลูกค้าไม่ได้รับ
// @Accept json
// @Produce json
// @Param request body finalize_items_req true "ข้อมูลการยกเลิกรายการ"
// @Success 200 {object} map[string]interface{}
// @Router /api/orders/finalize [post]
// @Tags Order_ใหม่
func FinalizeOrderItems(c *fiber.Ctx) error {
	var req finalize_items_req
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ข้อมูลไม่ถูกต้อง",
		})
	}

	// ตรวจสอบว่ามีเหตุผลในการยกเลิกหรือไม่
	if req.Reason == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "กรุณาระบุเหตุผลในการยกเลิกรายการ",
		})
	}

	tx := db.DB.Begin()

	// ตรวจสอบพนักงาน
	var staff models.Users
	if err := tx.First(&staff, req.StaffID).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "ไม่พบข้อมูลพนักงาน",
		})
	}

	// ดึงข้อมูลออเดอร์ทั้งหมดของโต๊ะ (ที่ยังไม่เสร็จสิ้นหรือยกเลิก)
	var orders []models.Order
	if err := tx.Where("table_id = ? AND status NOT IN (?, ?)",
		req.TableID, "completed", "cancelled").
		Preload("Items.MenuItem").
		Preload("Items.Options.MenuOption").
		Find(&orders).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบออเดอร์ของโต๊ะนี้",
		})
	}

	var printContents []string

	// จัดการรายการเสริม
	for _, reqOption := range req.Options {
		var orderItemOption models.OrderItemOption
		if err := tx.Where("order_item_id = ? AND id = ?",
			reqOption.OrderItemID,
			reqOption.OptionID).
			Preload("MenuOption").
			Preload("OrderItem").
			First(&orderItemOption).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("ไม่พบตัวเลือกเสริมรหัส %d", reqOption.OptionID),
			})
		}

		// Soft delete ตัวเลือกเสริม
		if err := tx.Delete(&orderItemOption).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถยกเลิกตัวเลือกเสริมได้",
			})
		}

		// อัพเดทยอดรวมของออเดอร์
		if err := updateOrderTotal(tx, orderItemOption.OrderItem.OrderID); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถอัพเดทยอดรวมได้",
			})
		}

		printContents = append(printContents,
			fmt.Sprintf("ยกเลิกตัวเลือก %s", orderItemOption.MenuOption.Name))
	}

	// ดึงข้อมูล OrderItems ที่จะยกเลิก
	var orderItems []models.OrderItem
	ids := make([]uint, len(req.Items))
	for i, item := range req.Items {
		ids[i] = item.OrderItemID
	}

	if err := tx.Where("id IN ?", ids).
		Preload("PromotionUsage.Promotion").
		Preload("MenuItem").
		Find(&orderItems).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "ไม่พบรายการอาหารที่ต้องการยกเลิก",
		})
	}

	// แยกรายการตาม PromotionUsage
	promoUsageMap := make(map[uint][]models.OrderItem)
	var normalItems []models.OrderItem

	for _, item := range orderItems {
		if item.PromotionUsageID != nil {
			promoUsageMap[*item.PromotionUsageID] = append(promoUsageMap[*item.PromotionUsageID], item)
		} else {
			normalItems = append(normalItems, item)
		}
	}

	// จัดการรายการโปรโมชั่น
	for promoUsageID, items := range promoUsageMap {
		var totalPromoItems int64
		if err := tx.Model(&models.OrderItem{}).
			Where("promotion_usage_id = ?", promoUsageID).
			Count(&totalPromoItems).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถตรวจสอบรายการโปรโมชั่นได้",
			})
		}

		if int64(len(items)) < totalPromoItems {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "ต้องยกเลิกรายการในโปรโมชั่นทั้งหมดพร้อมกัน",
			})
		}

		for _, item := range items {
			printContents = append(printContents,
				fmt.Sprintf("%s|%d (โปรโมชั่น %s)",
					item.MenuItem.Name,
					item.Quantity,
					item.PromotionUsage.Promotion.Name))
		}

		// อัพเดทสถานะรายการในโปรโมชั่น
		if err := tx.Model(&models.OrderItem{}).
			Where("promotion_usage_id = ?", promoUsageID).
			Updates(map[string]interface{}{
				"status":   "cancelled",
				"quantity": 0,
			}).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถยกเลิกรายการโปรโมชั่นได้",
			})
		}

		// Soft delete promotion usage
		if err := tx.Delete(&models.PromotionUsage{}, promoUsageID).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถยกเลิกการใช้โปรโมชั่นได้",
			})
		}
	}

	// จัดการรายการปกติ
	for _, reqItem := range req.Items {
		// ข้ามรายการที่เป็นโปรโมชั่น (จัดการไปแล้ว)
		isPromoItem := false
		for _, items := range promoUsageMap {
			for _, item := range items {
				if item.ID == reqItem.OrderItemID {
					isPromoItem = true
					break
				}
			}
			if isPromoItem {
				break
			}
		}
		if isPromoItem {
			continue
		}

		var orderItem models.OrderItem
		if err := tx.Where("id = ?", reqItem.OrderItemID).
			Preload("MenuItem").
			First(&orderItem).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": fmt.Sprintf("ไม่พบรายการอาหารรหัส %d", reqItem.OrderItemID),
			})
		}

		if reqItem.Quantity > orderItem.Quantity {
			tx.Rollback()
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": fmt.Sprintf("จำนวนที่ยกเลิกมากกว่าจำนวนที่สั่ง สำหรับรายการ %d", reqItem.OrderItemID),
			})
		}

		// อัพเดทข้อมูลการยกเลิก
		updates := map[string]interface{}{}

		if orderItem.Quantity == reqItem.Quantity {
			updates["status"] = "cancelled"
			updates["quantity"] = 0
		} else {
			updates["quantity"] = orderItem.Quantity - reqItem.Quantity
		}

		if err := tx.Model(&orderItem).Updates(updates).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถอัพเดทรายการอาหารได้",
			})
		}

		printContents = append(printContents, fmt.Sprintf("%s|%d", orderItem.MenuItem.Name, reqItem.Quantity))
	}

	for _, order := range orders {
		// ตรวจสอบจำนวน OrderItems ที่ยังไม่ถูกยกเลิกในออเดอร์นี้
		var nonCancelledItemCount int64
		if err := tx.Model(&models.OrderItem{}).
			Where("order_id = ? AND status != ?", order.ID, "cancelled").
			Count(&nonCancelledItemCount).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถตรวจสอบสถานะรายการอาหารได้",
			})
		}

		// ถ้าไม่มีรายการที่ไม่ถูกยกเลิก ให้อัปเดตสถานะออเดอร์เป็น cancelled
		if nonCancelledItemCount == 0 {
			if err := tx.Model(&order).
				Update("status", "cancelled").Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "ไม่สามารถอัปเดตสถานะออเดอร์ได้",
				})
			}
		}
	}

	// อัพเดทยอดรวมของทุกออเดอร์
	for _, order := range orders {
		if err := updateOrderTotal(tx, order.ID); err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถอัพเดทยอดรวมได้",
			})
		}
	}

	// บันทึก log การยกเลิก
	notification := models.Notification{
		UserID:    staff.ID,
		Message:   fmt.Sprintf("ยกเลิกรายการอาหารโต๊ะ %d โดย %s เหตุผล: %s", req.TableID, staff.Name, req.Reason),
		Status:    "unread",
		CreatedAt: time.Now(),
	}

	if err := tx.Create(&notification).Error; err != nil {
		tx.Rollback()
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถบันทึก notification ได้",
		})
	}

	// ส่งการแจ้งเตือนไปยังครัวและพนักงาน
	if len(printContents) > 0 {
		// ค้นหาเครื่องพิมพ์ทั้งหมดที่เกี่ยวข้อง
		var printers []models.Printer
		if err := tx.Preload("Categories").Find(&printers).Error; err != nil {
			tx.Rollback()
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "ไม่สามารถค้นหาเครื่องพิมพ์ได้",
			})
		}

		// สร้างเนื้อหาที่จะพิมพ์
		printContentString := fmt.Sprintf(
			"ยกเลิกรายการก่อนชำระเงิน\n"+
				"โต๊ะ: %d\n"+
				"พนักงาน: %s\n"+
				"เหตุผล: %s\n"+
				"เวลา: %s\n\n"+
				"รายการที่ยกเลิก:\n%s",
			req.TableID,
			staff.Name,
			req.Reason,
			time.Now().Format("15:04:05"),
			strings.Join(printContents, "\n"))

		// สร้าง print job สำหรับแต่ละเครื่องพิมพ์
		for _, printer := range printers {
			printJob := models.PrintJob{
				PrinterID:         printer.ID,
				Status:            "pending",
				JobType:           "final_cancelation",
				CancelledQuantity: len(req.Items) + len(req.Options),
				Content:           []byte(printContentString),
			}

			if err := tx.Create(&printJob).Error; err != nil {
				tx.Rollback()
				return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": fmt.Sprintf("ไม่สามารถสร้างงานพิมพ์สำหรับเครื่องพิมพ์ %s", printer.Name),
				})
			}
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "ไม่สามารถบันทึกข้อมูลได้",
		})
	}

	return c.JSON(fiber.Map{
		"message":               "ยกเลิกรายการอาหารสำเร็จ",
		"staff":                 staff.Name,
		"table_id":              req.TableID,
		"cancelled_items_count": len(req.Items) + len(req.Options),
	})
}

type finalize_items_req struct {
	TableID uint                 `json:"table_id" binding:"required"`
	StaffID uint                 `json:"staff_id" binding:"required"`
	Reason  string               `json:"reason" binding:"required"`
	Items   []CancelItemDetail   `json:"items" binding:"required_without=Options"`
	Options []CancelOptionDetail `json:"options,omitempty"`
}

type CancelItemDetail struct {
	OrderItemID uint `json:"order_item_id" binding:"required"`
	Quantity    int  `json:"quantity" binding:"required,min=1"`
}

type CancelOptionDetail struct {
	OrderItemID uint `json:"order_item_id" binding:"required"`
	OptionID    uint `json:"option_id" binding:"required"`
}

// Helper function to update order total
func updateOrderTotal(tx *gorm.DB, orderID uint) error {
	var total float64

	// Calculate total from non-cancelled items
	if err := tx.Model(&models.OrderItem{}).
		Where("order_id = ? AND status != ?", orderID, "cancelled").
		Select("COALESCE(SUM(price * quantity), 0)").
		Scan(&total).Error; err != nil {
		return err
	}

	// Add options total
	var optionsTotal float64
	if err := tx.Model(&models.OrderItemOption{}).
		Joins("JOIN order_items ON order_items.id = order_item_options.order_item_id").
		Where("order_items.order_id = ? AND order_items.status != ?", orderID, "cancelled").
		Select("COALESCE(SUM(order_item_options.price * order_item_options.quantity), 0)").
		Scan(&optionsTotal).Error; err != nil {
		return err
	}

	total += optionsTotal

	// Update order total
	return tx.Model(&models.Order{}).Where("id = ?", orderID).Update("total", total).Error
}
