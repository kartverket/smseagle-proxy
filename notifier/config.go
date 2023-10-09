package notifier

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

var cfg config

type config struct {
	Appdrift           phoneConfig `mapstructure:"app-drift"`
	InfrastrukturDrift phoneConfig `mapstructure:"infrastruktur-drift"`
}

type phoneConfig struct {
	PhoneNumber string `mapstructure:"phone-number"`
	AccessToken string `mapstructure:"access-token"`
	Url         string `mapstructure:"url"`
}

func init() {
	viper.SetConfigFile("config.yaml")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		fmt.Errorf("fatal error config file: %w", err)
	}
	err = viper.Unmarshal(&cfg)

	if err != nil {
		fmt.Errorf("something went wrong unmarshaling config: %w", err)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		fmt.Errorf("missing config: %w", err)
	}
}
