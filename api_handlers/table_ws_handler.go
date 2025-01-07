package api_handlers

import (
	"food-ordering-api/models"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

func TableWebSocketHandler(db *gorm.DB) fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		defer conn.Close()

		// background task สำหรับส่งข้อมูลซ้ำๆ
		go func() {
			ticker := time.NewTicker(2 * time.Second)
			defer ticker.Stop()

			for range ticker.C {
				var tables []models.Table
				if err := db.Find(&tables).Error; err == nil {
					if err := conn.WriteJSON(map[string]interface{}{
						"type": "table_update",
						"data": tables,
					}); err != nil {
						return
					}
				}
			}
		}()

		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	})
}
