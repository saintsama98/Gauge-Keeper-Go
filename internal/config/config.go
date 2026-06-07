// Package config loads keeper settings from environment variables. Secrets
// (the curator private key) are referenced by env-var NAME only — never stored
// here — so the key is read at runtime from a separate variable.
package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	RPCURL       string
	VaultAddress string
	ChainID      int64
	KeyEnv       string // name of the env var holding the curator private key
	DryRun       bool

	PollInterval         time.Duration
	HarvestInterval      time.Duration
	GuardInterval        time.Duration
	RebalanceMinInterval time.Duration
}

// Load reads configuration from the environment, applying sensible defaults.
func Load() *Config {
	return &Config{
		RPCURL:               env("KEEPER_RPC_URL", ""),
		VaultAddress:         env("KEEPER_VAULT_ADDRESS", ""),
		ChainID:              envInt("KEEPER_CHAIN_ID", 42161), // Arbitrum One
		KeyEnv:               env("KEEPER_KEY_ENV", "KEEPER_PRIVATE_KEY"),
		DryRun:               envBool("KEEPER_DRY_RUN", true),
		PollInterval:         envDur("KEEPER_POLL_INTERVAL", 30*time.Second),
		HarvestInterval:      envDur("KEEPER_HARVEST_INTERVAL", 6*time.Hour),
		GuardInterval:        envDur("KEEPER_GUARD_INTERVAL", 5*time.Minute),
		RebalanceMinInterval: envDur("KEEPER_REBALANCE_MIN_INTERVAL", time.Hour),
	}
}

func env(k, def string) string {
	if v, ok := os.LookupEnv(k); ok {
		return v
	}
	return def
}

func envInt(k string, def int64) int64 {
	if v, ok := os.LookupEnv(k); ok {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil {
			return n
		}
	}
	return def
}

func envBool(k string, def bool) bool {
	if v, ok := os.LookupEnv(k); ok {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return def
}

func envDur(k string, def time.Duration) time.Duration {
	if v, ok := os.LookupEnv(k); ok {
		if d, err := time.ParseDuration(v); err == nil {
			return d
		}
	}
	return def
}
