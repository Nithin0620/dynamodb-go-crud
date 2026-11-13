package config 

import (
	"github.com/Nithin0620/dynamodb-go-crud/utils/env"
	"strconv"
)

type config struct {
	Port int
	Timeout int
	Dialect string
	DatabaseURI string
}

func GetConfig() *config {
	return Config{
		Port:	parseEnvToInt("PORT","8080"),  
		Timeout: parseEnvToInt("TIMEOUT","30"),
		Dialect: env.GetEnv("DIALECT","sqlite3"),
		DatabaseURI: env.GetEnv("DATABASE_URI",":memory:"),
	}
}

func parseEnvToInt(envName string,defaultValue sting) int{
	num,err := strconv.Atoi(env.GetEnv(envName,defaultValue))

	if err != nil {
		return 0
	}

	return num
}