package models

import (
	"time"
)

const (
	RoleUser     = "user"
	RoleDoctor   = "doctor"
	RoleAdmin    = "admin"
)

type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Phone     string    `gorm:"size:20;uniqueIndex;not null" json:"phone"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Username  string    `gorm:"size:50;not null" json:"username"`
	Role      string    `gorm:"size:20;default:user;not null" json:"role"`
	Avatar    string    `gorm:"size:255" json:"avatar"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Doctor struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	Name        string    `gorm:"size:50;not null" json:"name"`
	Title       string    `gorm:"size:50" json:"title"`
	Specialty   string    `gorm:"size:100" json:"specialty"`
	Description string    `gorm:"type:text" json:"description"`
	Avatar      string    `gorm:"size:255" json:"avatar"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Service struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	Name        string    `gorm:"size:100;not null" json:"name"`
	Category    string    `gorm:"size:50;not null" json:"category"`
	Description string    `gorm:"type:text" json:"description"`
	Price       float64   `gorm:"type:decimal(10,2);not null" json:"price"`
	Duration    int       `gorm:"not null" json:"duration"`
	IsActive    bool      `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Pet struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	OwnerID   uint      `gorm:"index;not null" json:"owner_id"`
	Name      string    `gorm:"size:50;not null" json:"name"`
	Species   string    `gorm:"size:30;not null" json:"species"`
	Age       int       `gorm:"default:0" json:"age"`
	Gender    string    `gorm:"size:10" json:"gender"`
	Breed     string    `gorm:"size:50" json:"breed"`
	Note      string    `gorm:"type:text" json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Schedule struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	DoctorID  uint      `gorm:"index;not null" json:"doctor_id"`
	Date      string    `gorm:"size:10;index;not null" json:"date"`
	StartTime string    `gorm:"size:5;not null" json:"start_time"`
	EndTime   string    `gorm:"size:5;not null" json:"end_time"`
	MaxSlots  int       `gorm:"default:1;not null" json:"max_slots"`
	Booked    int       `gorm:"default:0;not null" json:"booked"`
	IsActive  bool      `gorm:"default:true" json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	Doctor Doctor `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
}

const (
	AppointmentStatusPending  = "pending"
	AppointmentStatusConfirmed = "confirmed"
	AppointmentStatusRejected = "rejected"
	AppointmentStatusCancelled = "cancelled"
	AppointmentStatusCompleted = "completed"
)

type Appointment struct {
	ID          uint      `gorm:"primaryKey" json:"id"`
	UserID      uint      `gorm:"index;not null" json:"user_id"`
	DoctorID    uint      `gorm:"index;not null" json:"doctor_id"`
	ScheduleID  uint      `gorm:"index;not null" json:"schedule_id"`
	ServiceID   uint      `gorm:"index;not null" json:"service_id"`
	PetID       uint      `gorm:"index" json:"pet_id"`
	PetName     string    `gorm:"size:50;not null" json:"pet_name"`
	PetType     string    `gorm:"size:30" json:"pet_type"`
	Date        string    `gorm:"size:10;index;not null" json:"date"`
	StartTime   string    `gorm:"size:5;not null" json:"start_time"`
	EndTime     string    `gorm:"size:5;not null" json:"end_time"`
	Status      string    `gorm:"size:20;default:pending;index;not null" json:"status"`
	Note        string    `gorm:"type:text" json:"note"`
	RejectReason string   `gorm:"type:text" json:"reject_reason"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	User     User     `gorm:"foreignKey:UserID" json:"user,omitempty"`
	Doctor   Doctor   `gorm:"foreignKey:DoctorID" json:"doctor,omitempty"`
	Service  Service  `gorm:"foreignKey:ServiceID" json:"service,omitempty"`
	Pet      Pet      `gorm:"foreignKey:PetID" json:"pet,omitempty"`
	Schedule Schedule `gorm:"foreignKey:ScheduleID" json:"schedule,omitempty"`
}
