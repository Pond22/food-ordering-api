package api_handlers

import (
	"food-ordering-api/db"
	"food-ordering-api/models"
	"net/http"
	"strconv"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type ResetPasswordRequest struct {
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=6"`
}

// @Summary ดูรายชื่อผู้ใช้ทั้งหมด
// @Description ดึงรายชื่อผู้ใช้ทั้งหมดในระบบ (เฉพาะ manager เท่านั้น)
// @Produce json
// @Security BearerAuth
// @Success 200 {array} UserProfile "รายการผู้ใช้ทั้งหมด"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Router /api/member [get]
// @Tags users
func GetUsers(c *fiber.Ctx) error {
	var users []models.Users
	if err := db.DB.Select("id, username, name, role, created_at").Find(&users).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to fetch users",
		})
	}

	var profiles []UserProfile
	for _, user := range users {
		profiles = append(profiles, UserProfile{
			ID:       user.ID,
			Username: user.Username,
			Name:     user.Name,
			Role:     user.Role,
		})
	}

	return c.JSON(profiles)
}

// @Summary สร้างผู้ใช้ใหม่
// @Description สร้างผู้ใช้ใหม่ในระบบ (เฉพาะ manager เท่านั้น)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body models.CreateUserRequest true "ข้อมูลผู้ใช้ใหม่"
// @Success 201 {object} UserProfile "ข้อมูลผู้ใช้ที่สร้าง"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 409 {object} map[string]interface{} "ชื่อผู้ใช้ซ้ำ"
// @Router /api/member [post]
// @Tags users
func CreateUser(c *fiber.Ctx) error {
	var req models.CreateUserRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// ตรวจสอบข้อมูล
	if len(req.Username) < 3 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Username must be at least 3 characters",
		})
	}
	if len(req.Password) < 6 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Password must be at least 6 characters",
		})
	}
	if !req.Role.IsValid() {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid role",
		})
	}

	// ตรวจสอบชื่อผู้ใช้ซ้ำ
	var existingUser models.Users
	if err := db.DB.Where("username = ?", req.Username).First(&existingUser).Error; err == nil {
		return c.Status(http.StatusConflict).JSON(fiber.Map{
			"error": "Username already exists",
		})
	}

	// เข้ารหัสรหัสผ่าน
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	// สร้างผู้ใช้ใหม่
	user := models.Users{
		Username: req.Username,
		Password: string(hashedPassword),
		Role:     req.Role,
		Name:     req.Name,
	}

	if err := db.DB.Create(&user).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to create user",
		})
	}

	return c.Status(http.StatusCreated).JSON(UserProfile{
		ID:       user.ID,
		Username: user.Username,
		Name:     user.Name,
		Role:     user.Role,
	})
}

// @Summary ดูข้อมูลผู้ใช้
// @Description ดูข้อมูลผู้ใช้ของตัวเอง
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UserProfile "ข้อมูลผู้ใช้"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Router /api/member/get_member_profile [get]
// @Tags users
func GetUserProfile(c *fiber.Ctx) error {
	claims := c.Locals("user").(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var user models.Users
	if err := db.DB.Select("id, username, name, role").First(&user, userID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	return c.JSON(UserProfile{
		ID:       user.ID,
		Username: user.Username,
		Name:     user.Name,
		Role:     user.Role,
	})
}

// @Summary เปลี่ยนรหัสผ่าน
// @Description เปลี่ยนรหัสผ่านของตัวเอง
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body ChangePasswordRequest true "รหัสผ่านเก่าและใหม่"
// @Success 200 {object} map[string]interface{} "เปลี่ยนรหัสผ่านสำเร็จ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาตหรือรหัสผ่านเก่าไม่ถูกต้อง"
// @Router /api/member/change-password [put]
// @Tags users
func ChangePassword(c *fiber.Ctx) error {

	var req ChangePasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	// ตรวจสอบความยาวรหัสผ่านใหม่
	if len(req.NewPassword) < 6 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "New password must be at least 6 characters",
		})
	}

	claims := c.Locals("user").(jwt.MapClaims)
	userID := uint(claims["user_id"].(float64))

	var user models.Users
	if err := db.DB.First(&user, userID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	// ตรวจสอบรหัสผ่านเก่า
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.CurrentPassword)); err != nil {
		return c.Status(http.StatusUnauthorized).JSON(fiber.Map{
			"error": "Current password is incorrect",
		})
	}

	// เข้ารหัสและบันทึกรหัสผ่านใหม่
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	if err := db.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
	})
}

// @Summary รีเซ็ตรหัสผ่านผู้ใช้
// @Description รีเซ็ตรหัสผ่านของผู้ใช้ (เฉพาะ manager เท่านั้น)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID ของผู้ใช้"
// @Param request body ResetPasswordRequest true "รหัสผ่านใหม่"
// @Success 200 {object} map[string]interface{} "รีเซ็ตรหัสผ่านสำเร็จ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 404 {object} map[string]interface{} "ไม่พบผู้ใช้"
// @Router /api/member/{id}/reset-password [put]
// @Tags users
func ResetUserPassword(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	var req struct {
		NewPassword string `json:"new_password"`
	}

	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request format",
		})
	}

	if len(req.NewPassword) < 6 {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "New password must be at least 6 characters",
		})
	}

	var user models.Users
	if err := db.DB.First(&user, userID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Error hashing password",
		})
	}

	if err := db.DB.Model(&user).Update("password", string(hashedPassword)).Error; err != nil {
		return c.Status(http.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to update password",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Password reset successfully",
	})
}

// @Summary ลบผู้ใช้
// @Description ลบผู้ใช้ (เฉพาะ manager เท่านั้น)
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID ของผู้ใช้"
// @Success 200 {object} map[string]interface{} "ลบผู้ใช้สำเร็จ"
// @Failure 400 {object} map[string]interface{} "ข้อมูลไม่ถูกต้อง"
// @Failure 401 {object} map[string]interface{} "ไม่ได้รับอนุญาต"
// @Failure 403 {object} map[string]interface{} "ไม่มีสิทธิ์เข้าถึง"
// @Failure 404 {object} map[string]interface{} "ไม่พบผู้ใช้"
// @Router /api/member/{id}/delete-member [delete]
// @Tags users
func Delete_user(c *fiber.Ctx) error {
	userID, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid user ID",
		})
	}

	// ป้องกันการลบตัวเอง
	// claims := c.Locals("user").(jwt.MapClaims)
	// if userID == int(claims["user_id"].(float64)) {
	// 	return c.Status(http.StatusForbidden).JSON(fiber.Map{
	// 		"error": "Cannot delete your own account",
	// 	})
	// }

	var user models.Users
	if err := db.DB.First(&user, userID).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "User not found",
		})
	}

	if err := db.DB.Delete(&user).Error; err != nil {
		return c.Status(http.StatusNotFound).JSON(fiber.Map{
			"error": "Can't Delete that user",
		})
	}

	return c.JSON(fiber.Map{
		"message": "Delete User successfully",
	})
}