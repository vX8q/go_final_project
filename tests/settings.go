package tests

import (
	"os"
	"strconv"
)

func getEnvInt(key string, defaultValue int) int {
	if value, exists := os.LookupEnv(key); exists {
		if p, err := strconv.Atoi(value); err == nil {
			return p
		}
	}
	return defaultValue
}

func getEnvString(key string, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}

var (
	Port   = getEnvInt("TODO_PORT", 7540)
	DBFile = getEnvString("TODO_DBFILE", "../scheduler.db")
)
