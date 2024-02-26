package main

import (
	"errors"
	"fmt"
	"os"
	"time"
)

const (
	EnvSrvAddr      = "SRV_ADDR"
	EnvSrvTimeout   = "SRV_TIMEOUT"
	EnvDBConnString = "DB_CONN_STRING"
)

type Config struct {
	Addr         string
	SrvTimeout   time.Duration
	DBConnString string
}

func Parse() (Config, error) {
	config := Config{}
	config.Addr = envStr(EnvSrvAddr, ":8080")

	srvTimeout, err := envDur(EnvSrvTimeout, 15*time.Second)
	if err != nil {
		return Config{}, err
	}
	config.SrvTimeout = srvTimeout

	dbConnStr := envStr(EnvDBConnString, "")
	if len(dbConnStr) == 0 {
		return Config{}, errors.New("database connection string must not be empty")
	}
	config.DBConnString = dbConnStr

	return config, nil
}

func envStr(n, def string) string {
	str := os.Getenv(n)
	if len(str) == 0 {
		return def
	}

	return str
}

func envDur(n string, def time.Duration) (time.Duration, error) {
	str := os.Getenv(n)
	if len(str) == 0 {
		return def, nil
	}

	dur, err := time.ParseDuration(str)
	if err != nil {
		return def, fmt.Errorf("can't parse %s: %v", n, err)
	}

	return dur, nil
}
