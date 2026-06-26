package common

type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

type ErrorCode int

const (
	Success              ErrorCode = 0
	ErrInvalidParams     ErrorCode = 10001
	ErrUnauthorized      ErrorCode = 10002
	ErrForbidden         ErrorCode = 10003
	ErrNotFound          ErrorCode = 10004
	ErrInternalServer    ErrorCode = 10005
	ErrDatabase          ErrorCode = 10006

	ErrUserExists        ErrorCode = 20001
	ErrUserNotFound      ErrorCode = 20002
	ErrWrongPassword     ErrorCode = 20003
	ErrTokenInvalid      ErrorCode = 20004
	ErrTokenExpired      ErrorCode = 20005

	ErrDoctorNotFound    ErrorCode = 30001
	ErrScheduleNotFound  ErrorCode = 30002
	ErrScheduleExists    ErrorCode = 30003

	ErrAppointmentConflict ErrorCode = 40001
	ErrAppointmentNotFound ErrorCode = 40002
	ErrAppointmentStatus   ErrorCode = 40003
	ErrServiceNotFound     ErrorCode = 40004
	ErrScheduleFull        ErrorCode = 40005
	ErrScheduleMismatch    ErrorCode = 40006

	ErrPetNotFound   ErrorCode = 50001
	ErrPetNotOwned   ErrorCode = 50002
	ErrPetNameRequired ErrorCode = 50003
)

var errorMessages = map[ErrorCode]string{
	Success:              "success",
	ErrInvalidParams:     "参数错误",
	ErrUnauthorized:      "未授权",
	ErrForbidden:         "无权限访问",
	ErrNotFound:          "资源不存在",
	ErrInternalServer:    "服务器内部错误",
	ErrDatabase:          "数据库操作失败",

	ErrUserExists:        "用户已存在",
	ErrUserNotFound:      "用户不存在",
	ErrWrongPassword:     "密码错误",
	ErrTokenInvalid:      "Token无效",
	ErrTokenExpired:      "Token已过期",

	ErrDoctorNotFound:    "医生不存在",
	ErrScheduleNotFound:  "排班不存在",
	ErrScheduleExists:    "该时段已有排班",

	ErrAppointmentConflict: "预约时段冲突",
	ErrAppointmentNotFound: "预约不存在",
	ErrAppointmentStatus:   "预约状态不允许此操作",
	ErrServiceNotFound:     "服务项目不存在",
	ErrScheduleFull:        "该时段预约已满",
	ErrScheduleMismatch:    "排班与医生不匹配",

	ErrPetNotFound:     "宠物不存在",
	ErrPetNotOwned:     "只能预约自己名下的宠物",
	ErrPetNameRequired: "宠物ID或宠物名字必填",
}

func (e ErrorCode) Message() string {
	if msg, ok := errorMessages[e]; ok {
		return msg
	}
	return "未知错误"
}

func SuccessResponse(data interface{}) Response {
	return Response{
		Code:    int(Success),
		Message: Success.Message(),
		Data:    data,
	}
}

func ErrorResponse(code ErrorCode) Response {
	return Response{
		Code:    int(code),
		Message: code.Message(),
	}
}

func ErrorResponseWithMsg(code ErrorCode, msg string) Response {
	return Response{
		Code:    int(code),
		Message: msg,
	}
}
