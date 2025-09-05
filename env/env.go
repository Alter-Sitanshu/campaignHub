package env

import (
	"os"
	"strconv"
)

type Config struct {
	DBADDR string
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string
}

var Cfg Config

func New() *Config {
	return &Config{
		DBADDR: GetString("DB_ADDR", ""),
		DBHost: GetString("DB_HOST", "localhost"),
		DBPort: GetString("DB_PORT", "5432"),
		DBUser: GetString("DB_USER", "postgres"),
		DBPass: GetString("DB_PASS", ""),
		DBName: GetString("DB_NAME", "postgres"),
	}
}

func GetInt(key string, fallback int) int {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	intVal, err := strconv.Atoi(val)
	if err != nil {
		return fallback
	}
	return intVal
}

func GetString(key, fallback string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		return fallback
	}
	return val
}
