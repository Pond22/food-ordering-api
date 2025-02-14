package main

import (
	"fmt"
	"food-ordering-api/models"
	"log"
	"os"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	var (
		DBUser     = os.Getenv("DB_USER")     // ค่าจาก environment variable DB_USER
		DBPassword = os.Getenv("DB_PASSWORD") // ค่าจาก environment variable DB_PASSWORD
		DBName     = os.Getenv("DB_NAME")     // ค่าจาก environment variable DB_NAME
		DBHost     = os.Getenv("DB_HOST")     // ค่าจาก environment variable DB_HOST
		DBPort     = os.Getenv("DB_PORT")     // ค่าจาก environment variable DB_PORT
	)

	// สร้าง DSN string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=Asia/Bangkok",
		DBHost, DBUser, DBPassword, DBName, DBPort)

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// ข้อมูล user ที่จะสร้าง
	username := "admin"
	password := "admin123" // ควรเปลี่ยนรหัสผ่านหลังจาก login ครั้งแรก
	role := models.RoleManager

	// เข้ารหัสรหัสผ่าน
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatal("Failed to hash password:", err)
	}

	// สร้าง user ใหม่
	user := models.Users{
		Username:  username,
		Password:  string(hashedPassword),
		Role:      role,
		Name:      "Admin",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// ตรวจสอบว่ามี user นี้อยู่แล้วหรือไม่
	var existingUser models.Users
	result := database.Where("username = ?", username).First(&existingUser)
	if result.Error == nil {
		fmt.Printf("User '%s' already exists\n", username)
		return
	}

	// เพิ่ม user ลงในฐานข้อมูล
	if err := database.Create(&user).Error; err != nil {
		log.Fatal("Failed to create user:", err)
	}

	fmt.Printf("Successfully created user:\nUsername: %s\nPassword: %s\nRole: %s\n", username, password, role)
	fmt.Println("Please change your password after first login!")
}
