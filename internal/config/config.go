package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	AppEnv              string
	Port                string
	DBURL               string
	CloudinaryCloudName string
	CloudinaryAPIKey    string
	CloudinaryAPISecret string
	ResendAPIKey        string
	ClerkSecretKey      string
	ClerkPublishableKey string
	AllowedOrigins      string // Optional: comma-separated list of allowed CORS origins.
}

func LoadConfig() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, reading from environment")
	}

	cfg := &Config{
		AppEnv:              getEnv("APP_ENV"),
		Port:                getEnv("PORT"),
		DBURL:               getEnv("DB_URL"),
		CloudinaryCloudName: getEnv("CLOUDINARY_CLOUD_NAME"),
		CloudinaryAPIKey:    getEnv("CLOUDINARY_API_KEY"),
		CloudinaryAPISecret: getEnv("CLOUDINARY_API_SECRET"),
		ResendAPIKey:        getEnv("RESEND_API_KEY"),
		ClerkSecretKey:      getEnv("CLERK_SECRET_KEY"),
		ClerkPublishableKey: getEnv("CLERK_PUBLISHABLE_KEY"),
		AllowedOrigins:      os.Getenv("ALLOWED_ORIGINS"), // Not required; falls back to allow-all.
	}

	return cfg
}

func getEnv(key string) string {
	val := os.Getenv(key)
	if val == "" {
		log.Fatalf("required environment variable %s is not set", key)
	}
	return val
}
