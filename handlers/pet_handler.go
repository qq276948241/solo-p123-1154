package handlers

import (
	"net/http"
	"pet-clinic-backend/common"
	"pet-clinic-backend/database"
	"pet-clinic-backend/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreatePetRequest struct {
	Name    string `json:"name" binding:"required"`
	Species string `json:"species" binding:"required"`
	Age     int    `json:"age"`
	Gender  string `json:"gender"`
	Breed   string `json:"breed"`
	Note    string `json:"note"`
}

type UpdatePetRequest struct {
	Name    string `json:"name"`
	Species string `json:"species"`
	Age     *int   `json:"age"`
	Gender  string `json:"gender"`
	Breed   string `json:"breed"`
	Note    string `json:"note"`
}

func CreatePet(c *gin.Context) {
	var req CreatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	userID, _ := c.Get("userID")

	pet := models.Pet{
		OwnerID: userID.(uint),
		Name:    req.Name,
		Species: req.Species,
		Age:     req.Age,
		Gender:  req.Gender,
		Breed:   req.Breed,
		Note:    req.Note,
	}

	if err := database.DB.Create(&pet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(pet))
}

func GetMyPets(c *gin.Context) {
	userID, _ := c.Get("userID")

	var pets []models.Pet
	if err := database.DB.Where("owner_id = ?", userID.(uint)).
		Order("created_at DESC").Find(&pets).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(pets))
}

func GetPet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	userID, _ := c.Get("userID")

	var pet models.Pet
	if err := database.DB.First(&pet, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponseWithMsg(common.ErrNotFound, "宠物不存在"))
		return
	}

	if pet.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, common.ErrorResponseWithMsg(common.ErrForbidden, "无权操作该宠物"))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(pet))
}

func UpdatePet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	userID, _ := c.Get("userID")

	var pet models.Pet
	if err := database.DB.First(&pet, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponseWithMsg(common.ErrNotFound, "宠物不存在"))
		return
	}

	if pet.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, common.ErrorResponseWithMsg(common.ErrForbidden, "无权操作该宠物"))
		return
	}

	var req UpdatePetRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	if req.Name != "" {
		pet.Name = req.Name
	}
	if req.Species != "" {
		pet.Species = req.Species
	}
	if req.Age != nil {
		pet.Age = *req.Age
	}
	if req.Gender != "" {
		pet.Gender = req.Gender
	}
	if req.Breed != "" {
		pet.Breed = req.Breed
	}
	if req.Note != "" {
		pet.Note = req.Note
	}

	if err := database.DB.Save(&pet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(pet))
}

func DeletePet(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	userID, _ := c.Get("userID")

	var pet models.Pet
	if err := database.DB.First(&pet, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponseWithMsg(common.ErrNotFound, "宠物不存在"))
		return
	}

	if pet.OwnerID != userID.(uint) {
		c.JSON(http.StatusForbidden, common.ErrorResponseWithMsg(common.ErrForbidden, "无权操作该宠物"))
		return
	}

	if err := database.DB.Delete(&pet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(nil))
}
