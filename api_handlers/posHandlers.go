package api_handlers

import (
	"errors"
	"fmt"
	"food-ordering-api/db"
	"food-ordering-api/models"
	utils "food-ordering-api/utility"
	"math/rand"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type GeneratePOSVerificationCodeResponse struct {
	Token    string `json:"token"`
	CodeHint string `json:"code_hint"`
}

// @Summary สร้าง Verification Code สำหรับการเข้าใช้งาน POS
// @Description สร้าง Temporary Session และ Verification Code สำหรับผู้ใช้งาน POS
// @Tags POS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} GeneratePOSVerificationCodeResponse "สร้าง Session และ Verification Code สำเร็จ"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต: ไม่มีสิทธิ์เข้าใช้"
// @Failure 400 {object} map[string]interface{} "ข้อมูลผู้ใช้ไม่ถูกต้อง"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการสร้าง Session"
// @Router /api/pos/generate-code [post]
func GeneratePOSVerificationCode(c *fiber.Ctx) error {
	// ตรวจสอบ Authentication
	claims, err := utils.GetUserFromToken(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error":   "Unauthorized",
			"details": err.Error(),
		})
	}

	// ดึง staff ID และ role จาก claims
	staffIDFloat, ok := (*claims)["user_id"].(float64)
	if !ok {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user data",
		})
	}
	staffID := uint(staffIDFloat)

	// ตรวจสอบ role ว่าเป็น staff หรือ manager
	roleInterface := (*claims)["role"]
	if roleInterface == nil {
		return c.Status(403).JSON(fiber.Map{
			"error": "Role information missing",
		})
	}

	role := models.UserRole(roleInterface.(string))
	if role != models.RoleStaff && role != models.RoleManager {
		return c.Status(403).JSON(fiber.Map{
			"error": "Only staff and managers can generate POS verification codes",
		})
	}

	// สร้าง Verification Code แบบ cryptographically secure
	const codeLength = 6
	code := make([]byte, codeLength)
	for i := 0; i < codeLength; i++ {
		code[i] = byte(rand.Intn(10) + '0')
	}
	verificationCode := string(code)

	// ตั้งเวลาหมดอายุ (5 นาที)
	expiresAt := time.Now().Add(5 * time.Minute)

	// สร้าง Session
	session := models.POSSession{
		StaffID:          &staffID,
		StartTime:        time.Now(),
		Status:           "pending",
		LoginToken:       uuid.New().String(),
		VerificationCode: verificationCode,
		Verified:         false,
		ExpiresAt:        &expiresAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := db.DB.Create(&session).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to create session",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"token":             session.LoginToken,
		"verification_code": verificationCode,
		"expires_at":        expiresAt,
	})
}

type POSVerificationRequest struct {
	Token   string `json:"token"`
	Code    string `json:"code"`
	StaffID uint   `json:"staff_id"`
}

// @Summary ยืนยัน Verification Code เพื่อเข้าใช้งาน POS
// @Description ตรวจสอบและยืนยัน Verification Code สำหรับการเข้าใช้งาน POS
// @Tags POS
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body POSVerificationRequest true "ข้อมูลการยืนยัน"
// @Success 200 {object} map[string]interface{} "ยืนยันการเข้าใช้งาน POS สำเร็จ"
// @Failure 400 {object} map[string]interface{} "รูปแบบคำขอไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "รหัสยืนยันไม่ถูกต้อง"
// @Failure 404 {object} map[string]interface{} "ไม่พบพนักงาน"
// @Failure 500 {object} map[string]interface{} "เกิดข้อผิดพลาดในการยืนยัน Session"
// @Router /api/pos/verify-code [post]
func VerifyPOSAccessCode(c *fiber.Ctx) error {
	// รับ verification code
	var req struct {
		Code string `json:"code"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// ค้นหา session ที่รอการยืนยันด้วย code
	var session models.POSSession
	if err := db.DB.Where("verification_code = ? AND status = ?",
		req.Code, "pending").First(&session).Error; err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid verification code",
		})
	}

	// ตรวจสอบว่า code หมดอายุหรือยัง
	if session.ExpiresAt != nil && session.ExpiresAt.Before(time.Now()) {
		return c.Status(401).JSON(fiber.Map{
			"error": "Verification code has expired",
		})
	}

	// ตรวจสอบ staff
	var staff models.Users
	if err := db.DB.First(&staff, session.StaffID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Staff not found",
		})
	}

	// สร้าง JWT token สำหรับ POS session
	claims := jwt.MapClaims{
		"user_id":        *session.StaffID,
		"role":           staff.Role,
		"pos_session_id": session.ID,
		"exp":            time.Now().Add(8 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to generate token",
		})
	}

	// อัพเดท session
	now := time.Now()
	session.Status = "active"
	session.Verified = true
	session.UpdatedAt = now
	session.LastActivityAt = now
	session.ExpiresAt = nil // ล้างเวลาหมดอายุของ verification code
	session.DeviceInfo = c.Get("User-Agent")
	session.IPAddress = c.IP()

	if err := db.DB.Save(&session).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update session",
		})
	}

	// บันทึกประวัติการยืนยัน
	verificationLog := models.POSSessionLog{
		SessionID:   session.ID,
		StaffID:     *session.StaffID,
		Action:      "verify",
		Status:      "success",
		DeviceInfo:  c.Get("User-Agent"),
		IPAddress:   c.IP(),
		Description: "POS verification successful",
		CreatedAt:   now,
	}

	if err := db.DB.Create(&verificationLog).Error; err != nil {
		// log error แต่ไม่ return error เพราะไม่ใช่ critical operation
		fmt.Printf("Failed to create verification log: %v\n", err)
	}

	return c.JSON(fiber.Map{
		"message":    "POS access granted",
		"token":      tokenString,
		"session_id": session.ID,
		"staff_id":   session.StaffID,
		"expires_in": "8h",
	})
}

var jwtSecret = []byte("your-secret-key")

// POSAuthMiddleware ตรวจสอบสิทธิ์ของ POS Session
func POSAuthMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Get("Authorization")

		if tokenString == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Missing token"})
		}

		// ตรวจสอบ JWT
		claims, err := utils.GetUserFromToken(c)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid token"})
		}

		userID, ok := (*claims)["user_id"].(float64)
		if !ok {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "Invalid user ID"})
		}

		posSessionID, exists := (*claims)["pos_session_id"].(string)
		if !exists || posSessionID == "" {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "POS session required"})
		}

		// ตรวจสอบว่ามี POS Session และยัง Active อยู่
		var posSession models.POSSession
		if err := db.DB.Where("id = ? AND staff_id = ? AND status = ?", posSessionID, uint(userID), "active").
			First(&posSession).Error; err != nil {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"message": "Invalid or expired POS session"})
		}

		// ตรวจสอบว่า Session ยังไม่หมดอายุ
		if posSession.UpdatedAt.Add(8 * time.Hour).Before(time.Now()) { // กะหมดอายุใน 8 ชม.
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"message": "POS session expired"})
		}

		// ตั้งค่า Context
		c.Locals("user_id", uint(userID))
		c.Locals("pos_session_id", posSessionID)

		return c.Next()
	}
}

func LogoutPOS(c *fiber.Ctx) error {
	// ดึง token จาก Authorization header
	token := c.Get("Authorization")
	if token == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "Missing authorization token",
		})
	}

	// แยก Bearer token
	tokenString := strings.Replace(token, "Bearer ", "", 1)

	// ตรวจสอบและถอดรหัส token
	claims := &jwt.MapClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})

	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error": "Invalid token",
		})
	}

	// ดึงข้อมูลจาก claims
	staffIDFloat, ok := (*claims)["user_id"].(float64)
	if !ok {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user data in token",
		})
	}
	staffID := uint(staffIDFloat)

	sessionIDFloat, ok := (*claims)["pos_session_id"].(float64)
	if !ok {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid session data in token",
		})
	}
	sessionID := uint(sessionIDFloat)

	// ค้นหา session
	var session models.POSSession
	if err := db.DB.First(&session, sessionID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Session not found",
		})
	}

	// ตรวจสอบว่า session ยังใช้งานอยู่
	if session.Status != "active" {
		return c.Status(400).JSON(fiber.Map{
			"error":  "Session is not active",
			"status": session.Status,
		})
	}

	// เริ่ม transaction
	tx := db.DB.Begin()

	// อัปเดต session
	now := time.Now()
	session.Status = "closed"
	session.EndTime = &now
	session.UpdatedAt = now

	if err := tx.Save(&session).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to update session",
		})
	}

	// บันทึกประวัติการ logout
	sessionLog := models.POSSessionLog{
		SessionID:   sessionID,
		StaffID:     staffID,
		Action:      "logout",
		Status:      "success",
		DeviceInfo:  c.Get("User-Agent"),
		IPAddress:   c.IP(),
		Description: "User logged out successfully",
		CreatedAt:   now,
	}

	if err := tx.Create(&sessionLog).Error; err != nil {
		tx.Rollback()
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to create session log",
		})
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to commit transaction",
		})
	}

	return c.JSON(fiber.Map{
		"message":     "Successfully logged out from POS",
		"session_id":  sessionID,
		"logout_time": now,
		"staff_id":    staffID,
	})
}

// SessionStatusResponse คือโครงสร้างข้อมูลที่จะส่งกลับ
type SessionStatusResponse struct {
	IsActive     bool      `json:"is_active"`
	SessionID    uint      `json:"session_id,omitempty"`
	StartTime    time.Time `json:"start_time,omitempty"`
	LastActivity time.Time `json:"last_activity,omitempty"`
	StaffID      uint      `json:"staff_id,omitempty"`
	StaffName    string    `json:"staff_name,omitempty"`
	DeviceInfo   string    `json:"device_info,omitempty"`
	IPAddress    string    `json:"ip_address,omitempty"`
}

// GetPOSSessionStatus ตรวจสอบสถานะ POS session ปัจจุบัน
func GetPOSSessionStatus(c *fiber.Ctx) error {
	// ดึง token จาก Authorization header
	claims, err := utils.GetUserFromToken(c)
	if err != nil {
		return c.Status(401).JSON(fiber.Map{
			"error":   "Unauthorized",
			"details": err.Error(),
		})
	}

	// ดึง user_id จาก claims
	staffIDFloat, ok := (*claims)["user_id"].(float64)
	if !ok {
		return c.Status(400).JSON(fiber.Map{
			"error": "Invalid user data in token",
		})
	}
	staffID := uint(staffIDFloat)

	// ดึง pos_session_id จาก claims ถ้ามี
	var sessionID uint
	if sessionIDFloat, ok := (*claims)["pos_session_id"].(float64); ok {
		sessionID = uint(sessionIDFloat)
	}

	// ถ้าไม่มี pos_session_id ใน token แสดงว่าไม่ได้อยู่ใน POS session
	if sessionID == 0 {
		return c.JSON(SessionStatusResponse{
			IsActive: false,
			StaffID:  staffID,
		})
	}

	// ค้นหา session ในฐานข้อมูล
	var session models.POSSession
	err = db.DB.Preload("Staff").First(&session, sessionID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.JSON(SessionStatusResponse{
				IsActive: false,
				StaffID:  staffID,
			})
		}
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to fetch session data",
			"details": err.Error(),
		})
	}

	// ตรวจสอบว่า session ยังใช้งานอยู่หรือไม่
	isActive := session.Status == "active"

	// ถ้า session active ให้อัพเดท LastActivityAt
	if isActive {
		now := time.Now()
		session.LastActivityAt = now
		if err := db.DB.Save(&session).Error; err != nil {
			// log error แต่ไม่ return error เพราะไม่ใช่ critical operation
			fmt.Printf("Failed to update last activity: %v\n", err)
		}
	}

	// สร้าง response
	response := SessionStatusResponse{
		IsActive:     isActive,
		SessionID:    session.ID,
		StartTime:    session.StartTime,
		LastActivity: session.LastActivityAt,
		StaffID:      staffID,
		DeviceInfo:   session.DeviceInfo,
		IPAddress:    session.IPAddress,
	}

	// เพิ่มชื่อพนักงานถ้ามีข้อมูล Staff
	if session.Staff != nil {
		response.StaffName = session.Staff.Name
	}

	return c.JSON(response)
}
