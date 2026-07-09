package config

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Env = string

const ConfigPathEnv = "CONFIG_PATH"

const (
	EnvLocal Env = "local"
	EnvDev   Env = "dev"
	EnvProd  Env = "prod"
)

type Config struct {
	Env         string        `yaml:"env" env-required:"true"`
	StoragePath string        `yaml:"storage_path" env-required:"true"`
	TokenTTL    time.Duration `yaml:"token_ttl" env-required:"true"`
}

func MustLoad() Config {
	path := getConfigPath()
	if path == "" {
		log.Fatalf(`Couldn't load configuration file: path to configuration file was not specified. Env "%s" - empty`, ConfigPathEnv)
	}

	fatalErr := func(err error) {
		log.Fatalf(`Couldn't load configuration file on path: "%s"; Error: %s`, path, err.Error())
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			fatalErr(fmt.Errorf(`Config file does not exist: "%s"`, path))
		}

		fatalErr(err)
	}

	cfg := Config{}

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		fatalErr(err)
	}

	if err := validateConfig(&cfg); err != nil {
		fatalErr(err)
	}

	return cfg
}

func getConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to yaml config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv(ConfigPathEnv)
	}

	return res
}

func validateConfig(cfg *Config) error {
	if cfg == nil {
		return errors.New(`Config has nil value`)
	}

	switch cfg.Env {
	case EnvLocal:
	case EnvDev:
	case EnvProd:
	default:
		return fmt.Errorf(`"env" value must be: %s, %s or %s; Passed: %s`, EnvLocal, EnvDev, EnvProd, cfg.Env)
	}

	return nil
}
