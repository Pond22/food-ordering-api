package utils

import (
	"errors"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

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

// IsOwner ตรวจสอบว่าผู้ใช้เป็น owner หรือไม่
func IsOwner(c *fiber.Ctx) bool {
	claims, err := GetUserFromToken(c)
	if err != nil {
		return false
	}

	role := (*claims)["role"].(string)
	return role == string(models.RoleOwner)
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
	fmt.Println("\n=== Starting API Key Initialization ===")
	APIKeyStore = map[string]APIKeyConfig{}

	// เพิ่ม Table Key
	if tableKey := os.Getenv("WS_TABLE_KEY"); tableKey != "" {
		APIKeyStore[tableKey] = APIKeyConfig{
			Key:      tableKey,
			Type:     "websocket_table",
			IsActive: true,
		}
	} else {
		fmt.Println("\nWarning: WS_TABLE_KEY not found in environment")
	}

	// เพิ่ม Printer Key
	if printerKey := os.Getenv("WS_PRINTER_KEY"); printerKey != "" {
		APIKeyStore[printerKey] = APIKeyConfig{
			Key:      printerKey,
			Type:     "websocket_printer",
			IsActive: true,
		}
	} else {
		fmt.Println("\nWarning: WS_PRINTER_KEY not found in environment")
	}

	// เพิ่ม Staff Key
	if staffKey := os.Getenv("WS_STAFF_KEY"); staffKey != "" {
		fmt.Printf("\nAdding staff key: %s\n", staffKey)
		APIKeyStore[staffKey] = APIKeyConfig{
			Key:      staffKey,
			Type:     "websocket_staff",
			IsActive: true,
		}
		// ตรวจสอบว่า key ถูกเพิ่มเข้าไปจริงๆ
		if _, exists := APIKeyStore[staffKey]; exists {
			fmt.Printf("Successfully added staff key to store\n")
		} else {
			fmt.Printf("Failed to add staff key to store!\n")
		}
	} else {
		fmt.Println("\nWarning: WS_STAFF_KEY not found in environment")
	}

	for key, config := range APIKeyStore {
		fmt.Printf("Key: %s, Type: %s, Active: %v\n", key, config.Type, config.IsActive)
	}
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
func WebSocketAPIKeyMiddleware(wsType string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		fmt.Printf("\n=== WebSocket Connection Attempt ===\n")

		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		fmt.Printf("Received request:\n")
		fmt.Printf("- API Key: %s\n", apiKey)
		fmt.Printf("- WebSocket Type: %s\n", wsType)

		fmt.Printf("\nCurrent API Key Store:\n")
		for key, config := range APIKeyStore {
			fmt.Printf("- Key: %s, Type: %s, Active: %v\n", key, config.Type, config.IsActive)
		}

		if !ValidateAPIKeyByType(apiKey, wsType) {
			fmt.Printf("\nValidation Failed!\n")
			fmt.Printf("- Received Key: %s\n", apiKey)
			fmt.Printf("- Required Type: %s\n", wsType)
			fmt.Println("=== Connection Rejected ===\n")

			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": fmt.Sprintf("Invalid %s API Key", wsType),
			})
		}

		fmt.Println("=== Connection Accepted ===\n")
		return c.Next()
	}
}

// PrinterAPIKeyMiddleware สำหรับ Printer req แบบ http
func PrinterAPIKeyMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		apiKey := c.Get("X-API-Key")
		if apiKey == "" {
			apiKey = c.Query("api_key")
		}

		fmt.Printf("Received Printer API Key: %s\n", apiKey)

		if !ValidateAPIKeyByType(apiKey, "websocket_printer") {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid Printer API Key",
			})
		}

		return c.Next()
	}
}

// GetPOSSecretKey ดึง secret key สำหรับ POS
func GetPOSSecretKey() []byte {
	posSecret := os.Getenv("POS_JWT_SECRET")
	return []byte(posSecret)
}

// GetUserFromPOSToken ดึงข้อมูลผู้ใช้จาก POS JWT token
func GetUserFromPOSToken(c *fiber.Ctx) (*jwt.MapClaims, error) {
	// ดึง token จาก Authorization header
	token := c.Get("Authorization")
	if token == "" {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Missing authorization token")
	}

	// แยก Bearer token
	tokenString := strings.Replace(token, "Bearer ", "", 1)

	// ตรวจสอบและถอดรหัส token ด้วย POS secret key
	claims := &jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return GetPOSSecretKey(), nil
	})

	if err != nil {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid POS token")
	}

	// ตรวจสอบว่ามี pos_session_id
	if _, exists := (*claims)["pos_session_id"]; !exists {
		return nil, fiber.NewError(fiber.StatusUnauthorized, "Invalid POS token: missing session ID")
	}

	return claims, nil
}

// POSAuthRequired middleware สำหรับตรวจสอบการ authentication ของ POS
func POSAuthRequired() fiber.Handler {
	return func(c *fiber.Ctx) error {
		claims, err := GetUserFromPOSToken(c)
		if err != nil {
			return err
		}

		// ตรวจสอบว่า session ยังใช้งานได้
		sessionIDFloat, ok := (*claims)["pos_session_id"].(float64)
		if !ok {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid session ID format")
		}

		var session models.POSSession
		if err := db.DB.Where("id = ? AND status = ?", uint(sessionIDFloat), "active").First(&session).Error; err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "Invalid or expired POS session")
		}

		// ตรวจสอบเวลาหมดอายุของ session
		if session.UpdatedAt.Add(8 * time.Hour).Before(time.Now()) {
			return fiber.NewError(fiber.StatusUnauthorized, "POS session expired")
		}

		return c.Next()
	}
}
