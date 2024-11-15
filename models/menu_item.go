package models

import (
	"time"
)

type MenuItem struct {
	ID          uint     `gorm:"primaryKey"`
	Name        string   `gorm:"not null"`
	Description string   `gorm:"type:varchar(255)"`
	Image       []byte   `gorm:"type:bytea"`            // ฟิลด์ Image เป็น type bytea
	CategoryID  uint     `gorm:"not null"`              // foreign key ที่เชื่อมกับ Category
	Category    Category `gorm:"foreignKey:CategoryID"` // ลิงก์ไปยังตาราง Category
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Category struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"not null"`
}

type QRCode struct {
	ID        string    `json:"id" db:"id"`                 // UUID of the table
	TableID   string    `json:"table_id" db:"table_id"`     // Table ID
	CreatedAt time.Time `json:"created_at" db:"created_at"` // Time when the QR code was created
	Qr_Image  []byte    `gorm:"type:bytea"`
	ExpiryAt  time.Time `json:"expiry_at" db:"expiry_at"` // Expiration time for the QR code
	IsActive  bool      `json:"is_active" db:"is_active"` // Whether the QR code is still active
}
