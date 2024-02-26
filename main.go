package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	conf, err := Parse()
	if err != nil {
		logger.Error("error parsing application config", slog.Any("error", err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	poolConf, err := pgxpool.ParseConfig(conf.DBConnString)
	if err != nil {
		logger.Error("error parsing postgresql connection string", slog.Any("error", err))
		os.Exit(1)
		return
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConf)
	if err != nil {
		logger.Error("error creating postgresql connection pool", slog.Any("error", err))
		os.Exit(1)
		return
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte("pong")) })
	mux.Handle("POST /clientes/{id}/transacoes", &HandlerTransacao{logger, pool})

	server := &http.Server{Handler: mux, Addr: ":8080", ReadTimeout: conf.SrvTimeout, WriteTimeout: conf.SrvTimeout}

	go func() {
		<-ctx.Done()
		defer cancel()

		c, fn := context.WithTimeout(ctx, 5*time.Second)
		defer fn()

		err := server.Shutdown(c)
		if err != nil {
			logger.Error("error during shutdown", slog.String("error", err.Error()))
			return
		}
	}()

	logger.Info("server addr :8080")

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("error during server close",
			slog.String("error", err.Error()))
	}
}
