package database

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"pet-clinic-backend/config"
	"pet-clinic-backend/models"
	"time"

	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitRedis() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.AppConfig.RedisHost, config.AppConfig.RedisPort),
		Password: config.AppConfig.RedisPassword,
		DB:       config.AppConfig.RedisDB,
	})

	_, err := RedisClient.Ping(Ctx).Result()
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}

	log.Println("Redis connected successfully")
}

func GetTodayScheduleCacheKey(doctorID uint, date string) string {
	return fmt.Sprintf("schedule:doctor:%d:date:%s", doctorID, date)
}

func CacheTodaySchedule(doctorID uint, date string, schedules []models.Schedule) error {
	key := GetTodayScheduleCacheKey(doctorID, date)
	data, err := json.Marshal(schedules)
	if err != nil {
		return err
	}
	return RedisClient.Set(Ctx, key, data, 24*time.Hour).Err()
}

func GetCachedTodaySchedule(doctorID uint, date string) ([]models.Schedule, error) {
	key := GetTodayScheduleCacheKey(doctorID, date)
	data, err := RedisClient.Get(Ctx, key).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	var schedules []models.Schedule
	err = json.Unmarshal([]byte(data), &schedules)
	if err != nil {
		return nil, err
	}
	return schedules, nil
}

func InvalidateScheduleCache(doctorID uint, date string) error {
	key := GetTodayScheduleCacheKey(doctorID, date)
	return RedisClient.Del(Ctx, key).Err()
}
