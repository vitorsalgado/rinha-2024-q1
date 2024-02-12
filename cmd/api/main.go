package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/vitorsalgado/rinha-backend-2024-q1-go/cmd/api/config"
	"github.com/vitorsalgado/rinha-backend-2024-q1-go/internal/handler"
)

func main() {
	godotenv.Load()
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	conf, err := config.Parse()
	if err != nil {
		log.Fatal().Err(err).Msg("error parsing application config")
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	mux := http.NewServeMux()

	// lazy routing system :)
	mux.HandleFunc("/ping", handler.Pong)
	mux.HandleFunc("/clientes/{id}/transacoes", handler.ListTransactions)
	mux.HandleFunc("/clientes/{id}/extrato", handler.BankStatement)

	server := &http.Server{Handler: mux, Addr: ":8080", ReadTimeout: conf.SrvTimeout, WriteTimeout: conf.SrvTimeout}

	go func() {
		<-ctx.Done()

		c, fn := context.WithTimeout(ctx, 5*time.Second)
		defer fn()

		err := server.Shutdown(c)
		if err != nil {
			log.Error().Err(err).Msg("error during shutdown")
			return
		}

		cancel()
	}()

	log.Info().Msg("server will run on addr :8080")

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
	}
}
