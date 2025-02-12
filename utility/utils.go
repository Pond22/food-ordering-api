package utils

import (
	"errors"
	"fmt"
	"food-ordering-api/models"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// GetUserFromToken ดึงข้อมูลผู้ใช้จาก JWT token
func GetUserFromToken(c *fiber.Ctx) (*jwt.MapClaims, error) {
	// ดึง token จาก Authorization header
	token := c.Get("Authorization")
	if token == "" {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Missing authorization token")
	}

	// แยก Bearer token
	tokenString := strings.Replace(token, "Bearer ", "", 1)

	// ตรวจสอบและถอดรหัส token
	claims := &jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte("your-secret-key"), nil // ควรย้าย secret key ไปไว้ใน config
	})

	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid token")
	}

	return claims, nil
}

// IsManager ตรวจสอบว่าผู้ใช้เป็น manager หรือไม่
func IsManager(c *fiber.Ctx) bool {
	claims, err := GetUserFromToken(c)
	if err != nil {
		return false
	}

	role := (*claims)["role"].(string)
	return role == string(models.RoleManager)
}

// IsChef ตรวจสอบว่าผู้ใช้เป็น chef หรือไม่
func IsChef(c *fiber.Ctx) bool {
	claims, err := GetUserFromToken(c)
	if err != nil {
		return false
	}

	role := (*claims)["role"].(string)
	return role == string(models.RoleChef)
}

// IsStaff ตรวจสอบว่าผู้ใช้เป็น staff หรือไม่
func IsStaff(c *fiber.Ctx) bool {
	claims, err := GetUserFromToken(c)
	if err != nil {
		return false
	}

	role := (*claims)["role"].(string)
	return role == string(models.RoleStaff)
}

// GetUserID ดึง user ID จาก token
func GetUserID(c *fiber.Ctx) (uint, error) {
	claims, err := GetUserFromToken(c)
	if err != nil {
		return 0, err
	}

	userID := uint((*claims)["user_id"].(float64))
	return userID, nil
}

// GetUserRole ดึง role จาก token
func GetUserRole(c *fiber.Ctx) (models.UserRole, error) {
	claims, err := GetUserFromToken(c)
	if err != nil {
		return "", err
	}

	roleInterface, exists := (*claims)["role"]
	if !exists {
		return "", errors.New("role not found in token")
	}

	roleStr, ok := roleInterface.(string)
	if !ok {
		return "", errors.New("invalid role type in token")
	}

	// เพิ่มการตรวจสอบค่า role ที่ได้
	role := models.UserRole(roleStr)
	if !role.IsValid() {
		return "", fmt.Errorf("invalid role value: %s", roleStr)
	}

	return role, nil
}

// CheckPermission ตรวจสอบว่าผู้ใช้มีสิทธิ์เข้าถึงหรือไม่
func CheckPermission(c *fiber.Ctx, allowedRoles ...models.UserRole) error {
	userRole, err := GetUserRole(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized access")
	}

	// เพิ่ม logging เพื่อ debug
	fmt.Printf("Current user role: %s\n", userRole)
	fmt.Printf("Allowed roles: %v\n", allowedRoles)

	// ตรวจสอบ role อย่างเข้มงวด
	isAllowed := false
	for _, role := range allowedRoles {
		if userRole == role {
			isAllowed = true
			break
		}
	}

	if !isAllowed {
		return fiber.NewError(fiber.StatusForbidden,
			fmt.Sprintf("Permission denied. Required roles: %v, Current role: %s", allowedRoles, userRole))
	}

	return nil
}

// AuthRequired middleware สำหรับตรวจสอบการ authentication
func AuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		_, err := GetUserFromToken(c)
		if err != nil {
			return err
		}
		return c.Next()
	}
}

// RoleRequired middleware สำหรับตรวจสอบ role
func RoleRequired(roles ...models.UserRole) fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := GetUserFromToken(c)
		if err != nil {
			fmt.Printf("Token error: %v\n", err)
			return err
		}

		roleInterface := (*claims)["role"]
		userRole := models.UserRole(roleInterface.(string))

		fmt.Printf("User role from token: %s\n", userRole)
		fmt.Printf("Required roles: %v\n", roles)

		// ตรวจสอบ role
		for _, allowedRole := range roles {
			if userRole == allowedRole {
				return c.Next()
			}
		}

		return fiber.NewError(fiber.StatusForbidden,
			fmt.Sprintf("Access denied. Required roles: %v, Current role: %s", roles, userRole))
	}
}

func SetupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	err = db.AutoMigrate(&models.Table{}, &models.QRCode{})
	if err != nil {
		t.Fatalf("Failed to migrate test database: %v", err)
	}

	return db
}

func ConvertUintToString(id uint) string {
	return strconv.FormatUint(uint64(id), 10)
}

type APIKeyConfig struct {
	Key      string
	Type     string // ประเภทของ key เช่น "websocket", "api"
	IsActive bool
}

func InitAPIKeys() {
	APIKeyStore = map[string]APIKeyConfig{}

	// เพิ่ม key เฉพาะเมื่อมีค่าใน environment
	if tableKey := os.Getenv("WS_TABLE_KEY"); tableKey != "" {
		fmt.Printf("Initializing table key: %s\n", tableKey) // เพิ่ม log
		APIKeyStore[tableKey] = APIKeyConfig{
			Key:      tableKey,
			Type:     "websocket_table",
			IsActive: true,
		}
	} else {
		fmt.Println("Warning: WS_TABLE_KEY not found in environment") // เพิ่ม log
	}

	if printerKey := os.Getenv("WS_PRINTER_KEY"); printerKey != "" {
		fmt.Printf("Initializing printer key: %s\n", printerKey) // เพิ่ม log
		APIKeyStore[printerKey] = APIKeyConfig{
			Key:      printerKey,
			Type:     "websocket_printer",
			IsActive: true,
		}
	}

	fmt.Printf("Initialized API Key Store: %v\n", APIKeyStore) // เพิ่ม log
}

// APIKeyStore เก็บ API Keys ที่อนุญาต
var APIKeyStore map[string]APIKeyConfig

// ValidateAPIKey ตรวจสอบว่า API Key ถูกต้องไหม
func ValidateAPIKey(key string) bool {
	config, exists := APIKeyStore[key]
	return exists && config.IsActive
}

// ValidateAPIKeyByType ตรวจสอบ API Key ตามประเภท
func ValidateAPIKeyByType(key string, keyType string) bool {
	config, exists := APIKeyStore[key]
	return exists && config.IsActive && config.Type == keyType
}

// WebSocketAPIKeyMiddleware สำหรับ WebSocket โดยเฉพาะ
func WebSocketAPIKeyMiddleware(c *fiber.Ctx) error {
	apiKey := c.Get("X-API-Key")
	if apiKey == "" {
		apiKey = c.Query("api_key")
	}

	fmt.Printf("Received API Key: %s\n", apiKey)   // เพิ่ม log
	fmt.Printf("API Key Store: %v\n", APIKeyStore) // เพิ่ม log เพื่อดู APIKeyStore

	if !ValidateAPIKeyByType(apiKey, "websocket_table") { // แก้ไขประเภทให้ตรงกับที่กำหนดใน InitAPIKeys
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid WebSocket API Key",
		})
	}

	return c.Next()
}
