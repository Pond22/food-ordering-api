package utils

import (
	"food-ordering-api/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
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

	role := models.UserRole((*claims)["role"].(string))
	return role, nil
}

// CheckPermission ตรวจสอบว่าผู้ใช้มีสิทธิ์เข้าถึงหรือไม่
func CheckPermission(c *fiber.Ctx, allowedRoles ...models.UserRole) error {
	userRole, err := GetUserRole(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "Unauthorized access")
	}

	for _, role := range allowedRoles {
		if userRole == role {
			return nil
		}
	}

	return fiber.NewError(fiber.StatusForbidden, "Permission denied")
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
		return CheckPermission(c, roles...)
	}
}