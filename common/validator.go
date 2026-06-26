package common

import (
	"net/http"
	"pet-clinic-backend/database"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func CheckExists(id uint, dest interface{}, notFoundCode ErrorCode) ErrorCode {
	if err := database.DB.First(dest, id).Error; err != nil {
		return notFoundCode
	}
	return Success
}

func CheckExistsWhere(dest interface{}, notFoundCode ErrorCode, query string, args ...interface{}) ErrorCode {
	if err := database.DB.Where(query, args...).First(dest).Error; err != nil {
		return notFoundCode
	}
	return Success
}

func CheckOwned(id uint, ownerID uint, dest interface{}, ownerGetter func() uint, notFoundCode ErrorCode, notOwnedCode ErrorCode) ErrorCode {
	if err := database.DB.First(dest, id).Error; err != nil {
		return notFoundCode
	}
	if ownerGetter() != ownerID {
		return notOwnedCode
	}
	return Success
}

func Ensure(c *gin.Context, code ErrorCode, httpStatus int) bool {
	if code == Success {
		return true
	}
	c.JSON(httpStatus, ErrorResponse(code))
	c.Abort()
	return false
}

func EnsureOK(c *gin.Context, err error, dbCode ErrorCode) bool {
	if err == nil {
		return true
	}
	status := http.StatusInternalServerError
	if dbCode == ErrDatabase {
		status = http.StatusInternalServerError
	}
	c.JSON(status, ErrorResponse(dbCode))
	c.Abort()
	return false
}

func DB() *gorm.DB {
	return database.DB
}
