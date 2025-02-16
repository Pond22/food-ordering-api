package api_handlers

import (
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// LoginRequest สำหรับรับข้อมูล login
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse สำหรับส่ง login กลับ
type LoginResponse struct {
	Token     string      `json:"token"`
	ExpiresAt int64       `json:"expires_at"`
	User      UserProfile `json:"user"`
}

// UserProfile สำหรับส่งข้อมูลผู้ใช้กลับ
type UserProfile struct {
	ID       uint            `json:"id"`
	Username string          `json:"username"`
	Name     string          `json:"name"`
	Role     models.UserRole `json:"role"`
}

// GetJWTSecretKey ดึง secret key สำหรับ JWT
func GetJWTSecretKey() []byte {
	secret := os.Getenv("JWT_SECRET_KEY")
	return []byte(secret)
}

// @Summary เข้าสู่ระบบ
// @Description ล็อกอินเข้าสู่ระบบโดยใช้ username และ password
// @Accept json
// @Produce json
// @Param request body LoginRequest true "ข้อมูลสำหรับเข้าสู่ระบบ (username, password)"
// @Success 200 {object} LoginResponse "ข้อมูลการเข้าสู่ระบบและ token"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง"
// @Router /api/auth/login [post]
// @Tags authentication
func Login(c *fiber.Ctx) error {
	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// ตรวจสอบว่ามีการส่งข้อมูลครบถ้วน
	if req.Username == "" || req.Password == "" {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Username and password are required",
		})
	}

	// ค้นหาผู้ใช้จากฐานข้อมูล
	var user models.Users
	if err := db.DB.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	// ตรวจสอบรหัสผ่าน
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid username or password",
		})
	}

	// สร้าง JWT token
	expirationTime := time.Now().Add(24 * time.Hour) // token หมดอายุใน 24 ชั่วโมง
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      expirationTime.Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(GetJWTSecretKey())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Could not generate token",
		})
	}

	// สร้าง response
	response := LoginResponse{
		Token:     tokenString,
		ExpiresAt: expirationTime.Unix(),
		User: UserProfile{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
			Role:     user.Role,
		},
	}

	return c.JSON(response)
}

// @Summary ตรวจสอบ token
// @Security BearerAuth
// @Description ตรวจสอบความถูกต้องของ token และส่งข้อมูลผู้ใช้กลับ โดยส่ง token ในรูปแบบ "Bearer <token>"
// @Produce json
// @Success 200 {object} UserProfile "ข้อมูลผู้ใช้"
// @Failure 401 {object} map[string]interface{} "Token ไม่ถูกต้องหรือหมดอายุ"
// @Router /api/auth/verify-token [get]
// @Tags authentication
func VerifyToken(c *fiber.Ctx) error {
	// ดึง token จาก Authorization header
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Missing authorization token",
		})
	}

	// แยก Bearer token
	tokenString := strings.Replace(token, "Bearer ", "", 1)

	// ตรวจสอบและถอดรหัส token
	claims := jwt.MapClaims{}
	parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return GetJWTSecretKey(), nil
	})

	if err != nil || !parsedToken.Valid {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Invalid or expired token",
		})
	}

	// ดึงข้อมูลผู้ใช้จาก claims
	userID := uint(claims["user_id"].(float64))
	var user models.Users
	if err := db.DB.Select("id, username, name, role").First(&user, userID).Error; err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// ส่งข้อมูลผู้ใช้กลับไป
	profile := UserProfile{
		ID:       user.ID,
		Username: user.Username,
		Name:     user.Name,
		Role:     user.Role,
	}

	return c.JSON(profile)
}

// สร้าง middleware สำหรับตรวจสอบ token
func AuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// ดึง token จาก Authorization header
		token := c.Get("Authorization")
		if token == "" {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Missing authorization token",
			})
		}

		// แยก Bearer token
		tokenString := strings.Replace(token, "Bearer ", "", 1)

		// ตรวจสอบและถอดรหัส token
		claims := jwt.MapClaims{}
		parsedToken, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return GetJWTSecretKey(), nil
		})

		if err != nil || !parsedToken.Valid {
			return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
				"error": "Invalid or expired token",
			})
		}

		// เพิ่มข้อมูลผู้ใช้ลงใน context
		c.Locals("user", claims)

		return c.Next()
	}
}
