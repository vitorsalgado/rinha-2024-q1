package main

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/vitorsalgado/rinha-2024-q1/internal/mod"
)

const (
	CmdExtratoQry = `
(select s.saldo as valor, s.limite, '' as descricao, '' as tipo, now() as data
from saldos s
where s.cliente_id = $1)
		
union all
		
(select t.valor, 0, t.descricao, t.tipo, t.realizado_em
from transacoes t
where cliente_id = $1
order by t.realizado_em desc
limit 10)
`
)

type HandlerExtrato struct {
	pool   *pgxpool.Pool
	logger *slog.Logger

	c [10]mod.ExtratoTransacao
}

func (h *HandlerExtrato) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	clienteid := r.PathValue("id")
	if len(clienteid) == 0 {
		http.Error(w, "identificador de cliente nao informado", http.StatusUnprocessableEntity)
		return
	}

	rows, err := h.pool.Query(r.Context(), CmdExtratoQry, clienteid)
	if err != nil {
		http.Error(w, "erro ao executar operacao", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	// moving to the first line.
	// the first line contains balance information.
	if !rows.Next() {
		http.Error(w, "informacao do cliente nao encontrada", http.StatusNotFound)
		return
	}

	extrato := mod.Extrato{}
	err = rows.Scan(&extrato.Saldo.Total, &extrato.Saldo.Limite, nil, nil, &extrato.Saldo.DataExtrato)
	if err != nil {
		http.Error(w, "erro ao obter informacao de saldo", http.StatusInternalServerError)
		return
	}

	extrato.UltimasTransacoes = h.c[:0]

	// iterate the remaining entries to get the transactions.
	for rows.Next() {
		tr := mod.ExtratoTransacao{}
		rows.Scan(&tr.Valor, nil, &tr.Descricao, &tr.Tipo, &tr.RealizadaEm)
		extrato.UltimasTransacoes = append(extrato.UltimasTransacoes, tr)
	}

	w.Header().Add("content-type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(&extrato); err != nil {
		h.logger.Error("extrato: erro ao serializar a resposta", slog.Any("error", err))
	}
}
