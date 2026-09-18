package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

type Env = string

const EnvFileNameVar = "ENV_FILE"

const (
	EnvLocal Env = "local"
	EnvDev   Env = "dev"
	EnvProd  Env = "prod"
)

type Config struct {
	ServiceName string
	Env         string
	Crypt       CryptConfig
	Auth        AuthConfig
	Store       StoreConfig
	GRPC        GRPCConfig
	HTTP        HTTPConfig
}

type HTTPConfig struct {
	Addr string
}

type GRPCConfig struct {
	Addr    string
	Timeout time.Duration
}

type AuthConfig struct {
	AccessTokenTTL              time.Duration
	RefreshTokenTTL             time.Duration
	RefreshTokenRevokedStoreTTL time.Duration
	RefreshTokenGracePeriod     time.Duration
}

type StoreConfig struct {
	PGX   *pgx.ConnConfig
	Redis *redis.Options
}

type CryptConfig struct {
	KeyRotationInterval    time.Duration
	KeyRotationGracePeriod time.Duration
	Hash                   HashConfig
}

type HashConfig struct {
	Memory           uint32
	Iterations       uint32
	Parallelism      uint8
	KeyLength        uint32
	SaltLength       uint32
	HashSecretPepper string
}

func MustLoad() Config {
	envName := getEnvFileName()

	fatalErr := func(err error) {
		log.Fatalf(`Configuration load error: %s`, err.Error())
	}

	if envName != "" {
		fatalErr := func(err error) {
			log.Fatalf(`Couldn't load .env file: "%s"; Error: %s`, envName, err.Error())
		}

		err := godotenv.Load()
		if err != nil {
			fatalErr(err)
		}
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

		return res
	}

	return res
}

func mapEnvToConfig() (Config, error) {
	cfg := &Config{}

	serviceName := os.Getenv("SERVICE_NAME")
	if serviceName == "" {
		return Config{}, fmt.Errorf(`"SERVICE_NAME" env var is required`)
	}
	cfg.ServiceName = serviceName

	env := os.Getenv("ENV")
	switch env {
	case EnvLocal:
	case EnvDev:
	case EnvProd:
	default:
		return Config{}, fmt.Errorf(`value of "ENV" var must be: %s, %s or %s; Passed: %s`, EnvLocal, EnvDev, EnvProd, env)
	}
	cfg.Env = env

	err := mapStore(cfg)
	if err != nil {
		return Config{}, err
	}

	err = mapCrypt(cfg)
	if err != nil {
		return Config{}, err
	}

	err = mapAuth(cfg)
	if err != nil {
		return Config{}, err
	}

	err = mapHTTP(cfg)
	if err != nil {
		return Config{}, err
	}

	err = mapGRPC(cfg)
	if err != nil {
		return Config{}, err
	}

	return *cfg, nil
}

func mapHTTP(cfg *Config) error {
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		return fmt.Errorf(`"HTTP_ADDR" env var is required`)
	}

	cfg.HTTP.Addr = httpAddr

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

func mapStore(cfg *Config) error {
	dbConnStr := os.Getenv("DB_CONN_STR")
	if dbConnStr == "" {
		return fmt.Errorf(`"DB_CONN_STR" env var is required`)
	}
	pgxOpts, err := pgx.ParseConfig(dbConnStr)
	if err != nil {
		return fmt.Errorf(`"DB_CONN_STR" env var must be a valid PostgreSQL connection string: %s`, err.Error())
	}
	cfg.Store.PGX = pgxOpts

	redisConnStr := os.Getenv("REDIS_CONN_STR")
	if redisConnStr == "" {
		return fmt.Errorf(`"REDIS_CONN_STR" env var is required`)
	}
	redisOpts, err := redis.ParseURL(redisConnStr)
	if err != nil {
		return fmt.Errorf(`"REDIS_CONN_STR" env var must be a valid Redis URL: %s`, err.Error())
	}
	cfg.Store.Redis = redisOpts

	return nil
}

func mapAuth(cfg *Config) error {
	accessTokenTTL, err := parseDuration("ACCESS_TOKEN_TTL", defaultAccessTokenTTL)
	if err != nil {
		return err
	}

	cfg.Auth.AccessTokenTTL = accessTokenTTL

	refreshTokenTTL, err := parseDuration("REFRESH_TOKEN_TTL", defaultRefreshTokenTTL)
	if err != nil {
		return err
	}

	cfg.Auth.RefreshTokenTTL = refreshTokenTTL

	refreshTokenRevokedStoreTTL, err := parseDuration("REFRESH_TOKEN_REVOKED_STORE_TTL", defaultRefreshTokenRevokedStoreTTL)
	if err != nil {
		return err
	}
	cfg.Auth.RefreshTokenRevokedStoreTTL = refreshTokenRevokedStoreTTL

	refreshTokenGracePeriod, err := parseDuration("REFRESH_TOKEN_GRACE_PERIOD", defaultRefreshTokenGracePeriod)
	if err != nil {
		return err
	}
	cfg.Auth.RefreshTokenGracePeriod = refreshTokenGracePeriod

	return nil
}

func mapCrypt(cfg *Config) error {
	hashMemory, err := parseUint(os.Getenv("HASH_MEMORY"), defaultHashMemory)
	if err != nil {
		return err
	}
	cfg.Crypt.Hash.Memory = uint32(hashMemory)

	hashIterations, err := parseUint(os.Getenv("HASH_ITERATIONS"), defaultHashIterations)
	if err != nil {
		return err
	}
	cfg.Crypt.Hash.Iterations = uint32(hashIterations)

	hashParallelism, err := parseUint(os.Getenv("HASH_PARALLELISM"), int64(runtime.NumCPU()))
	if err != nil {
		return err
	}
	cfg.Crypt.Hash.Parallelism = uint8(hashParallelism)

	hashKeyLength, err := parseUint(os.Getenv("HASH_KEY_LENGTH"), defaultHashKeyLength)
	if err != nil {
		return err
	}
	cfg.Crypt.Hash.KeyLength = uint32(hashKeyLength)

	hashSaltLength, err := parseUint(os.Getenv("HASH_SALT_LENGTH"), defaultHashSaltLength)
	if err != nil {
		return err
	}
	cfg.Crypt.Hash.SaltLength = uint32(hashSaltLength)

	keyRotationInterval, err := parseDuration("KEY_ROTATION_INTERVAL", defaultKeyRotationInterval)
	if err != nil {
		return err
	}
	cfg.Crypt.KeyRotationInterval = keyRotationInterval

	keyRotationGracePeriod, err := parseDuration("KEY_ROTATION_GRACE_PERIOD", defaultKeyRotationGracePeriod)
	if err != nil {
		return err
	}
	cfg.Crypt.KeyRotationGracePeriod = keyRotationGracePeriod

	return nil
}

func parseDuration(envVar string, defaultDuration time.Duration) (time.Duration, error) {
	errMsg := `"%s" env var must be a valid duration: %s`

	durationStr := os.Getenv(envVar)
	if durationStr == "" {
		if defaultDuration > 0 {
			return defaultDuration, nil
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
