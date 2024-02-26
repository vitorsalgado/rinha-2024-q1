package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitorsalgado/rinha-backend-2024-q1-go/internal/mod"
)

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

	rows, err := h.pool.Query(r.Context(), cmd, clienteid)
	if err != nil {
		http.Error(w, "erro ao executar operacao", http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}
	defer rows.Close()

	if !rows.Next() {
		http.Error(w, "informacao do cliente nao encontrada", http.StatusNotFound)
		return
	}

	saldo := mod.ExtratoSaldo{}
	err = rows.Scan(&saldo.Total, &saldo.Limite, nil, nil, &saldo.DataExtrato)
	if err != nil {
		http.Error(w, "erro ao obter informacao de saldo", http.StatusInternalServerError)
		h.logger.Error(err.Error())
		return
	}

	transacoes := make([]mod.ExtratoTransacao, 0, 10)
	for rows.Next() {
		tr := mod.ExtratoTransacao{}
		rows.Scan(&tr.Valor, nil, &tr.Descricao, &tr.Tipo, &tr.RealizadaEm)
		transacoes = append(transacoes, tr)
	}

	extrato := mod.Extrato{Saldo: saldo, UltimasTransacoes: transacoes}

	w.Header().Add("content-type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&extrato)
}
