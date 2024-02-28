package main

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitorsalgado/rinha-backend-2024-q1-go/internal/mod"
)

type FnReturnCode int

const (
	FnReturnCodeSuccess FnReturnCode = iota + 1
	FnReturnCodeInsufficientBalance
	FnReturnCodeCustomerNotFound
)

type HandlerTransacao struct {
	pool *pgxpool.Pool
}

func (h *HandlerTransacao) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clienteid := r.PathValue("id")
	if len(clienteid) == 0 {
		http.Error(w, "identificador de cliente nao informado", http.StatusUnprocessableEntity)
		return
	}

	tr := mod.Transacao{}
	err := json.NewDecoder(r.Body).Decode(&tr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if !h.validate(&tr, w) {
		return
	}

	row := h.pool.QueryRow(r.Context(), "SELECT * FROM fn_crebito($1, $2, $3, $4)", clienteid, tr.Descricao, tr.Tipo, tr.Valor)
	code := FnReturnCode(0)
	result := mod.Resumo{}
	if err := row.Scan(&result.Limite, &result.Saldo, &code); err != nil {
		http.Error(w, "erro ao executar operacao", http.StatusInternalServerError)
		return
	}

	switch code {
	case FnReturnCodeSuccess:
		w.Header().Add("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&result)
	case FnReturnCodeInsufficientBalance:
		http.Error(w, "saldo insuficiente", http.StatusUnprocessableEntity)
	case FnReturnCodeCustomerNotFound:
		http.Error(w, "cliente nao encontrado", http.StatusNotFound)
	default:
		http.Error(w, "estado invalido ou desconhecido", http.StatusUnprocessableEntity)
	}
}

func (h *HandlerTransacao) validate(tr *mod.Transacao, w http.ResponseWriter) bool {
	if len(tr.Descricao) == 0 ||
		len(tr.Descricao) > 10 {
		http.Error(w, "descricao pode conter ate 10 caracteres", http.StatusUnprocessableEntity)
		return false
	}

	if tr.Valor <= 0 {
		http.Error(w, "valor da transacao precisa ser maior que 0", http.StatusUnprocessableEntity)
		return false
	}

	if !(tr.Tipo == "c" || tr.Tipo == "d") {
		http.Error(w, "tipo da transacao precisar ser: c ou d", http.StatusUnprocessableEntity)
		return false
	}

	return true
}
