package handlers

import (
	"net/http"
	"pet-clinic-backend/common"
	"pet-clinic-backend/database"
	"pet-clinic-backend/models"
	"strconv"

	"github.com/gin-gonic/gin"
)

type CreateAppointmentRequest struct {
	DoctorID   uint   `json:"doctor_id" binding:"required"`
	ScheduleID uint   `json:"schedule_id" binding:"required"`
	ServiceID  uint   `json:"service_id" binding:"required"`
	PetName    string `json:"pet_name" binding:"required"`
	PetType    string `json:"pet_type"`
	Note       string `json:"note"`
}

type RejectAppointmentRequest struct {
	Reason string `json:"reason" binding:"required"`
}

func checkTimeConflict(doctorID uint, date, startTime, endTime string, excludeID ...uint) (bool, error) {
	query := database.DB.Model(&models.Appointment{}).Where(
		"doctor_id = ? AND date = ? AND status IN ? AND ((start_time <= ? AND end_time > ?) OR (start_time < ? AND end_time >= ?))",
		doctorID, date,
		[]string{models.AppointmentStatusPending, models.AppointmentStatusConfirmed},
		endTime, startTime, endTime, startTime,
	)

	if len(excludeID) > 0 {
		query = query.Where("id != ?", excludeID[0])
	}

	var count int64
	if err := query.Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func CreateAppointment(c *gin.Context) {
	var req CreateAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	userID, _ := c.Get("userID")

	var schedule models.Schedule
	if err := database.DB.First(&schedule, req.ScheduleID).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrScheduleNotFound))
		return
	}

	if schedule.DoctorID != req.DoctorID {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	if !schedule.IsActive {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrScheduleNotFound))
		return
	}

	if schedule.Booked >= schedule.MaxSlots {
		c.JSON(http.StatusBadRequest, common.ErrorResponseWithMsg(common.ErrAppointmentConflict, "该时段预约已满"))
		return
	}

	var service models.Service
	if err := database.DB.First(&service, req.ServiceID).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrServiceNotFound))
		return
	}

	conflict, err := checkTimeConflict(req.DoctorID, schedule.Date, schedule.StartTime, schedule.EndTime)
	if err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}
	if conflict {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrAppointmentConflict))
		return
	}

	appointment := models.Appointment{
		UserID:     userID.(uint),
		DoctorID:   req.DoctorID,
		ScheduleID: req.ScheduleID,
		ServiceID:  req.ServiceID,
		PetName:    req.PetName,
		PetType:    req.PetType,
		Date:       schedule.Date,
		StartTime:  schedule.StartTime,
		EndTime:    schedule.EndTime,
		Status:     models.AppointmentStatusPending,
		Note:       req.Note,
	}

	tx := database.DB.Begin()

	if err := tx.Create(&appointment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	if err := tx.Model(&schedule).UpdateColumn("booked", schedule.Booked+1).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	tx.Commit()

	database.InvalidateScheduleCache(req.DoctorID, schedule.Date)

	c.JSON(http.StatusOK, common.SuccessResponse(appointment))
}

func GetAppointment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var appointment models.Appointment
	if err := database.DB.Preload("User").Preload("Doctor").Preload("Service").
		First(&appointment, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrAppointmentNotFound))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(appointment))
}

func GetMyAppointments(c *gin.Context) {
	userID, _ := c.Get("userID")
	status := c.Query("status")

	query := database.DB.Where("user_id = ?", userID.(uint))
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var appointments []models.Appointment
	if err := query.Preload("Doctor").Preload("Service").
		Order("created_at DESC").Find(&appointments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(appointments))
}

func GetDoctorAppointments(c *gin.Context) {
	userID, _ := c.Get("userID")
	role, _ := c.Get("role")

	var doctor models.Doctor
	var doctorID uint

	if role.(string) == models.RoleAdmin {
		idParam := c.Query("doctor_id")
		if idParam != "" {
			id, err := strconv.ParseUint(idParam, 10, 32)
			if err != nil {
				c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
				return
			}
			doctorID = uint(id)
		}
	} else {
		if err := database.DB.Where("user_id = ?", userID.(uint)).First(&doctor).Error; err != nil {
			c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrDoctorNotFound))
			return
		}
		doctorID = doctor.ID
	}

	date := c.Query("date")
	status := c.Query("status")

	query := database.DB.Where("doctor_id = ?", doctorID)
	if date != "" {
		query = query.Where("date = ?", date)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	var appointments []models.Appointment
	if err := query.Preload("User").Preload("Service").
		Order("date ASC, start_time ASC").Find(&appointments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(appointments))
}

func GetAllAppointments(c *gin.Context) {
	status := c.Query("status")
	date := c.Query("date")

	query := database.DB.Model(&models.Appointment{})
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if date != "" {
		query = query.Where("date = ?", date)
	}

	var appointments []models.Appointment
	if err := query.Preload("User").Preload("Doctor").Preload("Service").
		Order("created_at DESC").Find(&appointments).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(appointments))
}

func ConfirmAppointment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var appointment models.Appointment
	if err := database.DB.First(&appointment, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrAppointmentNotFound))
		return
	}

	if appointment.Status != models.AppointmentStatusPending {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrAppointmentStatus))
		return
	}

	appointment.Status = models.AppointmentStatusConfirmed

	if err := database.DB.Save(&appointment).Error; err != nil {
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	c.JSON(http.StatusOK, common.SuccessResponse(appointment))
}

func RejectAppointment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var req RejectAppointmentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	var appointment models.Appointment
	if err := database.DB.First(&appointment, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrAppointmentNotFound))
		return
	}

	if appointment.Status != models.AppointmentStatusPending {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrAppointmentStatus))
		return
	}

	appointment.Status = models.AppointmentStatusRejected
	appointment.RejectReason = req.Reason

	tx := database.DB.Begin()

	if err := tx.Save(&appointment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	var schedule models.Schedule
	if err := tx.First(&schedule, appointment.ScheduleID).Error; err == nil {
		if schedule.Booked > 0 {
			tx.Model(&schedule).UpdateColumn("booked", schedule.Booked-1)
		}
	}

	tx.Commit()

	database.InvalidateScheduleCache(appointment.DoctorID, appointment.Date)

	c.JSON(http.StatusOK, common.SuccessResponse(appointment))
}

func CancelAppointment(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrInvalidParams))
		return
	}

	userID, _ := c.Get("userID")

	var appointment models.Appointment
	if err := database.DB.First(&appointment, uint(id)).Error; err != nil {
		c.JSON(http.StatusNotFound, common.ErrorResponse(common.ErrAppointmentNotFound))
		return
	}

	if appointment.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, common.ErrorResponse(common.ErrForbidden))
		return
	}

	if appointment.Status != models.AppointmentStatusPending && appointment.Status != models.AppointmentStatusConfirmed {
		c.JSON(http.StatusBadRequest, common.ErrorResponse(common.ErrAppointmentStatus))
		return
	}

	appointment.Status = models.AppointmentStatusCancelled

	tx := database.DB.Begin()

	if err := tx.Save(&appointment).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, common.ErrorResponse(common.ErrDatabase))
		return
	}

	var schedule models.Schedule
	if err := tx.First(&schedule, appointment.ScheduleID).Error; err == nil {
		if schedule.Booked > 0 {
			tx.Model(&schedule).UpdateColumn("booked", schedule.Booked-1)
		}
	}

	tx.Commit()

	database.InvalidateScheduleCache(appointment.DoctorID, appointment.Date)

	c.JSON(http.StatusOK, common.SuccessResponse(appointment))
}
