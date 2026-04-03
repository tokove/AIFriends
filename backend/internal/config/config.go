package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port         int
	Mode         string
	AllowedHosts []string
}

type DBConfig struct {
	Driver   string
	Host     string
	Port     int
	Name     string
	User     string
	Password string
	SSLMode  string
}

type CorsConfig struct {
	AllowOrigins     []string
	AllowCredentials bool
}

type LogConfig struct {
	Level      string
	Filename   string
	MaxSize    int
	MaxBackups int
	MaxAge     int
}

type Config struct {
	Server    ServerConfig
	DB        DBConfig
	Cors      CorsConfig
	JWTSecret string
	Log       LogConfig
}

// LoadConfig 使用 Viper 读取 YAML + .env
func LoadConfig(path string) *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}
	
	v := viper.New()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	cfg := &Config{
		Server: ServerConfig{
			Port: v.GetInt("server.port"),
			Mode: v.GetString("server.mode"),
		},
		DB: DBConfig{
			Driver:   v.GetString("db.driver"),
			Host:     v.GetString("db.host"),
			Port:     v.GetInt("db.port"),
			Name:     v.GetString("db.name"),
			User:     os.Getenv("DB_USER"),     // 从环境变量
			Password: os.Getenv("DB_PASSWORD"), // 从环境变量
			SSLMode:  v.GetString("db.sslmode"),
		},
		Log: LogConfig{
			Level:      v.GetString("log.level"),
			Filename:   v.GetString("log.filename"),
			MaxSize:    v.GetInt("log.max_size"),
			MaxAge:     v.GetInt("log.max_age"),
			MaxBackups: v.GetInt("log.max_backups"),
		},
		Cors: CorsConfig{
			AllowOrigins:     v.GetStringSlice("cors.allow_origins"),
			AllowCredentials: v.GetBool("cors.allow_credentials"),
		},
		JWTSecret: os.Getenv("JWT_SECRET"),
	}

	return cfg
}
