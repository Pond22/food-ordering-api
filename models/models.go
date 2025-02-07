package models

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
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
	ID            uint          `gorm:"primaryKey"`
	Name          string        `gorm:"not null"`
	NameEn        string        `gorm:"not null"`
	NameCh        string        `gorm:"not null"`
	Description   string        `gorm:"type:varchar(255)"`
	DescriptionEn string        `gorm:"type:varchar(255)"`
	DescriptionCh string        `gorm:"type:varchar(255)"`
	Image         []byte        `gorm:"type:bytea"`            // ฟิลด์ Image เป็น type bytea
	CategoryID    uint          `gorm:"not null"`              // foreign key ที่เชื่อมกับ Category
	Category      Category      `gorm:"foreignKey:CategoryID"` // ลิงก์ไปยังตาราง Category
	Price         int16         `gorm:"not null"`
	OptionGroups  []OptionGroup `gorm:"foreignKey:MenuItemID"`
	Is_available  bool          `gorm:"not null;default:true"` //พร้อมขายหรือไม่
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `json:"-" swaggerignore:"true"` //เอาไว้ทำ softdelete จะได้ restore ง่ายๆ
}

type MenuOption struct {
	ID          uint        `gorm:"primaryKey"`
	GroupID     uint        `gorm:"not null"`
	OptionGroup OptionGroup `gorm:"foreignKey:GroupID" json:"-"`
	Name        string      `gorm:"not null"`
	NameEn      string      `gorm:"not null"`
	NameCh      string      `gorm:"not null"`
	Price       float64     `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `json:"-" swaggerignore:"true"` //เอาไว้ทำ softdelete จะได้ restore ง่ายๆ
}

// กรณีที่อาหารอาจมีหลายตัวเลือกแต่เลือกได้แค่ == MaxSelections ที่กำหนด เช่น น้ำซุปเลือกได้แค่ 1 จาก 5
type OptionGroup struct {
	ID            uint         `gorm:"primaryKey"`
	MenuItemID    uint         `gorm:"not null"`
	MenuItem      MenuItem     `gorm:"foreignKey:MenuItemID" json:"-"`
	Name          string       `gorm:"not null"` // เช่น "น้ำซุป", "ความเผ็ด"
	NameEn        string       `gorm:"not null"`
	NameCh        string       `gorm:"not null"`
	MaxSelections int          `gorm:"not null;default:1"` // จำนวนที่เลือกได้
	IsRequired    bool         `gorm:"not null;default:false"`
	Options       []MenuOption `gorm:"foreignKey:GroupID"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `json:"-" swaggerignore:"true"` //เอาไว้ทำ softdelete จะได้ restore ง่ายๆ
}

type Category struct {
	ID        uint           `gorm:"primaryKey"`
	Name      string         `gorm:"not null"`
	NameEn    string         `gorm:"not null"`
	NameCh    string         `gorm:"not null"`
	DeletedAt gorm.DeletedAt `json:"-" swaggerignore:"true"`
}

type Table struct {
	ID        uint   `gorm:"primaryKey"`
	Name      string `gorm:"not null"`
	Capacity  int    `gorm:"not null;default:2"`           //default 2 ที่นั่ง
	Status    string `gorm:"not null;default:'available'"` // available, reserved, occupied, unavailable
	ParentID  *uint
	GroupID   *string // ID กลุ่มสำหรับโต๊ะที่รวมกัน
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `json:"-" swaggerignore:"true"`
}

type TableReservation struct {
	ID           uint      `gorm:"primaryKey"`
	TableID      uint      `gorm:"not null"`                   // Foreign key to Table
	Table        Table     `gorm:"foreignKey:TableID"`         // Relation to Table
	CustomerName string    `gorm:"type:varchar(100);not null"` // ชื่อลูกค้าที่จอง
	PhoneNumber  string    `gorm:"type:varchar(20);not null"`  // เบอร์โทรศัพท์
	GuestCount   int       `gorm:"not null"`                   // จำนวนลูกค้า
	ReservedFor  time.Time `gorm:"not null"`                   // เวลาที่ลูกค้าจะมาใช้บริการ
	Status       string    `gorm:"not null;default:'active'"`  // active, cancelled, completed
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `json:"-" swaggerignore:"true"`
}

// TableHistory - เก็บประวัติการย้าย/รวมโต๊ะ
// type TableHistory struct {
// 	ID          uint   `gorm:"primaryKey"`
// 	TableID     uint   `gorm:"not null"`
// 	Action      string `gorm:"not null"` // move, merge, split
// 	FromStatus  string `gorm:"not null"`
// 	ToStatus    string `gorm:"not null"`
// 	FromGroupID *string
// 	ToGroupID   *string
// 	StaffID     uint `gorm:"not null"`
// 	CreatedAt   time.Time
// }

type QRCode struct {
	ID        uint      `gorm:"primaryKey"`
	TableID   int       `gorm:"not null" json:"table_id"` // One-to-One กับ Table
	UUID      string    `gorm:"not null;uniqueIndex"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	Qr_Image  []byte    `gorm:"type:bytea"`
	ExpiryAt  time.Time `json:"expiry_at" db:"expiry_at"`
	IsActive  bool      `gorm:"not null;default:true;index"`
}

// FE-4 การจัดการออเดอร์
type Order struct {
	ID        uint    `gorm:"primaryKey"`
	UUID      string  `gorm:"not null;index"`
	TableID   int     `gorm:"not null"`
	Status    string  `gorm:"not null"` //  "completed", "uncompleted", "cancelled"
	Total     float64 `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	Items     []OrderItem
	ReceiptID *uint   `gorm:"index"`
	Receipt   Receipt `gorm:"foreignKey:ReceiptID"`
}

// FE-4 การจัดการออเดอร์
type OrderItem struct {
	ID               uint     `gorm:"primaryKey"`
	OrderID          uint     `gorm:"not null"`
	Order            Order    `gorm:"foreignKey:OrderID"`
	MenuItemID       uint     `gorm:"not null"`
	MenuItem         MenuItem `gorm:"foreignKey:MenuItemID"`
	Quantity         int      `gorm:"not null"`
	Price            float64  `gorm:"not null"`
	Notes            string
	Status           string            `gorm:"not null;default:'pending'"` // pending, served, cancelled
	ServedAt         *time.Time        // เวลาที่เสิร์ฟอาหาร
	Options          []OrderItemOption `gorm:"foreignKey:OrderItemID"` // เพิ่ม relation กับ options
	CreatedAt        time.Time
	UpdatedAt        time.Time
	PromotionUsageID *uint           `gorm:"index"`                       // เพิ่มฟิลด์ใหม่
	PromotionUsage   *PromotionUsage `gorm:"foreignKey:PromotionUsageID"` // เพิ่มความสัมพันธ์
}

// FE-4 การจัดการออเดอร์
type OrderItemOption struct {
	ID           uint       `gorm:"primaryKey"`
	OrderItemID  uint       `gorm:"not null"` // Foreign key to OrderItem
	OrderItem    OrderItem  `gorm:"foreignKey:OrderItemID"`
	MenuOptionID uint       `gorm:"not null"` // Foreign key to MenuOption
	MenuOption   MenuOption `gorm:"foreignKey:MenuOptionID"`
	Value        string     `gorm:"not null"`           // ค่าที่เลือก เช่น "เผ็ดมาก", "เพิ่มไข่ดาว"
	Quantity     int        `gorm:"not null;default:1"` // จำนวนตัวเลือกเสริม
	Price        float64    `gorm:"not null"`           // ราคา ณ เวลาที่สั่ง
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index" swaggerignore:"true"` // เพิ่ม Soft Delete
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

type POSSession struct { //เอาไว้เวลาพนักงานสแกนเพื่อเข้าทำงานที่เครื่องคอมหลักหรือ POS เคาท์เตอร์
	ID         uint      `gorm:"primaryKey"`
	StaffID    uint      `gorm:"not null"`           // เชื่อมกับตาราง Users
	Staff      Users     `gorm:"foreignKey:StaffID"` // Relation กับ Users
	StartTime  time.Time `gorm:"not null"`
	EndTime    *time.Time
	LoginToken string `gorm:"unique"`   // Token สำหรับ QR Login
	Status     string `gorm:"not null"` // active, ended
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// SalesAnalysis - ตารางสำหรับวิเคราะห์การขายและ Association
type SalesAnalysis struct {
	ID   uint      `gorm:"primaryKey"`
	Date time.Time `gorm:"not null"`

	// ข้อมูลการขาย
	MenuItemID   uint     `gorm:"not null"`
	MenuItem     MenuItem `gorm:"foreignKey:MenuItemID"`
	OrderCount   int      `gorm:"not null"`
	TotalRevenue float64  `gorm:"not null"`

	// ข้อมูล Association
	RelatedItemID uint     `gorm:"not null"` // เมนูที่ถูกสั่งร่วม
	RelatedItem   MenuItem `gorm:"foreignKey:RelatedItemID"`
	JointCount    int      `gorm:"not null"` // จำนวนครั้งที่ถูกสั่งร่วมกัน

	// ข้อมูลเพิ่มเติมสำหรับวิเคราะห์
	TimeSegment string `gorm:"not null"` // ช่วงเวลา (เช้า, กลางวัน, เย็น)
	DayType     string `gorm:"not null"` // วันธรรมดา/วันหยุด

	CreatedAt time.Time
	UpdatedAt time.Time
}

// /-----------------------------------------------------
// DiscountType - ประเภทส่วนลด (เช่น percentage, amount)
type DiscountType struct {
	ID        uint    `gorm:"primaryKey"`
	Name      string  `gorm:"not null;unique"` // เช่น "ส่วนลดพนักงาน", "ส่วนลดสมาชิก"
	Type      string  `gorm:"not null"`        // percentage/amount
	Value     float64 `gorm:"not null"`        // จำนวนหรือเปอร์เซ็นต์
	IsActive  bool    `gorm:"not null;default:true"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `json:"-" swaggerignore:"true"`
}

// AdditionalChargeType - ประเภทค่าใช้จ่ายเพิ่มเติม
type AdditionalChargeType struct {
	ID            uint    `gorm:"primaryKey"`
	Name          string  `gorm:"not null;unique"` // เช่น "แก้วแตก", "จานแตก"
	DefaultAmount float64 `gorm:"not null"`        // ราคาเริ่มต้น
	IsActive      bool    `gorm:"not null;default:true"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     gorm.DeletedAt `json:"-" swaggerignore:"true"`
}

type ReceiptDiscount struct {
	ID             uint         `gorm:"primaryKey"`
	ReceiptID      uint         `gorm:"not null"` // เปลี่ยนจาก OrderID
	Receipt        Receipt      `gorm:"foreignKey:ReceiptID"`
	DiscountTypeID uint         `gorm:"not null"`
	DiscountType   DiscountType `gorm:"foreignKey:DiscountTypeID"`
	Value          float64      `gorm:"not null"`
	StaffID        uint         `gorm:"not null"`
	Staff          Users        `gorm:"foreignKey:StaffID"`
	Reason         string
	CreatedAt      time.Time
}

type ReceiptCharge struct {
	ID           uint                 `gorm:"primaryKey"`
	ReceiptID    uint                 `gorm:"not null"` // เปลี่ยนจาก OrderID
	Receipt      Receipt              `gorm:"foreignKey:ReceiptID"`
	ChargeTypeID uint                 `gorm:"not null"`
	ChargeType   AdditionalChargeType `gorm:"foreignKey:ChargeTypeID"`
	Amount       float64              `gorm:"not null"`
	Quantity     int                  `gorm:"not null;default:1"`
	StaffID      uint                 `gorm:"not null"`
	Staff        Users                `gorm:"foreignKey:StaffID"`
	Note         string
	CreatedAt    time.Time
}

// ใบเสร็จ
type Receipt struct {
	ID      uint   `gorm:"primaryKey"`
	UUID    string `gorm:"not null;index"`
	TableID string `gorm:"type:text"`
	// Orders        []Order `gorm:"foreignKey:ReceiptID"`
	Orders        []Order `gorm:"foreignKey:ReceiptID"`
	OrderID       *uint   `gorm:"index"`
	SubTotal      float64 // ยอดรวมทุก order
	DiscountTotal float64
	ChargeTotal   float64
	ServiceCharge float64
	Total         float64
	PaymentMethod string
	StaffID       uint
	Staff         Users             `gorm:"foreignKey:StaffID"`
	Discounts     []ReceiptDiscount // เปลี่ยนจาก OrderDiscount เพราะมันไม่ตอบโจทย์ T T
	Charges       []ReceiptCharge   // เปลี่ยนจาก OrderAdditionalCharge
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

type Printer struct {
	ID          uint       `gorm:"primaryKey"`
	Name        string     `gorm:"not null"`                  // ชื่อเครื่องพิมพ์
	Type        string     `gorm:"not null"`                  // 'network' หรือ 'usb'
	IPAddress   string     `gorm:"unique;not null"`           // IP Address
	Port        int        `gorm:"not null"`                  // Port number
	VendorID    string     `gorm:"index"`                     // สำหรับ USB printer
	ProductID   string     `gorm:"index"`                     // สำหรับ USB printer
	Department  string     `gorm:"not null"`                  // แผนก/ฝ่ายที่ใช้งาน
	Description string     `gorm:"type:text"`                 // รายละเอียดเพิ่มเติม
	Status      string     `gorm:"not null;default:'active'"` // สถานะ: active, inactive, maintenance
	LastSeen    time.Time  // เวลาที่เห็นเครื่องพิมพ์ล่าสุด
	Categories  []Category `gorm:"many2many:printer_categories;"` // หมวดหมู่ที่พิมพ์ได้
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}

// type PrintJob struct {
// 	ID        uint   `gorm:"primaryKey"`
// 	PrinterIP string `gorm:"not null"`
// 	OrderID   *uint  // nullable, เพราะอาจเป็นการพิมพ์ทดสอบ
// 	Content   []byte `gorm:"type:bytea"`
// 	Status    string `gorm:"not null;default:'pending'"` // pending, processing, completed, failed
// 	CreatedAt time.Time
// 	UpdatedAt time.Time
// 	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

//		// Optional: เพิ่ม relation กับ Order ถ้าต้องการ
//		Order *Order `gorm:"foreignKey:OrderID"`
//	}
type PrintJob struct {
	ID                uint    `gorm:"primaryKey"`
	PrinterID         uint    `gorm:"not null;index"`
	Printer           Printer `gorm:"foreignKey:PrinterID"`
	OrderID           *uint   // nullable
	ReceiptID         *uint
	Content           []byte `gorm:"type:bytea"`
	Status            string `gorm:"not null;default:'pending'"`
	CreatedAt         time.Time
	UpdatedAt         time.Time
	DeletedAt         gorm.DeletedAt `gorm:"index" json:"-"`
	JobType           string
	CancelledQuantity int `gorm:"default:0"`

	Order   *Order   `gorm:"foreignKey:OrderID"`
	Receipt *Receipt `gorm:"foreignKey:ReceiptID"`
}

type Notification struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	UserID    uint      `json:"user_id" gorm:"not null"`      // ID ของผู้ใช้ที่ได้รับการแจ้งเตือน
	Message   string    `json:"message" gorm:"type:text"`     // ข้อความแจ้งเตือน
	Status    string    `json:"status" gorm:"default:unread"` // สถานะ "unread" หรือ "read"
	CreatedAt time.Time `json:"created_at"`
}