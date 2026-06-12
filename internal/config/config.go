package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Redis    RedisConfig    `json:"redis"`
	JWT      JWTConfig      `json:"jwt"`
	Log      LogConfig      `json:"log"`
	Kafka    KafkaConfig    `json:"kafka"`
	Alert    AlertConfig    `json:"alert"`
}

type ServerConfig struct {
	Port            int           `json:"port"`
	Mode            string        `json:"mode"`
	ReadTimeout     time.Duration `json:"read_timeout"`
	WriteTimeout    time.Duration `json:"write_timeout"`
	MaxHeaderBytes  int           `json:"max_header_bytes"`
	TrustedProxies  []string      `json:"trusted_proxies"`
}

type DatabaseConfig struct {
	Host            string `json:"host"`
	Port            int    `json:"port"`
	User            string `json:"user"`
	Password        string `json:"password"`
	DBName          string `json:"dbname"`
	SSLMode         string `json:"sslmode"`
	MaxOpenConns    int    `json:"max_open_conns"`
	MaxIdleConns    int    `json:"max_idle_conns"`
	ConnMaxLifetime int    `json:"conn_max_lifetime"` // seconds
}

func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode)
}

// SafeDSN returns a DSN string with password masked (for logging)
func (d *DatabaseConfig) SafeDSN() string {
	return fmt.Sprintf("host=%s port=%d user=%s password=**** dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.DBName, d.SSLMode)
}

type RedisConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	DB       int    `json:"db"`
	PoolSize int    `json:"pool_size"`
}

func (r *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}

type JWTConfig struct {
	Secret      string        `json:"secret"`
	ExpireHours time.Duration `json:"expire_hours"`
	Issuer      string        `json:"issuer"`
}

type LogConfig struct {
	Level  string `json:"level"`
	Format string `json:"format"`
}

type KafkaConfig struct {
	Brokers []string `json:"brokers"`
	GroupID string   `json:"group_id"`
	Enabled bool     `json:"enabled"`
}

type AlertConfig struct {
	EvaluateInterval int `json:"evaluate_interval"` // seconds
	Enabled          bool `json:"enabled"`
}

var globalConfig *Config

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		// Allow missing config file if env vars are set
		if os.IsNotExist(err) {
			globalConfig = defaultConfig()
			applyEnvOverrides(globalConfig)
			return globalConfig, nil
		}
		return nil, fmt.Errorf("read config file: %w", err)
	}
	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	applyDefaults(&cfg)
	applyEnvOverrides(&cfg)
	globalConfig = &cfg
	return &cfg, nil
}

func Get() *Config {
	return globalConfig
}

func defaultConfig() *Config {
	cfg := &Config{}
	applyDefaults(cfg)
	return cfg
}

func applyDefaults(cfg *Config) {
	if cfg.Server.Port == 0 {
		cfg.Server.Port = 8080
	}
	if cfg.Server.Mode == "" {
		cfg.Server.Mode = "debug"
	}
	if cfg.Server.ReadTimeout == 0 {
		cfg.Server.ReadTimeout = 30
	}
	if cfg.Server.WriteTimeout == 0 {
		cfg.Server.WriteTimeout = 30
	}
	if cfg.Server.MaxHeaderBytes == 0 {
		cfg.Server.MaxHeaderBytes = 1 << 20 // 1MB
	}
	if cfg.Database.Port == 0 {
		cfg.Database.Port = 5432
	}
	if cfg.Database.MaxOpenConns == 0 {
		cfg.Database.MaxOpenConns = 25
	}
	if cfg.Database.MaxIdleConns == 0 {
		cfg.Database.MaxIdleConns = 5
	}
	if cfg.Database.SSLMode == "" {
		cfg.Database.SSLMode = "disable"
	}
	if cfg.Redis.Port == 0 {
		cfg.Redis.Port = 6379
	}
	if cfg.Redis.PoolSize == 0 {
		cfg.Redis.PoolSize = 20
	}
	if cfg.JWT.ExpireHours == 0 {
		cfg.JWT.ExpireHours = 24
	}
	if cfg.JWT.Issuer == "" {
		cfg.JWT.Issuer = "live-commerce-bi"
	}
	if cfg.Log.Level == "" {
		cfg.Log.Level = "info"
	}
	if cfg.Alert.EvaluateInterval == 0 {
		cfg.Alert.EvaluateInterval = 60
	}
}

// applyEnvOverrides allows overriding config with environment variables
func applyEnvOverrides(cfg *Config) {
	// Server
	envInt(&cfg.Server.Port, "SERVER_PORT")
	envStr(&cfg.Server.Mode, "SERVER_MODE")

	// Database
	envStr(&cfg.Database.Host, "DB_HOST")
	envInt(&cfg.Database.Port, "DB_PORT")
	envStr(&cfg.Database.User, "DB_USER")
	envStr(&cfg.Database.Password, "DB_PASSWORD")
	envStr(&cfg.Database.DBName, "DB_NAME")
	envStr(&cfg.Database.SSLMode, "DB_SSLMODE")
	envInt(&cfg.Database.MaxOpenConns, "DB_MAX_OPEN_CONNS")
	envInt(&cfg.Database.MaxIdleConns, "DB_MAX_IDLE_CONNS")

	// Redis
	envStr(&cfg.Redis.Host, "REDIS_HOST")
	envInt(&cfg.Redis.Port, "REDIS_PORT")
	envStr(&cfg.Redis.Password, "REDIS_PASSWORD")
	envInt(&cfg.Redis.DB, "REDIS_DB")

	// JWT
	envStr(&cfg.JWT.Secret, "JWT_SECRET")

	// Kafka
	envStrSlice(&cfg.Kafka.Brokers, "KAFKA_BROKERS")
	envStr(&cfg.Kafka.GroupID, "KAFKA_GROUP_ID")

	// Log
	envStr(&cfg.Log.Level, "LOG_LEVEL")
}

func envStr(target *string, key string) {
	if v := os.Getenv(key); v != "" {
		*target = v
	}
}

func envInt(target *int, key string) {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			*target = i
		}
	}
}

func envStrSlice(target *[]string, key string) {
	if v := os.Getenv(key); v != "" {
		*target = strings.Split(v, ",")
	}
}
