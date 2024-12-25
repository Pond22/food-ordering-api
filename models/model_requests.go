package models

type CreateUserRequest struct {
	Username string   `json:"username" binding:"required,min=3"`
	Password string   `json:"password" binding:"required,min=6"`
	Role     UserRole `json:"role" binding:"required"`
	Name     string   `json:"name" binding:"required"`
}

type CreateCategoryRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateMenuRequest - รับข้อมูลสำหรับสร้างเมนูใหม่พร้อม options
type CreateMenuRequest struct {
	MenuItem     MenuItemRequest      `json:"menu_item"`
	OptionGroups []OptionGroupRequest `json:"option_groups"`
}

type MenuItemRequest struct {
	Name          string `json:"name" binding:"required"`
	NameEn        string `json:"name_en" binding:"required"`
	NameCh        string `json:"name_ch" binding:"required"`
	Description   string `json:"description"`
	DescriptionEn string `json:"description_en"`
	DescriptionCh string `json:"description_ch"`
	Image         []byte `json:"image"`
	CategoryID    uint   `json:"category_id" binding:"required"`
	Price         int16  `json:"price" binding:"required"`
}

type OptionRequest struct {
	Name   string  `json:"name" binding:"required"`
	NameEn string  `json:"name_en" binding:"required"`
	NameCh string  `json:"name_ch" binding:"required"`
	Price  float64 `json:"price"`
}

type OptionGroupRequest struct {
	Name          string          `json:"name" binding:"required"`
	NameEn        string          `json:"name_en" binding:"required"`
	NameCh        string          `json:"name_ch" binding:"required"`
	MaxSelections int             `json:"MaxSelections"`
	IsRequired    bool            `json:"is_required"`
	Options       []OptionRequest `json:"options"`
}