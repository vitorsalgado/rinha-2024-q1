package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type FnReturnCode int

const (
	FnReturnCodeSuccess FnReturnCode = iota + 1
	FnReturnCodeInsufficientBalance
)

func (c FnReturnCode) String() string {
	switch c {
	case FnReturnCodeInsufficientBalance:
		return "saldo insuficiente"
	default:
		return "estado invalido ou desconhecido"
	}
}

type Transacao struct {
	Descricao string `json:"descricao"`
	Tipo      string `json:"tipo"`
	Valor     int    `json:"valor"`
}

func (t *Transacao) isCredit() bool { return t.Tipo == "c" }

type Resumo struct {
	Limite int `json:"limite"`
	Saldo  int `json:"saldo"`
}

type HandlerTransacao struct {
	logger *slog.Logger
	pool   *pgxpool.Pool
	cache  *Cache
}

func (h *HandlerTransacao) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clienteid := r.PathValue("id")
	if len(clienteid) == 0 {
		http.Error(w, "identificador de cliente nao informado", http.StatusUnprocessableEntity)
		return
	}

	if !h.cache.has(clienteid) {
		http.Error(w, "cliente nao encontrado", http.StatusNotFound)
		return
	}

	var mod Transacao
	err := json.NewDecoder(r.Body).Decode(&mod)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if !h.validate(&mod, w) {
		return
	}

	ctx := r.Context()
	conn, err := h.pool.Acquire(ctx)
	if err != nil {
		http.Error(w, "erro ao obter uma conexao com o banco de dados.", http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}

	operation := ""
	if mod.isCredit() {
		operation = "SELECT * FROM creditar($1, $2, $3)"
	} else {
		operation = "SELECT * FROM debitar($1, $2, $3)"
	}

	row := conn.QueryRow(ctx, operation, clienteid, mod.Descricao, mod.Valor)
	code := FnReturnCode(0)
	result := Resumo{}
	if err := row.Scan(&result.Limite, &result.Saldo, &code); err != nil {
		http.Error(w, "erro ao executar operacao", http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}

	switch code {
	case FnReturnCodeSuccess:
		w.Header().Add("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&result)
	default:
		http.Error(w, code.String(), http.StatusUnprocessableEntity)
	}
}

func (h *HandlerTransacao) validate(mod *Transacao, w http.ResponseWriter) bool {
	if len(mod.Descricao) == 0 ||
		len(mod.Descricao) > 10 {
		http.Error(w, "descricao pode conter ate 10 caracteres", http.StatusUnprocessableEntity)
		return false
	}

	if mod.Valor <= 0 {
		http.Error(w, "valor da transacao precisa ser maior que 0", http.StatusUnprocessableEntity)
		return false
	}

	if !(mod.Tipo == "c" || mod.Tipo == "d") {
		http.Error(w, "tipo da transacao precisar ser: c ou d", http.StatusUnprocessableEntity)
		return false
	}

	return true
}
