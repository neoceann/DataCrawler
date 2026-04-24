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

func (c *Config) ReadEnv() error {

	if err := cleanenv.ReadEnv(c); err != nil {
		return fmt.Errorf("failed to read env")
	}

	return nil
}
