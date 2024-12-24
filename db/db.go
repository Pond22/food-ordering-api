package db

import (
	"fmt"
	"food-ordering-api/config"
	"food-ordering-api/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDatabase() {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
		config.DBHost, config.DBUser, config.DBPassword, config.DBName, config.DBPort)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(fmt.Sprintf("Failed to connect to database: %v", err))
	}
	DB = db
	migrate()
}

func migrate() {
	DB.AutoMigrate(&models.Users{})
	DB.AutoMigrate(&models.QRCode{})
	DB.AutoMigrate(&models.MenuItem{})
	DB.AutoMigrate(&models.MenuItemStats{})
	DB.AutoMigrate(&models.Category{})
	DB.AutoMigrate(&models.Table{})
	DB.AutoMigrate(&models.DailySales{})
	DB.AutoMigrate(&models.Order{})
	DB.AutoMigrate(&models.OrderItem{})
	DB.AutoMigrate(&models.MenuOption{})
	DB.AutoMigrate(&models.OrderItemOption{})
	DB.AutoMigrate(&models.DiscountType{})
	DB.AutoMigrate(&models.AdditionalChargeType{})
	DB.AutoMigrate(&models.OrderDiscount{})
	DB.AutoMigrate(&models.OrderAdditionalCharge{})
	DB.AutoMigrate(&models.Receipt{})
	DB.AutoMigrate(&models.OptionGroup{})
	DB.AutoMigrate(&models.Promotion{})
	DB.AutoMigrate(&models.PromotionItem{})
	DB.AutoMigrate(&models.PromotionUsage{})
	DB.AutoMigrate(&models.Printer{})
	DB.AutoMigrate(&models.PrintJob{})
	// DB.AutoMigrate(&models.PromoSalesReport{})
	// DB.AutoMigrate(&models.TableHistory{})
}
