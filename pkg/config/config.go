package config

import (
	"github.com/go-playground/validator/v10"
	"github.com/spf13/viper"
	"log/slog"
	"os"
	"strings"
)

type ProxyConfig struct {
	AppdriftPhoneNumber string         `mapstructure:"app-drift-phone-number"`
	InfraPhoneNumber    string         `mapstructure:"infra-drift-phone-number"`
	Call                SMSEagleConfig `mapstructure:"call"`
	SMS                 SMSEagleConfig `mapstructure:"sms"`
	Debug               bool           `mapstructure:"debug"`
	BasicAuth           BasicAuth      `mapstructure:"basic-auth"`
}

type BasicAuth struct {
	Enabled  bool   `mapstructure:"enabled"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type SMSEagleConfig struct {
	AccessToken string `mapstructure:"access-token"`
	Url         string `mapstructure:"url"`
}

func Read() *ProxyConfig {
	var cfg ProxyConfig
	cfgDir := "/opt/smseagle-proxy"
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(cfgDir)
	viper.AddConfigPath(".")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	viper.SetEnvPrefix("SP")
	viper.AutomaticEnv()

	slog.Info("Looking for config", slog.String("directory", cfgDir))

	err := viper.ReadInConfig()
	if err != nil {
		slog.Error("fatal error config file", "error", err)
		os.Exit(1)
	}
	err = viper.Unmarshal(&cfg)

	if err != nil {
		slog.Error("something went wrong unmarshaling config", "error", err)
		os.Exit(1)
	}

	validate := validator.New()
	if err := validate.Struct(&cfg); err != nil {
		slog.Error("missing config", "error", err)
		os.Exit(1)
	}

	return &cfg
}
