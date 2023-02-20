package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	DBDriver             string        `mapstructure:"DB_DRIVER"`
	DBSource             string        `mapstructure:"DB_SOURCE"`
	SymmetricKey         string        `mapstructure:"TOKEN_SYMMETRIC_KEY"`
	TokenDuration        time.Duration `mapstructure:"ACCESS_TOKEN_DURATION"`
	TokenRefreshDuration time.Duration `mapstructure:"REFRESH_TOKEN_DURATION"`
	Environment          string        `mapstructure:"ENVIRONMENT"`
	RedisAddr            string        `mapstructure:"REDIS_ADDRESS"`
	GmailName            string        `mapstructure:"GMAIL_NAME"`
	GmailAddress         string        `mapstructure:"GMAIL_ADDRESS"`
	GmailPassword        string        `mapstructure:"GMAIL_PASSWORD"`
}

func LoadConfig(path string) (conf Config, err error) {
	viper.AddConfigPath(path) // The path where the .env is located
	viper.SetConfigName("app")
	viper.SetConfigType("env")

	viper.AutomaticEnv() // Automcatically overwrite env values if they change
	if err = viper.ReadInConfig(); err != nil {
		return
	}

	if err = viper.Unmarshal(&conf); err != nil {
		return
	}

	return
}
