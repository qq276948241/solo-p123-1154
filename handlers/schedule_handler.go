package handlers

import (
	"net/http"
	"pet-clinic-backend/common"
	"pet-clinic-backend/database"
	"pet-clinic-backend/models"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type CreateScheduleRequest struct {
	DoctorID  uint   `json:"doctor_id" binding:"required"`
	Date      string `json:"date" binding:"required"`
	StartTime string `json:"start_time" binding:"required"`
	EndTime   string `json:"end_time" binding:"required"`
	MaxSlots  int    `json:"max_slots"`
}

type UpdateScheduleRequest struct {
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	MaxSlots  int    `json:"max_slots"`
	IsActive  *bool  `json:"is_active"`
}

func CreateSchedule(c *gin.Context) {
	var req CreateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	if req.MaxSlots <= 0 {
		req.MaxSlots = 1
	}

	var doctor models.Doctor
	if err := database.DB.First(&doctor, req.DoctorID).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrDoctorNotFound))
		return
	}

	var existing models.Schedule
	result := database.DB.Where("doctor_id = ? AND date = ? AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?))",
		req.DoctorID, req.Date, req.EndTime, req.StartTime, req.EndTime, req.StartTime).First(&existing)
	if result.Error == nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrScheduleExists))
		return
	}

	schedule := models.Schedule{
		DoctorID:  req.DoctorID,
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		MaxSlots:  req.MaxSlots,
		IsActive:  true,
	}

	if err := database.DB.Create(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	database.InvalidateScheduleCache(req.DoctorID, req.Date)

	c.JSON(http.StatusOK, common.SuccessResponse(schedule))
}

func GetSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var schedule models.Schedule
	if err := database.DB.Preload("Doctor").First(&schedule, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrScheduleNotFound))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(schedule))
}

func GetDoctorSchedules(c *gin.Context) {
	doctorID, err := strconv.ParseUint(c.Param("doctor_id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	date := c.Query("date")
	if date == "" {
		today := time.Now().Format("2006-01-02")
		date = today
	}

	cachedSchedules, err := database.GetCachedTodaySchedule(uint(doctorID), date)
	if err == nil && cachedSchedules != nil {
		c.JSON(http.StatusOK, common.SuccessResponse(cachedSchedules))
		return
	}

	var schedules []models.Schedule
	if err := database.DB.Where("doctor_id = ? AND date = ? AND is_active = ?", uint(doctorID), date, true).
		Preload("Doctor").Find(&schedules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	if len(schedules) > 0 {
		database.CacheTodaySchedule(uint(doctorID), date, schedules)
	}

	c.JSON(http.StatusOK, common.SuccessResponse(schedules))
}

func GetAllSchedules(c *gin.Context) {
	date := c.Query("date")
	doctorID := c.Query("doctor_id")

	query := database.DB.Model(&models.Schedule{})

	if date != "" {
		query = query.Where("date = ?", date)
	}
	if doctorID != "" {
		query = query.Where("doctor_id = ?", doctorID)
	}

	var schedules []models.Schedule
	if err := query.Preload("Doctor").Find(&schedules).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(schedules))
}

func UpdateSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var req UpdateScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var schedule models.Schedule
	if err := database.DB.First(&schedule, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrScheduleNotFound))
		return
	}

	if req.StartTime != "" {
		schedule.StartTime = req.StartTime
	}
	if req.EndTime != "" {
		schedule.EndTime = req.EndTime
	}
	if req.MaxSlots > 0 {
		schedule.MaxSlots = req.MaxSlots
	}
	if req.IsActive != nil {
		schedule.IsActive = *req.IsActive
	}

	if err := database.DB.Save(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	database.InvalidateScheduleCache(schedule.DoctorID, schedule.Date)

	c.JSON(http.StatusOK, common.SuccessResponse(schedule))
}

func DeleteSchedule(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var schedule models.Schedule
	if err := database.DB.First(&schedule, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrScheduleNotFound))
		return
	}

	database.InvalidateScheduleCache(schedule.DoctorID, schedule.Date)

	if err := database.DB.Delete(&schedule).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(nil))
}
