package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server struct {
		GRPCPort int `yaml:"GRPC_PORT" env-default:"50051"`
	}
	SubPub struct {
		BufferSize int `yaml:"buffer_size" env-default:"100"`
	}
}

func New(path string) (*Config, error) {
	var cfg Config
	if err := cleanenv.ReadConfig(path, &cfg); err != nil {
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
