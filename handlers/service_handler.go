package handlers

import (
	"net/http"
	"pet-clinic-backend/common"
	"pet-clinic-backend/database"
	"pet-clinic-backend/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateServiceRequest struct {
	Name        string  `json:"name" binding:"required"`
	Category    string  `json:"category" binding:"required"`
	Description string  `json:"description"`
	Price       float64 `json:"price" binding:"required,min=0"`
	Duration    int     `json:"duration" binding:"required,min=1"`
}

type UpdateServiceRequest struct {
	Name        string   `json:"name"`
	Category    string   `json:"category"`
	Description string   `json:"description"`
	Price       *float64 `json:"price"`
	Duration    *int     `json:"duration"`
	IsActive    *bool    `json:"is_active"`
}

type CreateDoctorRequest struct {
	UserID      uint   `json:"user_id" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Title       string `json:"title"`
	Specialty   string `json:"specialty"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
}

type UpdateDoctorRequest struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Specialty   string `json:"specialty"`
	Description string `json:"description"`
	Avatar      string `json:"avatar"`
	IsActive    *bool  `json:"is_active"`
}

func CreateService(c *gin.Context) {
	var req CreateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	service := models.Service{
		Name:        req.Name,
		Category:    req.Category,
		Description: req.Description,
		Price:       req.Price,
		Duration:    req.Duration,
		IsActive:    true,
	}

	if err := database.DB.Create(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(service))
}

func GetService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var service models.Service
	if err := database.DB.First(&service, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrServiceNotFound))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(service))
}

func GetAllServices(c *gin.Context) {
	category := c.Query("category")
	active := c.Query("active")

	query := database.DB.Model(&models.Service{})

	if category != "" {
		query = query.Where("category = ?", category)
	}
	if active != "" {
		isActive := active == "true"
		query = query.Where("is_active = ?", isActive)
	}

	var services []models.Service
	if err := query.Order("created_at DESC").Find(&services).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(services))
}

func GetServiceCategories(c *gin.Context) {
	var categories []string
	if err := database.DB.Model(&models.Service{}).
		Where("is_active = ?", true).
		Distinct("category").
		Pluck("category", &categories).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(categories))
}

func UpdateService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var req UpdateServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var service models.Service
	if err := database.DB.First(&service, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrServiceNotFound))
		return
	}

	if req.Name != "" {
		service.Name = req.Name
	}
	if req.Category != "" {
		service.Category = req.Category
	}
	if req.Description != "" {
		service.Description = req.Description
	}
	if req.Price != nil {
		service.Price = *req.Price
	}
	if req.Duration != nil {
		service.Duration = *req.Duration
	}
	if req.IsActive != nil {
		service.IsActive = *req.IsActive
	}

	if err := database.DB.Save(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(service))
}

func DeleteService(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var service models.Service
	if err := database.DB.First(&service, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrServiceNotFound))
		return
	}

	if err := database.DB.Delete(&service).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(nil))
}

func CreateDoctor(c *gin.Context) {
	var req CreateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var user models.User
	if err := database.DB.First(&user, req.UserID).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrUserNotFound))
		return
	}

	user.Role = models.RoleDoctor
	database.DB.Save(&user)

	doctor := models.Doctor{
		UserID:      req.UserID,
		Name:        req.Name,
		Title:       req.Title,
		Specialty:   req.Specialty,
		Description: req.Description,
		Avatar:      req.Avatar,
		IsActive:    true,
	}

	if err := database.DB.Create(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(doctor))
}

func GetDoctor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var doctor models.Doctor
	if err := database.DB.First(&doctor, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrDoctorNotFound))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(doctor))
}

func GetAllDoctors(c *gin.Context) {
	active := c.Query("active")

	query := database.DB.Model(&models.Doctor{})

	if active != "" {
		isActive := active == "true"
		query = query.Where("is_active = ?", isActive)
	}

	var doctors []models.Doctor
	if err := query.Order("created_at DESC").Find(&doctors).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(doctors))
}

func UpdateDoctor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var req UpdateDoctorRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var doctor models.Doctor
	if err := database.DB.First(&doctor, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrDoctorNotFound))
		return
	}

	if req.Name != "" {
		doctor.Name = req.Name
	}
	if req.Title != "" {
		doctor.Title = req.Title
	}
	if req.Specialty != "" {
		doctor.Specialty = req.Specialty
	}
	if req.Description != "" {
		doctor.Description = req.Description
	}
	if req.Avatar != "" {
		doctor.Avatar = req.Avatar
	}
	if req.IsActive != nil {
		doctor.IsActive = *req.IsActive
	}

	if err := database.DB.Save(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(doctor))
}

func DeleteDoctor(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var doctor models.Doctor
	if err := database.DB.First(&doctor, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrDoctorNotFound))
		return
	}

	if err := database.DB.Delete(&doctor).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(nil))
}
