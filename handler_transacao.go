package main

import (
	"encoding/json"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Transacao struct {
	Descricao string `json:"descricao"`
	Tipo      string `json:"tipo"`
	Valor     int    `json:"valor"`
}

func (t *Transacao) isDebit() bool  { return t.Tipo == "d" }
func (t *Transacao) isCredit() bool { return t.Tipo == "c" }

type Resumo struct {
	Limite int `json:"limite"`
	Saldo  int `json:"saldo"`
}

type HandlerTransacao struct {
	pool *pgxpool.Pool
}

func (h *HandlerTransacao) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clienteid := r.PathValue("id")
	if len(clienteid) == 0 {
		http.Error(w, "customer id must be present", http.StatusUnprocessableEntity)
		return
	}

	var transation Transacao
	err := json.NewDecoder(r.Body).Decode(&transation)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnprocessableEntity)
		return
	}

	if len(transation.Descricao) == 0 ||
		len(transation.Descricao) > 10 {
		http.Error(w, "description must have max 10 chars", http.StatusUnprocessableEntity)
		return
	}

	if transation.Valor <= 0 {
		http.Error(w, "value must greater than 0", http.StatusUnprocessableEntity)
		return
	}

	ctx := r.Context()
	conn, err := h.pool.Acquire(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// credito
	if transation.Tipo == "c" {
		row := conn.QueryRow(ctx,
			"SELECT * FROM creditar($1, $2, $3)",
			clienteid,
			transation.Descricao,
			transation.Valor,
		)

		res := Resumo{}
		if err := row.Scan(&res.Limite, &res.Saldo); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Add("content-type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(&res)
		return
	}

	// debito
	if transation.Tipo == "d" {
		row := conn.QueryRow(ctx,
			"SELECT * FROM debitar($1, $2, $3)",
			clienteid,
			transation.Descricao,
			transation.Valor,
		)

		res := Resumo{}
		code := 0
		if err := row.Scan(&res.Limite, &res.Saldo, &code); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		if code == 0 {
			w.Header().Add("content-type", "application/json; charset=utf-8")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(&res)
			return
		}

		http.Error(w, "saldo", http.StatusUnprocessableEntity)
		return
	}

	http.Error(w, "not implemented", http.StatusNotImplemented)
}
