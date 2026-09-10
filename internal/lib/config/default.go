package config

import (
	"time"
)

const defaultAppName = "IdP Service"
const defaultConfigEnv = ".env.local"

const defaultGRPCTimeout = 5 * time.Second

const defaultHashMemory = 2048 * 1024
const defaultHashIterations = 1
const defaultHashKeyLength = 32
const defaultHashSaltLength = 16
