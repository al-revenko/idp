package config

import (
	"time"
)

const defaultConfigEnv = ".env.local"

const defaultGRPCTimeout = 5 * time.Second

const defaultAccessTokenTTL = 15 * time.Minute
const defaultRefreshTokenTTL = 720 * time.Hour
const defaultRefreshTokenRevokedStoreTTL = 10 * time.Minute
const defaultRefreshTokenGracePeriod = 30 * time.Second

const defaultHashMemory = 2048 * 1024
const defaultHashIterations = 1
const defaultHashKeyLength = 32
const defaultHashSaltLength = 16
