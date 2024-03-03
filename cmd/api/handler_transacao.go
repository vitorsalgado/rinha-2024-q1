package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
)

// FnReturnCode represents the return code from the database.
type FnReturnCode int

const (
	FnReturnCodeSuccess FnReturnCode = iota + 1
	FnReturnCodeInsufficientBalance
	FnReturnCodeCustomerNotFound
)

const (
	TrTypeDebit  = "d"
	TrTypeCredit = "c"
)

const (
	CmdFnCrebito = "SELECT * FROM fn_crebito($1, $2, $3, $4)"
)

type Transacao struct {
	Descricao string `json:"descricao"`
	Tipo      string `json:"tipo"`
	Valor     int    `json:"valor"`
}

type Resumo struct {
	Limite int `json:"limite"`
	Saldo  int `json:"saldo"`
}

type HandlerTransacao struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func (h *HandlerTransacao) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	pid := r.PathValue("id")
	if len(pid) == 0 {
		http.Error(w, "identificador de cliente nao informado", http.StatusUnprocessableEntity)
		return
	}

	clienteid, err := strconv.Atoi(pid)
	if err != nil {
		http.Error(w, "identificador de cliente invalido", http.StatusUnprocessableEntity)
		return
	}

	limite, ok := Clientes[clienteid]
	if !ok {
		http.Error(w, "cliente nao encontrado", http.StatusNotFound)
		return
	}

	tr := Transacao{}
	err = json.NewDecoder(r.Body).Decode(&tr)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if !h.validate(&tr, w) {
		return
	}

	row := h.pool.QueryRow(r.Context(), CmdFnCrebito, pid, tr.Descricao, tr.Tipo, tr.Valor)
	fnReturnCode := FnReturnCode(0)
	result := Resumo{Limite: limite}
	if err := row.Scan(&result.Saldo, &fnReturnCode); err != nil {
		http.Error(w, "erro ao executar operacao", http.StatusInternalServerError)
		return
	}

	switch fnReturnCode {
	case FnReturnCodeSuccess:
		w.Header()[HeaderContentType] = []string{MimeApplicationJSON}
		w.WriteHeader(http.StatusOK)

		if err = json.NewEncoder(w).Encode(&result); err != nil {
			h.logger.Error("transacao: erro ao serializar a resposta", slog.Any("error", err))
		}

	case FnReturnCodeInsufficientBalance:
		http.Error(w, "saldo insuficiente", http.StatusUnprocessableEntity)
	case FnReturnCodeCustomerNotFound:
		http.Error(w, "cliente nao encontrado", http.StatusNotFound)
	default:
		http.Error(w, "estado invalido ou desconhecido", http.StatusUnprocessableEntity)
	}
}

func (h *HandlerTransacao) validate(tr *Transacao, w http.ResponseWriter) bool {
	sizedesc := len(tr.Descricao)
	if sizedesc == 0 || sizedesc > 10 {
		http.Error(w, "descricao pode conter ate 10 caracteres", http.StatusUnprocessableEntity)
		return false
	}

	if tr.Valor <= 0 {
		http.Error(w, "valor da transacao precisa ser maior que 0", http.StatusUnprocessableEntity)
		return false
	}

	if tr.Tipo != TrTypeCredit && tr.Tipo != TrTypeDebit {
		http.Error(w, "tipo da transacao precisar ser: c ou d", http.StatusUnprocessableEntity)
		return false
	}

	return true
}
