package main

import "github.com/ilyakaznacheev/cleanenv"

type PetnameConfig struct {
	Port string `yaml:"port" env:"PETNAME_GRPC_PORT" env-default:"8080"`
}

func LoadPetnameConfig(path string) (*PetnameConfig, error) {
	var cfg PetnameConfig

	if path != "" {
		if err := cleanenv.ReadConfig(path, &cfg); err != nil {
			return nil, err
		}
	}

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
