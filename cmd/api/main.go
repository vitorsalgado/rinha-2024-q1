package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "go.uber.org/automaxprocs"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	conf := newConfig()
	poolConf, err := pgxpool.ParseConfig(conf.DBConnString)
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
	mux.HandleFunc("/reset", func(w http.ResponseWriter, r *http.Request) {
		batch := pgx.Batch{}
		batch.Queue("update saldos set saldo = 0")
		batch.Queue("truncate table transacoes")
		res := pool.SendBatch(r.Context(), &batch)
		if _, err := res.Exec(); err != nil {
			http.Error(w, "error ao resetar banco", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	})
	mux.Handle("POST /clientes/{id}/transacoes", &HandlerTransacao{pool: pool, logger: logger})
	mux.Handle("GET /clientes/{id}/extrato", &HandlerExtrato{pool: pool, logger: logger})

	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
	mux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
	mux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))
	mux.Handle("/debug/pprof/block", pprof.Handler("block"))

	server := &http.Server{Handler: mux, Addr: conf.Addr}

	exit := make(chan os.Signal, 1)
	signal.Notify(exit, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-exit

		c, fn := context.WithTimeout(context.Background(), 2*time.Second)
		defer fn()

		err := server.Shutdown(c)
		if err != nil {
			logger.Error("error during shutdown", slog.Any("error", err))
		}

		pool.Close()
	}()

	logger.Info("server will listen to addr: " + conf.Addr)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("shutdown", slog.Any("error", err))
	}
}
