// config/config.go

// env 데이터 연결 및 설정을 관리

package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port string

	MongoURI string
	DBName   string

	JWTSecret string

	GoogleWebClientID string
	KakaoAdminKey     string
	AppleBundleID     string
	AppleP8Key        string
	AppleTeamID       string
	AppleKeyID        string
}

var AppConfig *Config

func LoadConfig() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found. Using system variables.")
	}

	AppConfig = &Config{
		Port: getEnv("PORT", "8080"),

		MongoURI: getEnv("MONGO_URI", "mongodb://localhost:27017"),
		DBName:   getEnv("DB_NAME", "bapddang-dev"),

		JWTSecret: getEnv("JWT_SECRET_KEY", "default_secret"),

		GoogleWebClientID: getEnv("GOOGLE_WEB_CLIENT_ID", ""),
		KakaoAdminKey:     getEnv("KAKAO_ADMIN_KEY", ""),
		AppleBundleID:     getEnv("APPLE_BUNDLE_ID", ""),
		AppleP8Key:        getEnv("APPLE_P8_KEY", ""),
		AppleTeamID:       getEnv("APPLE_TEAM_ID", ""),
		AppleKeyID:        getEnv("APPLE_KEY_ID", ""),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
