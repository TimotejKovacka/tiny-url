package config

import (
	"bytes"
	"os"
	"strings"

	"github.com/spf13/viper"
)

type ConfigYaml struct {
	Server SectionServer `yaml:"server"`
	DB     SectionDB     `yaml:"db"`
	Otel   SectionOtel   `yaml:"otel"`
}

type SectionServer struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

type SectionDB struct {
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	DBName   string `yaml:"db_name"`
}

type SectionOtel struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func setDefault() {
	// Server
	viper.SetDefault("server.host", "localhost")
	viper.SetDefault("server.port", "8080")
	// Database
	viper.SetDefault("db.host", "localhost")
	viper.SetDefault("db.port", "5432")
	viper.SetDefault("db.username", "admin")
	viper.SetDefault("db.password", "admin")
	viper.SetDefault("db.db_name", "test_db")
	// Otel
	viper.SetDefault("otel.host", "localhost")
	viper.SetDefault("otel.port", "4317")
}

func LoadConfig(configPath ...string) (*ConfigYaml, error) {
	config := &ConfigYaml{}

	setDefault()

	viper.SetConfigType("yaml")
	viper.AutomaticEnv() // read environment variables that match
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if len(configPath) > 0 && configPath[0] != "" {
		content, err := os.ReadFile(configPath[0])
		if err != nil {
			return config, err
		}

		if err := viper.ReadConfig(bytes.NewBuffer(content)); err != nil {
			return config, err
		}
	} else {
	}

	// Server
	config.Server.Host = viper.GetString("server.host")
	config.Server.Port = viper.GetString("server.port")

	// Database
	config.DB.Host = viper.GetString("db.host")
	config.DB.Port = viper.GetString("db.port")
	config.DB.Username = viper.GetString("db.username")
	config.DB.Password = viper.GetString("db.password")
	config.DB.DBName = viper.GetString("db.db_name")

	// Otel
	config.Otel.Host = viper.GetString("otel.host")
	config.Otel.Port = viper.GetString("otel.port")

	return config, nil
}
