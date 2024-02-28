package main

import (
	"os"
	"time"
)

const (
	EnvAddr         = "ADDR"
	EnvDBConnString = "DB_CONN_STRING"
)

type Config struct {
	Addr         string
	SrvTimeout   time.Duration
	DBConnString string
}

func Parse() (Config, error) {
	config := Config{}
	config.Addr = envStr(EnvAddr, ":8080")
	config.SrvTimeout = 10 * time.Second
	config.DBConnString = envStr(EnvDBConnString, "postgresql://rinha:rinha@db:5432/rinha?sslmode=disable")

	return config, nil
}

func envStr(n, def string) string {
	str := os.Getenv(n)
	if len(str) == 0 {
		return def
	}

	return str
}
