package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"
)

// -----
type UserRole string

const (
	RoleStaff   UserRole = "staff"
	RoleManager UserRole = "manager"
	RoleChef    UserRole = "chef"
)

// Value - Implementation for driver.Valuer interface
func (r UserRole) Value() (driver.Value, error) {
	switch r {
	case RoleStaff, RoleManager, RoleChef:
		return string(r), nil
	default:
		return nil, errors.New("invalid role")
	}
}

// Scan - Implementation for sql.Scanner interface
func (r *UserRole) Scan(value interface{}) error {
	if value == nil {
		return errors.New("role cannot be null")
	}

	str, ok := value.(string)
	if !ok {
		return errors.New("invalid role type")
	}

	switch UserRole(str) {
	case RoleStaff, RoleManager, RoleChef:
		*r = UserRole(str)
		return nil
	default:
		return fmt.Errorf("invalid role: %s", str)
	}
}

func (r UserRole) IsValid() bool {
	switch r {
	case RoleStaff, RoleManager, RoleChef:
		return true
	}
	return false
}

type MenuItem struct {
	ID          uint     `gorm:"primaryKey"`
	Name        string   `gorm:"not null"`
	Description string   `gorm:"type:varchar(255)"`
	Image       []byte   `gorm:"type:bytea"`            // ฟิลด์ Image เป็น type bytea
	CategoryID  uint     `gorm:"not null"`              // foreign key ที่เชื่อมกับ Category
	Category    Category `gorm:"foreignKey:CategoryID"` // ลิงก์ไปยังตาราง Category
	Price       int16    `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
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

// FE-4 การจัดการออเดอร์
type Order struct {
	ID        uint    `gorm:"primaryKey"`
	TableID   uint    `gorm:"not null"` // Foreign key to Table
	Status    string  `gorm:"not null"` // e.g., "pending", "cooking", "served", "completed"
	Total     float64 `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Items     []OrderItem
}

// FE-4 การจัดการออเดอร์
type OrderItem struct {
	ID         uint     `gorm:"primaryKey"`
	OrderID    uint     `gorm:"not null"` // Foreign key to Order
	Order      Order    `gorm:"foreignKey:OrderID"`
	MenuItemID uint     `gorm:"not null"` // Foreign key to MenuItem
	MenuItem   MenuItem `gorm:"foreignKey:MenuItemID"`
	Quantity   int      `gorm:"not null"`
	Price      float64  `gorm:"not null"` // Price at time of order
	Notes      string   // Special instructions
	Status     string   `gorm:"not null"` // e.g., "pending", "cooking", "served"
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// FE-2 ระบบจัดการผู้ใช้งาน
type Users struct {
	ID        uint     `gorm:"primaryKey"`
	Username  string   `gorm:"unique;not null"`
	Password  string   `gorm:"not null"`
	Role      UserRole `gorm:"type:varchar(10);not null;check:role in ('staff', 'manager', 'chef')"`
	Name      string   `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

// FE-5 รายงานและวิเคราะห์
type DailySales struct {
	ID            uint      `gorm:"primaryKey"`
	Date          time.Time `gorm:"not null;unique"`
	TotalSales    float64   `gorm:"not null"`
	OrderCount    int       `gorm:"not null"`
	CustomerCount int       `gorm:"not null"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// FE-5 รายงานและวิเคราะห์
type MenuItemStats struct {
	ID           uint      `gorm:"primaryKey"`
	MenuItemID   uint      `gorm:"not null"` // Foreign key to MenuItem
	OrderCount   int       `gorm:"not null"`
	TotalRevenue float64   `gorm:"not null"`
	Date         time.Time `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
