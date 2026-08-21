package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Env = string

const EnvFileNameVar = "ENV_FILE"

const (
	EnvLocal Env = "local"
	EnvDev   Env = "dev"
	EnvProd  Env = "prod"
)

type Config struct {
	Env      string
	TokenTTL time.Duration
	Secrets  SecretsConfig
	GRPC     GRPCConfig
	Hash     HashConfig
}

type SecretsConfig struct {
	RSAPrivateBase64 string
	RSAPublicBase64  string
	DBString         string
}

type GRPCConfig struct {
	Addr    string
	Timeout time.Duration
}

type HashConfig struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	KeyLength   uint32
	SaltLength  uint32
}

func MustLoad() Config {
	envName := getEnvFileName()

	fatalErr := func(err error) {
		log.Fatalf(`Couldn't load configuration file: "%s"; Error: %s`, envName, err.Error())
	}

	err := godotenv.Load(envName)
	if err != nil {
		fatalErr(err)
	}

	cfg, err := mapEnvToConfig()
	if err != nil {
		fatalErr(err)
	}

	return cfg
}

func getEnvFileName() string {
	var res string

	flag.StringVar(&res, "env", "", "name of .env config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv(EnvFileNameVar)

		if res == "" {
			res = defaultConfigEnv
		}
	}

	return res
}

func mapEnvToConfig() (Config, error) {
	cfg := &Config{}

	env := os.Getenv("ENV")

	switch env {
	case EnvLocal:
	case EnvDev:
	case EnvProd:
	default:
		return Config{}, fmt.Errorf(`"env" value must be: %s, %s or %s; Passed: %s`, EnvLocal, EnvDev, EnvProd, env)
	}

	cfg.Env = env

	tokenTTL := os.Getenv("TOKEN_TTL")
	if tokenTTL != "" {
		duration, err := time.ParseDuration(tokenTTL)
		if err != nil {
			return Config{}, fmt.Errorf(`"TOKEN_TTL" env var must be a valid duration: %s`, err.Error())
		}

		cfg.TokenTTL = duration
	} else {
		cfg.TokenTTL = defaultTokenTTL
	}

	err := mapSecrets(cfg)
	if err != nil {
		return Config{}, err
	}

	err = mapGRPC(cfg)
	if err != nil {
		return Config{}, err
	}

	err = mapHash(cfg)
	if err != nil {
		return Config{}, err
	}

	return *cfg, nil
}

func mapSecrets(cfg *Config) error {
	rsaPrivateBase64 := os.Getenv("RSA_PRIVATE_BASE64")
	if rsaPrivateBase64 == "" {
		return fmt.Errorf(`"RSA_PRIVATE_BASE64" env var is required`)
	}

	rsaPublicBase64 := os.Getenv("RSA_PUBLIC_BASE64")
	if rsaPublicBase64 == "" {
		return fmt.Errorf(`"RSA_PUBLIC_BASE64" env var is required`)
	}

	cfg.Secrets.RSAPrivateBase64 = rsaPrivateBase64
	cfg.Secrets.RSAPublicBase64 = rsaPublicBase64

	dbString := os.Getenv("DB_STRING")
	if dbString == "" {
		return fmt.Errorf(`"DB_STRING" env var is required`)
	}

	cfg.Secrets.DBString = dbString

	return nil
}

func mapGRPC(cfg *Config) error {
	grpcAddr := os.Getenv("GRPC_ADDR")
	if grpcAddr == "" {
		return fmt.Errorf(`"GRPC_ADDR" env var is required`)
	}

	cfg.GRPC.Addr = grpcAddr

	grpcTimeout := os.Getenv("GRPC_TIMEOUT")
	if grpcTimeout != "" {
		duration, err := time.ParseDuration(grpcTimeout)
		if err != nil {
			return fmt.Errorf(`"GRPC_TIMEOUT" env var must be a valid duration: %s`, err.Error())
		}

		cfg.GRPC.Timeout = duration
	} else {
		cfg.GRPC.Timeout = defaultGRPCTimeout
	}

	return nil
}

func mapHash(cfg *Config) error {
	hashMemory := os.Getenv("HASH_MEMORY")
	if hashMemory != "" {
		memory, err := strconv.Atoi(hashMemory)
		if err != nil {
			return fmt.Errorf(`"HASH_MEMORY" env var must be a valid integer: %s`, err.Error())
		}
		cfg.Hash.Memory = uint32(memory)
	} else {
		cfg.Hash.Memory = defaultHashMemory
	}

	hashIterations := os.Getenv("HASH_ITERATIONS")
	if hashIterations != "" {
		iterations, err := strconv.Atoi(hashIterations)
		if err != nil {
			return fmt.Errorf(`"HASH_ITERATIONS" env var must be a valid integer: %s`, err.Error())
		}
		cfg.Hash.Iterations = uint32(iterations)
	} else {
		cfg.Hash.Iterations = defaultHashIterations
	}

	hashParallelism := os.Getenv("HASH_PARALLELISM")
	if hashParallelism != "" {
		parallelism, err := strconv.Atoi(hashParallelism)
		if err != nil {
			return fmt.Errorf(`"HASH_PARALLELISM" env var must be a valid integer: %s`, err.Error())
		}
		cfg.Hash.Parallelism = uint8(parallelism)
	} else {
		cfg.Hash.Parallelism = uint8(runtime.NumCPU())
	}

	hashKeyLength := os.Getenv("HASH_KEY_LENGTH")
	if hashKeyLength != "" {
		keyLength, err := strconv.Atoi(hashKeyLength)
		if err != nil {
			return fmt.Errorf(`"HASH_KEY_LENGTH" env var must be a valid integer: %s`, err.Error())
		}
		cfg.Hash.KeyLength = uint32(keyLength)
	} else {
		cfg.Hash.KeyLength = defaultHashKeyLength
	}

	hashSaltLength := os.Getenv("HASH_SALT_LENGTH")
	if hashSaltLength != "" {
		saltLength, err := strconv.Atoi(hashSaltLength)
		if err != nil {
			return fmt.Errorf(`"HASH_SALT_LENGTH" env var must be a valid integer: %s`, err.Error())
		}
		cfg.Hash.SaltLength = uint32(saltLength)
	} else {
		cfg.Hash.SaltLength = defaultHashSaltLength
	}

	return nil
}
