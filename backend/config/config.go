package config

import (
	"os"
)

type Config struct {
	DatabaseDSN string
	ServerPort  string
}

func LoadConfig() *Config {
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		dsn = "root:root@tcp(127.0.0.1:3306)/codeagent?charset=utf8mb4&parseTime=True&loc=Local"
	}
	
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	return &Config{
		DatabaseDSN: dsn,
		ServerPort:  port,
	}
}
