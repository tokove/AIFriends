package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DBConfig struct {
	Driver   string `mapstructure:"driver"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Name     string `mapstructure:"name"`
	User     string
	Password string
	SSLMode  string `mapstructure:"sslmode"`
	TimeZone string `mapstructure:"time_zone"`
}

type CorsConfig struct {
	AllowOrigins     []string `mapstructure:"allow_origins"`
	AllowCredentials bool     `mapstructure:"allow_credentials"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxBackups int    `mapstructure:"max_backups"`
	MaxAge     int    `mapstructure:"max_age"`
}

type JwtConfig struct {
	AccessExp  int `mapstructure:"access_exp"`
	RefreshExp int `mapstructure:"refresh_exp"`
	Secret     string
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type Config struct {
	Server ServerConfig `mapstructure:"server"`
	DB     DBConfig     `mapstructure:"db"`
	Cors   CorsConfig   `mapstructure:"cors"`
	Log    LogConfig    `mapstructure:"log"`
	JWT    JwtConfig    `mapstructure:"jwt"`
	Redis  RedisConfig  `mapstructure:"redis"`
}

// LoadConfig 使用 Viper 读取 YAML + .env
func LoadConfig(path string) *Config {
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using system environment variables")
	}

	viper.SetConfigFile(path)

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Failed to read config file: %v", err)
	}

	cfg := &Config{}
	if err := viper.Unmarshal(cfg); err != nil {
		log.Fatalf("Failed to unmarshal config: %v", err)
	}

	cfg.DB.User = os.Getenv("DB_USER")
	cfg.DB.Password = os.Getenv("DB_PASSWORD")
	cfg.JWT.Secret = os.Getenv("JWT_SECRET")

	return cfg
}
