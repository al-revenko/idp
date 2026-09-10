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
	AppName string
	Env     string
	Secrets SecretsConfig
	GRPC    GRPCConfig
}

type GRPCConfig struct {
	Addr    string
	Timeout time.Duration
}

type SecretsConfig struct {
	RSAPrivateBase64 string
	RSAPublicBase64  string
	DBString         string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	Hash             HashConfig
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

	appName := os.Getenv("APP_NAME")
	if appName == "" {
		appName = defaultAppName
	}
	cfg.AppName = appName

	env := os.Getenv("ENV")
	switch env {
	case EnvLocal:
	case EnvDev:
	case EnvProd:
	default:
		return Config{}, fmt.Errorf(`"env" value must be: %s, %s or %s; Passed: %s`, EnvLocal, EnvDev, EnvProd, env)
	}
	cfg.Env = env

	err := mapSecrets(cfg)
	if err != nil {
		return Config{}, err
	}

	err = mapGRPC(cfg)
	if err != nil {
		return Config{}, err
	}

	return *cfg, nil
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

	accessTokenTTL, err := parseDuration("ACCESS_TOKEN_TTL", nil)
	if err != nil {
		return err
	}

	cfg.Secrets.AccessTokenTTL = accessTokenTTL

	refreshTokenTTL, err := parseDuration("REFRESH_TOKEN_TTL", nil)
	if err != nil {
		return err
	}

	cfg.Secrets.RefreshTokenTTL = refreshTokenTTL

	err = mapHash(cfg)
	if err != nil {
		return err
	}

	return nil
}

func mapHash(cfg *Config) error {
	hashMemory, err := parseUint(os.Getenv("HASH_MEMORY"), defaultHashMemory)
	if err != nil {
		return err
	}
	cfg.Secrets.Hash.Memory = uint32(hashMemory)

	hashIterations, err := parseUint(os.Getenv("HASH_ITERATIONS"), defaultHashIterations)
	if err != nil {
		return err
	}
	cfg.Secrets.Hash.Iterations = uint32(hashIterations)

	hashParallelism, err := parseUint(os.Getenv("HASH_PARALLELISM"), int64(runtime.NumCPU()))
	if err != nil {
		return err
	}
	cfg.Secrets.Hash.Parallelism = uint8(hashParallelism)

	hashKeyLength, err := parseUint(os.Getenv("HASH_KEY_LENGTH"), defaultHashKeyLength)
	if err != nil {
		return err
	}
	cfg.Secrets.Hash.KeyLength = uint32(hashKeyLength)

	hashSaltLength, err := parseUint(os.Getenv("HASH_SALT_LENGTH"), defaultHashSaltLength)
	if err != nil {
		return err
	}
	cfg.Secrets.Hash.SaltLength = uint32(hashSaltLength)

	return nil
}

func parseDuration(envVar string, defaultDuration *time.Duration) (time.Duration, error) {
	errMsg := `"%s" env var must be a valid duration: %s`

	durationStr := os.Getenv(envVar)
	if durationStr == "" {
		if defaultDuration != nil {
			return *defaultDuration, nil
		}

		return 0, fmt.Errorf(errMsg, envVar, `env empty`)
	}

	duration, err := time.ParseDuration(durationStr)
	if err != nil {
		return 0, fmt.Errorf(errMsg, envVar, err.Error())
	}

	return duration, nil
}

func parseUint(envVar string, defaultValue int64) (uint64, error) {
	errMsg := `"%s" env var must be a valid integer: %s`

	valueStr := os.Getenv(envVar)
	if valueStr == "" {
		if defaultValue >= 0 {
			return uint64(defaultValue), nil
		}

		return 0, fmt.Errorf(errMsg, envVar, `env empty`)
	}

	value, err := strconv.ParseUint(valueStr, 10, 32)
	if err != nil {
		return 0, fmt.Errorf(errMsg, envVar, err.Error())
	}

	return value, nil
}
