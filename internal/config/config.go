package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	DBName   string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSLMODE"`
	MaxConns int32  `env:"DB_POOL_MAX_CONNS"`
	MinConns int32  `env:"DB_POOL_MIN_CONNS"`
	SaveToDB bool   `env:"DB_SAVE_REQ"`

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
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}
