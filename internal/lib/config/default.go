package config

import (
	"time"
)

const defaultConfigEnv = ".env.local"
const defaultTokenTTL = 60 * time.Minute
const defaultGRPCTimeout = 5 * time.Second

const defaultHashMemory = 2048 * 1024
const defaultHashIterations = 1
const defaultHashKeyLength = 32
const defaultHashSaltLength = 16
