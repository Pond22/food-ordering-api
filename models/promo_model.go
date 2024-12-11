package models

import (
	"time"
)

// PromotionBundle
type PromotionBundle struct {
	ID          uint   `gorm:"primaryKey"`
	Name        string `gorm:"not null;index"` // เพิ่ม index เพื่อค้นหาเร็วขึ้น
	Description string
	Price       float64   `gorm:"not null;check:price >= 0"`       // ตรวจสอบราคาต้องไม่ติดลบ
	SelectCount int       `gorm:"not null;check:select_count > 0"` // ต้องเลือกอย่างน้อย 1 รายการ
	StartDate   time.Time `gorm:"not null;index"`                  // เพิ่ม index สำหรับค้นหาโปรโมชั่นที่กำลังใช้งาน
	EndDate     time.Time `gorm:"not null;index"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// BundleMenuItem
type BundleMenuItem struct {
	ID          uint            `gorm:"primaryKey"`
	BundleID    uint            `gorm:"not null;index:idx_bundle_menu"` // compound index
	Bundle      PromotionBundle `gorm:"foreignKey:BundleID"`
	MenuItemID  uint            `gorm:"not null;index:idx_bundle_menu"` // สำหรับค้นหาเมนูในบันเดิล
	MenuItem    MenuItem        `gorm:"foreignKey:MenuItemID"`
	MaxQuantity int             `gorm:"check:max_quantity >= 0"` // ถ้าเป็น 0 คือไม่จำกัด
	CreatedAt   time.Time
}

// OrderBundle
type OrderBundle struct {
	ID        uint            `gorm:"primaryKey"`
	OrderID   uint            `gorm:"not null;index"`
	Order     Order           `gorm:"foreignKey:OrderID"`
	BundleID  uint            `gorm:"not null;index"`
	Bundle    PromotionBundle `gorm:"foreignKey:BundleID"`
	Price     float64         `gorm:"not null;check:price >= 0"`
	CreatedAt time.Time
}

// OrderBundleItem
type OrderBundleItem struct {
	ID            uint        `gorm:"primaryKey"`
	OrderBundleID uint        `gorm:"not null;index:idx_bundle_menu_order"` // compound index
	OrderBundle   OrderBundle `gorm:"foreignKey:OrderBundleID"`
	MenuItemID    uint        `gorm:"not null;index:idx_bundle_menu_order"`
	MenuItem      MenuItem    `gorm:"foreignKey:MenuItemID"`
	Quantity      int         `gorm:"not null;check:quantity > 0"`
	CreatedAt     time.Time
}

// PromoSalesReport - สำหรับรายงานการขาย
type PromoSalesReport struct {
	MenuID        uint    `gorm:"primaryKey"`
	MenuName      string  `gorm:"not null"`
	OriginalPrice float64 `gorm:"not null"`
	PromoPrice    float64 `gorm:"not null"`
	SalesBefore   int     `gorm:"not null;default:0"`
	SalesDuring   int     `gorm:"not null;default:0"`
	SalesAfter    int     `gorm:"not null;default:0"`
	Revenue       float64 `gorm:"not null;default:0"`
	Profit        float64 `gorm:"not null;default:0"`
}
