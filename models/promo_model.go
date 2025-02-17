package models

import (
	"time"

	"gorm.io/gorm"
)

type Promotion struct {
	ID            uint   `gorm:"primaryKey"`
	Name          string `gorm:"not null"`
	NameEn        string
	NameCh        string
	Description   string
	DescriptionEn string
	DescriptionCh string
	StartDate     time.Time       `gorm:"not null"`
	EndDate       time.Time       `gorm:"not null"`
	IsActive      bool            `gorm:"not null;default:true"`
	Price         float64         `gorm:"not null"` // เพิ่มฟิลด์ราคา
	Items         []PromotionItem `gorm:"foreignKey:PromotionID"`
	Image         []byte          `gorm:"type:bytea"` // รูปโปร
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `json:"-" swaggerignore:"true"`
	MaxSelections int            `gorm:"not null;default:0"`
	MinSelections int            `gorm:"not null;default:0"`
	TotalItems    int            `gorm:"not null"` // จำนวนรายการทั้งหมดในโปรโมชั่น
}

type PromotionItem struct {
	ID          uint      `gorm:"primaryKey"`
	PromotionID uint      `gorm:"not null;index"`
	Promotion   Promotion `json:"-" gorm:"foreignKey:PromotionID;references:ID;constraint:OnDelete:CASCADE"`
	MenuItemID  uint      `gorm:"not null;index"`
	MenuItem    MenuItem  `gorm:"foreignKey:MenuItemID"`
	Quantity    int       `gorm:"not null;default:1"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `json:"-" swaggerignore:"true"`
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
	DeletedAt   gorm.DeletedAt `json:"-" swaggerignore:"true"`
}
