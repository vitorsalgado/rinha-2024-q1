package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Extrato struct {
	Saldo             ExtratoSaldo       `json:"saldo"`
	UltimasTransacoes []ExtratoTransacao `json:"ultimas_transacoes"`
}

type ExtratoSaldo struct {
	Total       int       `json:"total,omitempty"`
	DataExtrato time.Time `json:"data_extrato,omitempty"`
	Limite      int       `json:"limite,omitempty"`
}

type ExtratoTransacao struct {
	Tipo        string    `json:"tipo,omitempty"`
	Valor       int       `json:"valor,omitempty"`
	Descricao   string    `json:"descricao,omitempty"`
	RealizadaEm time.Time `json:"realizada_em,omitempty"`
}

type HandlerExtrato struct {
	logger *slog.Logger
	pool   *pgxpool.Pool
	cache  *Cache
}

func (h *HandlerExtrato) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clienteid := r.PathValue("id")
	if len(clienteid) == 0 {
		http.Error(w, "identificador de cliente nao informado", http.StatusUnprocessableEntity)
		return
	}

	if !h.cache.has(clienteid) {
		http.Error(w, "cliente nao encontrado", http.StatusNotFound)
		return
	}

	ctx := r.Context()
	conn, err := h.pool.Acquire(ctx)
	if err != nil {
		http.Error(w, "erro ao obter uma conexao com o banco de dados.", http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}

	const cmd = `
	(select s.saldo as valor, s.limite, '' as descricao, '' as tipo, now() as data
	from saldos s
	where s.cliente_id = $1)
	
	union all
	
	(select t.valor, 0, t.descricao, t.tipo, t.realizado_em
	from transacoes t
	where cliente_id = $1
	order by t.id desc
	limit 10)	
	`

	rows, err := conn.Query(ctx, cmd, clienteid)
	if err != nil {
		http.Error(w, "erro ao executar operacao", http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}
	defer rows.Close()

	rows.Next()

	saldo := ExtratoSaldo{}
	err = rows.Scan(&saldo.Total, &saldo.Limite, nil, nil, &saldo.DataExtrato)
	if err != nil {
		http.Error(w, "erro ao obter informacao de saldo", http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}

	transacoes := make([]ExtratoTransacao, 0, 10)
	for rows.Next() {
		tr := ExtratoTransacao{}
		rows.Scan(&tr.Valor, nil, &tr.Descricao, &tr.Tipo, &tr.RealizadaEm)
		transacoes = append(transacoes, tr)
	}

	extrato := Extrato{saldo, transacoes}

	w.Header().Add("content-type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&extrato)
}
