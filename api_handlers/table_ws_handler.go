package api_handlers

import (
	"log"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"gorm.io/gorm"
)

type TableResponse struct {
	ID        uint      `json:"ID"`
	Name      string    `json:"Name"`
	Capacity  int       `json:"Capacity"`
	Status    string    `json:"Status"`
	ParentID  *uint     `json:"ParentID"`
	GroupID   *string   `json:"GroupID"`
	UUID      string    `json:"UUID,omitempty"`
	CreatedAt time.Time `json:"CreatedAt"`
	UpdatedAt time.Time `json:"UpdatedAt"`
}

func TableWebSocketHandler(db *gorm.DB) fiber.Handler {
	return websocket.New(func(c *websocket.Conn) {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			var tables []TableResponse

			if err := db.Table("tables").
				Select("tables.*, COALESCE(qr_codes.uuid, '') as uuid").
				Joins("LEFT JOIN qr_codes ON qr_codes.table_id = tables.id AND qr_codes.is_active = ? AND qr_codes.expiry_at > ?", true, time.Now()).
				Scan(&tables).Error; err != nil {
				log.Printf("Error querying tables: %v", err)
				continue
			}

			if err := c.WriteJSON(map[string]interface{}{
				"type": "table_update",
				"data": tables,
			}); err != nil {
				return
			}
		}
	})
}

// {
//     "data": [
//         {
//             "ID": 6,
//             "Name": "A4",
//             "Capacity": 3,
//             "Status": "available",
//             "ParentID": null,
//             "GroupID": null,
//             "CreatedAt": "2025-01-13T16:01:10.162931Z",
//             "UpdatedAt": "2025-01-13T16:03:00.453988Z"
//         },
//         {
//             "ID": 7,
//             "Name": "A5",
//             "Capacity": 555,
//             "Status": "available",
//             "ParentID": null,
//             "GroupID": null,
//             "CreatedAt": "2025-01-13T16:01:56.532545Z",
//             "UpdatedAt": "2025-01-13T16:03:00.455973Z"
//         },
//         {
//             "ID": 5,
//             "Name": "A3",
//             "Capacity": 2,
//             "Status": "available",
//             "ParentID": null,
//             "GroupID": null,
//             "CreatedAt": "2025-01-13T14:14:50.21426Z",
//             "UpdatedAt": "2025-01-13T14:22:09.507785Z"
//         },
//         {
//             "ID": 4,
//             "Name": "A2",
//             "Capacity": 2,
//             "Status": "occupied",
//             "ParentID": 1,
//             "GroupID": "dc9a8c4f-2af8-4722-a65d-ea64c8674c53",
//             "CreatedAt": "2025-01-13T14:14:28.941067Z",
//             "UpdatedAt": "2025-01-27T20:34:05.682767Z"
//         },
//         {
//             "ID": 1,
//             "Name": "A1",
//             "Capacity": 2,
//             "Status": "occupied",
//             "ParentID": null,
//             "GroupID": "dc9a8c4f-2af8-4722-a65d-ea64c8674c53",
//             "CreatedAt": "2025-01-13T14:10:28.955571Z",
//             "UpdatedAt": "2025-01-27T20:34:05.682767Z"
//         }
//     ],
//     "type": "table_update"
// }
