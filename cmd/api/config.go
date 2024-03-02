package main

import (
	"os"
)

const (
	EnvAddr         = "ADDR"
	EnvDBConnString = "DB_CONN_STRING"
)

type Config struct {
	Addr         string
	DBConnString string
}

func newConfig() Config {
	config := Config{}
	config.Addr = envStr(EnvAddr, ":8080")
	config.DBConnString = envStr(EnvDBConnString, "postgresql://rinha:rinha@db:5432/rinha?sslmode=disable")

	return config
}

func envStr(n, def string) string {
	str := os.Getenv(n)
	if len(str) == 0 {
		return def
	}

	return str
}
