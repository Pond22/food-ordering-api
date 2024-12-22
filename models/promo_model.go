package models

import (
	"time"
)

type Promotion struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null"`
	Description string
	StartDate   time.Time       `gorm:"not null"`
	EndDate     time.Time       `gorm:"not null"`
	IsActive    bool            `gorm:"not null;default:true"`
	Price       float64         `gorm:"not null"` // เพิ่มฟิลด์ราคา
	Items       []PromotionItem `gorm:"foreignKey:PromotionID"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type PromotionItem struct {
	ID          uint      `gorm:"primaryKey"`
	PromotionID uint      `gorm:"not null;index"`
	Promotion   Promotion `gorm:"foreignKey:PromotionID"`
	MenuItemID  uint      `gorm:"not null;index"`
	MenuItem    MenuItem  `gorm:"foreignKey:MenuItemID"`
	Quantity    int       `gorm:"not null;default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// PromotionUsage - บันทึกการใช้โปรโมชั่น
type PromotionUsage struct {
	ID          uint      `gorm:"primaryKey"`
	PromotionID uint      `gorm:"not null;index"`
	Promotion   Promotion `gorm:"foreignKey:PromotionID"`
	OrderID     uint      `gorm:"not null;index"`
	Order       Order     `gorm:"foreignKey:OrderID"`
	SaveAmount  float64   `gorm:"not null"` // จำนวนเงินที่ประหยัดได้
	CreatedAt   time.Time
}