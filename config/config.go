package config

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
)

type ProxyConfig struct {
	AppdriftPhoneNumber string         `mapstructure:"app-drift-phone-number"`
	InfraPhoneNumber    string         `mapstructure:"infra-drift-phone-number"`
	Call                SMSEagleConfig `mapstructure:"call"`
	SMS                 SMSEagleConfig `mapstructure:"sms"`
}

type SMSEagleConfig struct {
	AccessToken string `mapstructure:"access-token"`
	Url         string `mapstructure:"url"`
}

func Read() *ProxyConfig {
	var cfg ProxyConfig
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

	return &cfg
}
