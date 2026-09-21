package config

import (
	"errors"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Database DatabaseConfig
	JWT JWTConfig
	Redis RedisConfig
}

type RedisConfig struct {
	Host string
	Port string
}

type JWTConfig struct {
    SecretKey string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load() (Config, error) {
	_ = godotenv.Load()
	// if err != nil {
	// 	return Config{}, fmt.Errorf("Error loading .env file:%w", err)
	// }

	dbUser := os.Getenv("DB_USER")
	if dbUser == "" {
		return Config{}, fmt.Errorf("DB_USER is required")
	}

	dbPassword := os.Getenv("DB_PASSWORD")
	if dbPassword == "" {
		return Config{}, fmt.Errorf("DB_PASSWORD is required")
	}

	dbHost := os.Getenv("DB_HOST")
	if dbHost == "" {
		return Config{}, fmt.Errorf("DB_HOST is required")
	}
	dbPort := os.Getenv("DB_PORT")
	if dbPort == "" {
		return Config{}, fmt.Errorf("DB_PORT is required")
	}

	dbName := os.Getenv("DB_NAME")
	if dbName == "" {
		return Config{}, fmt.Errorf("DB_NAME is required")
	}

	dbSSLMode := os.Getenv("DB_SSLMODE")
	if dbSSLMode == "" {
		return Config{}, fmt.Errorf("DB_SSLMODE is required")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		return Config{}, errors.New("JWT_SECRET is required")
	}

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		return Config{}, errors.New("REDIS_HOST is required")
	}

	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		return Config{}, errors.New("REDIS_PORT is required")
	}

	dbConfig := DatabaseConfig{
		Host:     dbHost,
		User:     dbUser,
		Port:     dbPort,
		Password: dbPassword,
		Name:     dbName,
		SSLMode:  dbSSLMode,
	}

	return Config{
		Database: dbConfig,
		JWT: JWTConfig{
			SecretKey: jwtSecret,
		},
		Redis: RedisConfig{
			Host: redisHost,
			Port: redisPort,
		},
	}, nil

}
