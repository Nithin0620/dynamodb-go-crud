package config

import (
	"strconv"

	"github.com/Nithin0620/dynamodb-go-crud/utils/env"
)

type config struct {
	Port        int
	Timeout     int
	Dialect     string
	DatabaseURI string
}

func GetConfig() *config {
	return &config{
		Port:        parseEnvToInt("PORT", "8080"),
		Timeout:     parseEnvToInt("TIMEOUT", "30"),
		Dialect:     env.GetEnv("DIALECT", "dynamodb"),
		DatabaseURI: env.GetEnv("DATABASE_URI", ""),
	}
}

func parseEnvToInt(envName string, defaultValue string) int {
	num, err := strconv.Atoi(env.GetEnv(envName, defaultValue))
	if err != nil {
		fallback, fallbackErr := strconv.Atoi(defaultValue)
		if fallbackErr != nil {
			return 0
		}
		return fallback
	}

	return num
}
