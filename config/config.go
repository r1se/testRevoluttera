// Package config описывает структуру конфига
package config

import (
	"fmt"
	"github.com/ilyakaznacheev/cleanenv"
)

const defaultConfigFilename = "config.yml"

// Paths to search files in, on order of priority, latest takes precedence
var configPaths = []string{"/usr/local/etc/", "./configs/"}

// Config struct with field name
type Config struct {
	LogFile  string `yaml:"log_file" env:"LOG_FILE" env-default:"user-service.log"`
	LogLevel string `yaml:"log_level" env:"LOG_LEVEL" env-default:"info"`
	RESTApi  struct {
		Host string `yaml:"host" env:"REST_HOST" env-default:"0.0.0.0"`
		Port int    `yaml:"port" env:"REST_PORT" env-default:"8080"`
	} `yaml:"rest_api"`
}

// ReadConfig вспомогательная функция для чтения файла конфигурации
func ReadConfig() *Config {
	var conf Config

	for _, file := range getConfigFiles() {
		_ = cleanenv.ReadConfig(file, &conf)
	}

	_ = cleanenv.ReadEnv(&conf)

	return &conf
}

// getConfigFiles вспомогательная функция для получения файлов конфигурации
func getConfigFiles() []string {
	files := make([]string, 0)
	for _, path := range configPaths {
		files = append(
			files,
			fmt.Sprintf("%s%s", path, defaultConfigFilename))
	}
	return files
}
