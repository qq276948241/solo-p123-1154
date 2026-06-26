package main

import (
	"pet-clinic-backend/handlers"
	"pet-clinic-backend/middleware"
	"pet-clinic-backend/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	api := r.Group("/api")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", handlers.Register)
			auth.POST("/login", handlers.Login)
			auth.GET("/me", middleware.JWTAuth(), handlers.GetCurrentUser)
		}

		doctors := api.Group("/doctors")
		{
			doctors.GET("", handlers.GetAllDoctors)
			doctors.GET("/:id", handlers.GetDoctor)

			adminDoctors := doctors.Group("")
			adminDoctors.Use(middleware.JWTAuth(), middleware.RoleAuth(models.RoleAdmin))
			{
				adminDoctors.POST("", handlers.CreateDoctor)
				adminDoctors.PUT("/:id", handlers.UpdateDoctor)
				adminDoctors.DELETE("/:id", handlers.DeleteDoctor)
			}
		}

		services := api.Group("/services")
		{
			services.GET("", handlers.GetAllServices)
			services.GET("/categories", handlers.GetServiceCategories)
			services.GET("/:id", handlers.GetService)

			adminServices := services.Group("")
			adminServices.Use(middleware.JWTAuth(), middleware.RoleAuth(models.RoleAdmin))
			{
				adminServices.POST("", handlers.CreateService)
				adminServices.PUT("/:id", handlers.UpdateService)
				adminServices.DELETE("/:id", handlers.DeleteService)
			}
		}

		schedules := api.Group("/schedules")
		schedules.Use(middleware.JWTAuth())
		{
			schedules.GET("", handlers.GetAllSchedules)
			schedules.GET("/:id", handlers.GetSchedule)
			schedules.GET("/doctor/:doctor_id", handlers.GetDoctorSchedules)

			adminSchedules := schedules.Group("")
			adminSchedules.Use(middleware.RoleAuth(models.RoleAdmin, models.RoleDoctor))
			{
				adminSchedules.POST("", handlers.CreateSchedule)
				adminSchedules.PUT("/:id", handlers.UpdateSchedule)
				adminSchedules.DELETE("/:id", handlers.DeleteSchedule)
			}
		}

		appointments := api.Group("/appointments")
		appointments.Use(middleware.JWTAuth())
		{
			appointments.POST("", handlers.CreateAppointment)
			appointments.GET("/my", handlers.GetMyAppointments)
			appointments.GET("/:id", handlers.GetAppointment)
			appointments.POST("/:id/cancel", handlers.CancelAppointment)

			doctorAppointments := appointments.Group("")
			doctorAppointments.Use(middleware.RoleAuth(models.RoleDoctor, models.RoleAdmin))
			{
				doctorAppointments.GET("", handlers.GetAllAppointments)
				doctorAppointments.GET("/doctor/mine", handlers.GetDoctorAppointments)
				doctorAppointments.POST("/:id/confirm", handlers.ConfirmAppointment)
				doctorAppointments.POST("/:id/reject", handlers.RejectAppointment)
			}
		}
	}

	return r
}
