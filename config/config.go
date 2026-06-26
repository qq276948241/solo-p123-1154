package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	ServerPort string

	MySQLHost     string
	MySQLPort     string
	MySQLUser     string
	MySQLPassword string
	MySQLDB       string

	RedisHost     string
	RedisPort     string
	RedisPassword string
	RedisDB       int

	JWTSecret string
	JWTExpire int
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env file not found, using environment variables")
	}

	AppConfig = &Config{
		ServerPort: getEnv("SERVER_PORT", "8080"),

		MySQLHost:     getEnv("MYSQL_HOST", "127.0.0.1"),
		MySQLPort:     getEnv("MYSQL_PORT", "3306"),
		MySQLUser:     getEnv("MYSQL_USER", "root"),
		MySQLPassword: getEnv("MYSQL_PASSWORD", ""),
		MySQLDB:       getEnv("MYSQL_DB", "pet_clinic"),

		RedisHost:     getEnv("REDIS_HOST", "127.0.0.1"),
		RedisPort:     getEnv("REDIS_PORT", "6379"),
		RedisPassword: getEnv("REDIS_PASSWORD", ""),
		RedisDB:       getEnvAsInt("REDIS_DB", 0),

		JWTSecret: getEnv("JWT_SECRET", "pet-clinic-secret-key-change-in-production"),
		JWTExpire: getEnvAsInt("JWT_EXPIRE_HOURS", 72),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if intVal, err := strconv.Atoi(value); err == nil {
			return intVal
		}
	}
	return defaultValue
}
