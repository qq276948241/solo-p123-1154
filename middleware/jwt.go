package middleware

import (
	"net/http"
	"pet-clinic-backend/common"
	"pet-clinic-backend/config"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	UserID   uint   `json:"user_id"`
	Phone    string `json:"phone"`
	Role     string `json:"role"`
	Username string `json:"username"`
	jwt.RegisteredClaims
}

func GenerateToken(userID uint, phone, role, username string) (string, error) {
	claims := Claims{
		UserID:   userID,
		Phone:    phone,
		Role:     role,
		Username: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(config.AppConfig.JWTExpire) * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "pet-clinic",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}

func ParseToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(config.AppConfig.JWTSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrSignatureInvalid
}

func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse(common.ErrUnauthorized))
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, common.ErrorResponse(common.ErrTokenInvalid))
			c.Abort()
			return
		}

		claims, err := ParseToken(parts[1])
		if err != nil {
			if err == jwt.ErrTokenExpired {
				c.JSON(http.StatusUnauthorized, common.ErrorResponse(common.ErrTokenExpired))
			} else {
				c.JSON(http.StatusUnauthorized, common.ErrorResponse(common.ErrTokenInvalid))
			}
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("phone", claims.Phone)
		c.Set("role", claims.Role)
		c.Set("username", claims.Username)
		c.Next()
	}
}

func RoleAuth(allowedRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusForbidden, common.ErrorResponse(common.ErrForbidden))
			c.Abort()
			return
		}

		userRole := role.(string)
		allowed := false
		for _, r := range allowedRoles {
			if r == userRole {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, common.ErrorResponse(common.ErrForbidden))
			c.Abort()
			return
		}

		c.Next()
	}
}
