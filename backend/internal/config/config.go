package config

import (
	"github.com/spf13/viper"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	JWT      JWTConfig      `mapstructure:"jwt"`
	Crypto   CryptoConfig   `mapstructure:"crypto"`
}

// CryptoConfig 加密配置
type CryptoConfig struct {
	SecretKey string `mapstructure:"secret_key"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port string `mapstructure:"port"`
	Mode string `mapstructure:"mode"` // debug, release, test
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Driver string `mapstructure:"driver"` // mysql, sqlite
	DSN    string `mapstructure:"dsn"`
}

// JWTConfig JWT配置
type JWTConfig struct {
	Secret     string `mapstructure:"secret"`
	ExpireHour int    `mapstructure:"expire_hour"`
}

// Load 加载配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")

	// 设置默认值
	viper.SetDefault("server.port", "8080")
	viper.SetDefault("server.mode", "debug")
	viper.SetDefault("database.driver", "sqlite")
	viper.SetDefault("database.dsn", "autostack.db")
	viper.SetDefault("jwt.secret", "autostack-secret-key")
	viper.SetDefault("jwt.expire_hour", 24)
	viper.SetDefault("crypto.secret_key", "autostack-crypto-secret-key32!")

	// 绑定环境变量
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.mode", "SERVER_MODE")
	viper.BindEnv("database.driver", "DATABASE_DRIVER")
	viper.BindEnv("database.dsn", "DATABASE_DSN")
	viper.BindEnv("jwt.secret", "JWT_SECRET")
	viper.BindEnv("crypto.secret_key", "CRYPTO_SECRET_KEY")

	// 读取配置文件（可选）
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, err
		}
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
