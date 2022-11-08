package config

import (
	"sync"
	"user_balance_service/pkg/logging"

	"github.com/ilyakaznacheev/cleanenv"
)

type HTTP struct {
	Port string `env:"PORT"  env-required:"true"`
	Host string `env:"HOST"  env-required:"true"`
}

type DBConfig struct {
	DBPort      string `env:"DB_PORT" env-required:"true"`
	DBHost      string `env:"DB_HOST" env-required:"true"`
	DBName      string `env:"DB_NAME" env-required:"true"`
	DBPassword  string `env:"DB_PASSWORD" env-required:"true"`
	DBUsername  string `env:"DB_USERNAME" env-required:"true"`
	AutoMigrate bool   `env:"AUTO_MIGRATE" env-default:"true"`
}

type Config struct {
	DBConfig
	IsDebug bool `env:"IS_DEBUG" env-default:"false"`
}

var instance *Config
var once sync.Once

func GetConfig() *Config {
	once.Do(func() {
		logger := logging.GetLogger()
		logger.Info("Read application configuration")

		instance = &Config{}
		if err := cleanenv.ReadConfig(".env", instance); err != nil {
			help, _ := cleanenv.GetDescription(instance, nil)
			logger.Info(help)
			logger.Fatal(err)
		}
	})

	return instance
}
