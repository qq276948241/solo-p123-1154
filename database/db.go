package database

import (
	"fmt"
	"log"
	"pet-clinic-backend/config"
	"pet-clinic-backend/models"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitMySQL() {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.AppConfig.MySQLUser,
		config.AppConfig.MySQLPassword,
		config.AppConfig.MySQLHost,
		config.AppConfig.MySQLPort,
		config.AppConfig.MySQLDB,
	)

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}

	err = DB.AutoMigrate(
		&models.User{},
		&models.Doctor{},
		&models.Service{},
		&models.Schedule{},
		&models.Appointment{},
	)
	if err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
	}

	log.Println("MySQL connected and migrated successfully")
}
