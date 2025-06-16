package config

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/nessibeliyeltay/task-api/pkg/logger"
)

type Config struct {
	Server ServerConfig `json:"server"`
	Logger LoggerConfig `json:"logger"`
}

type ServerConfig struct {
	Port int    `json:"port"`
	Env  string `json:"env"`
}

type LoggerConfig struct {
	LogFile     string `json:"log_file"`
	LogToFile   bool   `json:"log_to_file"`
	LogToStdout bool   `json:"log_to_stdout"`
	MaxSize     int    `json:"max_size"`
	MaxBackups  int    `json:"max_backups"`
	MaxAge      int    `json:"max_age"`
	Compress    bool   `json:"compress"`
}

func (lc LoggerConfig) ToLoggerConfig() logger.Config {
	return logger.Config{
		LogFile:     lc.LogFile,
		LogToFile:   lc.LogToFile,
		LogToStdout: lc.LogToStdout,
		MaxSize:     lc.MaxSize,
		MaxBackups:  lc.MaxBackups,
		MaxAge:      lc.MaxAge,
		Compress:    lc.Compress,
	}
}

func New() *Config {
	configFile := "config/config.json"
	data, err := os.ReadFile(configFile)
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		panic(fmt.Sprintf("Error parsing config file: %v", err))
	}

	return &config
}
