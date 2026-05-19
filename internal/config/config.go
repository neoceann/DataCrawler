package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DBHost             string `env:"DB_HOST"`
	DBPort             int    `env:"DB_PORT"`
	DBUser             string `env:"DB_USER"`
	DBPassword         string `env:"DB_PASSWORD"`
	DBName             string `env:"DB_NAME"`
	DBSSLMode          string `env:"DB_SSLMODE"`
	DBMaxConns         int32  `env:"DB_POOL_MAX_CONNS"`
	DBMinConns         int32  `env:"DB_POOL_MIN_CONNS"`
	SaveToDB           bool   `env:"DB_SAVE_REQ"`
	RedisAddr          string `env:"REDIS_ADDR"`
	RedisCacheDuration int32  `env:"REDIS_CACHE_TIME_MINUTE"`

	WBDeviceID    string `env:"WB_DEVICE_ID"`
	WBXWbaasToken string `env:"WB_WBAAS_TOKEN"`
	WBWbauid1     string `env:"WB_WBAUID_1"`
	WBWbauid2     string `env:"WB_WBAUID_2"`

	UserAgent string `env:"USER_AGENT"`
}

func ReadEnv(c *Config) error {

	if err := cleanenv.ReadEnv(c); err != nil {
		return fmt.Errorf("failed to read config")
	}

	return nil
}

func (c *Config) GetConnectionURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName, c.DBSSLMode)
}
