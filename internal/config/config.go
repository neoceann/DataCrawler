package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
	_ "github.com/joho/godotenv/autoload"
)

type Config struct {
	DeviceID    string `env:"DEVICE_ID"`
	XWbaasToken string `env:"WBAAS_TOKEN"`
	Wbauid1     string `env:"WBAUID_1"`
	Wbauid2     string `env:"WBAUID_2"`
	UserAgent   string `env:"USER_AGENT"`
}

type DBConfig struct {
	Host     string `env:"DB_HOST"`
	Port     int    `env:"DB_PORT"`
	User     string `env:"DB_USER"`
	Password string `env:"DB_PASSWORD"`
	DBName   string `env:"DB_NAME"`
	SSLMode  string `env:"DB_SSLMODE"`
	MaxConns int32  `env:"DB_POOL_MAX_CONNS"`
	MinConns int32  `env:"DB_POOL_MIN_CONNS"`
}

func ReadEnv(c *Config, cdb *DBConfig) error {

	if err := cleanenv.ReadEnv(c); err != nil {
		return fmt.Errorf("failed to read config")
	}
	if err := cleanenv.ReadEnv(cdb); err != nil {
		return fmt.Errorf("failed to read db_config")
	}

	return nil
}

func (c *DBConfig) GetConnectionURL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.SSLMode)
}
