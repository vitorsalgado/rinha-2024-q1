package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

//go:generate ./out/easyjson internal/mod/transacao.go internal/mod/extrato_bancario.go

func main() {
	godotenv.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	conf, err := Parse()
	if err != nil {
		logger.Error("error parsing application config", slog.Any("error", err))
	}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s sslmode=disable", "admin", "123", "rinha", "/var/run/postgresql")
	poolConf, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		logger.Error("error parsing postgresql connection string", slog.Any("error", err))
		os.Exit(1)
		return
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), poolConf)
	if err != nil {
		logger.Error("error creating postgresql connection pool", slog.Any("error", err))
		os.Exit(1)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("pong")) })
	mux.Handle("POST /clientes/{id}/transacoes", &HandlerTransacao{pool: pool})
	mux.Handle("GET /clientes/{id}/extrato", &HandlerExtrato{pool: pool})

	server := &http.Server{Handler: mux, Addr: conf.Addr, ReadTimeout: conf.SrvTimeout, WriteTimeout: conf.SrvTimeout}

	go func() {
		<-exit

		c, fn := context.WithTimeout(context.Background(), 2*time.Second)
		defer fn()

		err := server.Shutdown(c)
		if err != nil {
			logger.Error("error during shutdown", slog.String("error", err.Error()))
			return
		}
	}()

	logger.Info("server will listen to addr: " + conf.Addr)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("shutdown", slog.String("error", err.Error()))
	}
}
