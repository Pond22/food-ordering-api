package db

import (
	"fmt"
	"food-ordering-api/config"
	"food-ordering-api/models"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

//	func InitDatabase() {
//		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
//			config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)
//		db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
//		if err != nil {
//			panic(fmt.Sprintf("Failed to connect to database: %v", err))
//		}
//		DB = db
//		migrate()
//	}
func init() {
	// ตั้งค่า timezone เป็น Asia/Bangkok
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		log.Fatalf("Failed to load location: %v", err)
	}
	time.Local = loc
}

func InitDatabase() {
	// สร้าง connection string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)

	// กำหนด logger configuration
	logConfig := logger.Config{
		SlowThreshold:             time.Second, // แสดง log เฉพาะ query ที่ใช้เวลามากกว่า 1 วินาที
		LogLevel:                  logger.Warn, // เปลี่ยนเป็น Warn เพื่อลดจำนวน logs
		IgnoreRecordNotFoundError: true,
		Colorful:                  true,
	}

	// สร้าง logger instance
	newLogger := logger.New(
		log.New(log.Writer(), "\r\n", log.LstdFlags),
		logConfig,
	)

	// เปิด connection กับ database
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
		NowFunc: func() time.Time {
			return time.Now().Local() // ใช้ timezone ของเครื่อง
		},
	})

	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// ดึง *sql.DB จาก GORM
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("Failed to get database instance: %v", err)
	}

	// ตั้งค่า Connection Pool
	sqlDB.SetMaxIdleConns(10)                  // จำนวน connection ที่สามารถ idle ได้
	sqlDB.SetMaxOpenConns(100)                 // จำนวน connection สูงสุดที่สามารถเปิดได้พร้อมกัน
	sqlDB.SetConnMaxLifetime(time.Hour)        // อายุสูงสุดของ connection
	sqlDB.SetConnMaxIdleTime(time.Minute * 30) // เวลา idle สูงสุดก่อนจะถูกปิด

	DB = db
	log.Println("Successfully connected to database with connection pool")

	migrate()
}

func CloseDatabase() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err != nil {
			log.Printf("Error getting database instance: %v", err)
			return
		}
		if err := sqlDB.Close(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}
}

func migrate() {
	DB.AutoMigrate(&models.Users{})
	DB.AutoMigrate(&models.QRCode{})
	DB.AutoMigrate(&models.MenuItem{})
	DB.AutoMigrate(&models.Category{})
	DB.AutoMigrate(&models.Table{})
	DB.AutoMigrate(&models.TableReservation{})
	DB.AutoMigrate(&models.SalesAnalysis{})
	DB.AutoMigrate(&models.Order{})
	DB.AutoMigrate(&models.OrderItem{})
	DB.AutoMigrate(&models.MenuOption{})
	DB.AutoMigrate(&models.OrderItemOption{})
	DB.AutoMigrate(&models.DiscountType{})
	DB.AutoMigrate(&models.AdditionalChargeType{})
	DB.AutoMigrate(&models.ReceiptDiscount{})
	DB.AutoMigrate(&models.ReceiptCharge{})
	DB.AutoMigrate(&models.ReceiptDiscount{})
	DB.AutoMigrate(&models.ReceiptCharge{})
	DB.AutoMigrate(&models.Receipt{})
	DB.AutoMigrate(&models.OptionGroup{})
	DB.AutoMigrate(&models.Promotion{})
	DB.AutoMigrate(&models.PromotionItem{})
	DB.AutoMigrate(&models.PromotionUsage{})
	DB.AutoMigrate(&models.Printer{})
	DB.AutoMigrate(&models.PrintJob{})
	DB.AutoMigrate(&models.POSSession{})
	// DB.AutoMigrate(&models.PromoSalesReport{})
	DB.AutoMigrate(&models.Notification{})
	DB.AutoMigrate(&models.ReservationRules{})
}
