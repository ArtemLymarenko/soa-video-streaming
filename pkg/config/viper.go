package config

import (
	"flag"

	"github.com/spf13/viper"
)

var (
	configPath = flag.String("configPath", "", "Path to config file")
	configName = flag.String("configName", "", "Name of config file")
)

func NewViper[T any]() (*T, error) {
	v := viper.New()

	v.SetConfigType("yaml")
	v.SetConfigName(*configName)
	v.AddConfigPath(*configPath)

	if err := v.ReadInConfig(); err != nil {
		return nil, err
	}

	var cfg T
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

type BaseHTTPServerConfig struct {
	HTTP struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"http"`
}

type BaseGRPCServerConfig struct {
	GRPC struct {
		Addr string `mapstructure:"addr"`
	} `mapstructure:"grpc"`
}
