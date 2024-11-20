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

type CreateMenuRequest struct {
	Name        string `json:"Name" binding:"required,min=2,max=100"`
	Description string `json:"description" binding:"max=255"`
	Image       []byte `json:"image" binding:"max=5242880"` // max 5MB
	CategoryID  uint   `json:"CategoryID" binding:"required,min=1"`
	Price       int    `json:"Price" binding:"required,min=1"`
}
