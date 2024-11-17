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

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

type CreateMenuRequest struct {
	Name        string `json:"Name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=255"`
	Image       []byte `json:"image" binding:"max=5242880"` // max 5MB
	CategoryID  uint   `json:"CategoryID" binding:"required,min=1"`
}
type Category struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"not null"`
}

type Table struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type QRCode struct {
	ID        uint      `gorm:"primaryKey"`
	TableID   int       `gorm:"not null;unique" json:"table_id"` // One-to-One กับ Table
	UUID      string    `gorm:"not null"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Qr_Image  []byte    `gorm:"type:bytea"`
	ExpiryAt  time.Time `json:"expiry_at" db:"expiry_at"`
	IsActive  bool      `json:"is_active" db:"is_active"`
}
