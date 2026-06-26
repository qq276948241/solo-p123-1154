package handlers

import (
	"net/http"
	"pet-clinic-backend/common"
	"pet-clinic-backend/database"
	"pet-clinic-backend/middleware"
	"pet-clinic-backend/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type RegisterRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required,min=6"`
	Username string `json:"username" binding:"required"`
}

type LoginRequest struct {
	Phone    string `json:"phone" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	UserID   uint   `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var existing models.User
	result := database.DB.Where("phone = ?", req.Phone).First(&existing)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrUserExists))
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrInternalServer))
		return
	}

	user := models.User{
		Phone:    req.Phone,
		Password: string(hashedPassword),
		Username: req.Username,
		Role:     models.RoleUser,
	}

	if err := database.DB.Create(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Phone, user.Role, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(LoginResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}))
}

func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var user models.User
	result := database.DB.Where("phone = ?", req.Phone).First(&user)
	if result.Error != nil {
		c.JSON(http.StatusUnauthorized, common.ErrorResponse(common.ErrUserNotFound))
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, common.ErrorResponse(common.ErrWrongPassword))
		return
	}

	token, err := middleware.GenerateToken(user.ID, user.Phone, user.Role, user.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrInternalServer))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(LoginResponse{
		Token:    token,
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
	}))
}

func GetCurrentUser(c *gin.Context) {
	userID, _ := c.Get("userID")

	var user models.User
	if err := database.DB.First(&user, userID).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrUserNotFound))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(user))
}
