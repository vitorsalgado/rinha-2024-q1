package main

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Extrato struct {
	Saldo             ExtratoSaldo       `json:"saldo"`
	UltimasTransacoes []ExtratoTransacao `json:"ultimas_transacoes"`

	_c [10]ExtratoTransacao
}

type ExtratoSaldo struct {
	Total       int       `json:"total"`
	DataExtrato time.Time `json:"data_extrato"`
	Limite      int       `json:"limite"`
}

type ExtratoTransacao struct {
	Tipo        string    `json:"tipo"`
	Valor       int       `json:"valor"`
	Descricao   string    `json:"descricao"`
	RealizadaEm time.Time `json:"realizada_em"`
}

const (
	CmdExtratoQry = `
(select s.saldo as v, '' as d, '' as t, now() as d
from saldos s
where s.cliente_id = $1)
		
union all
		
(select t.valor, t.descricao, t.tipo, t.realizado_em
from transacoes t
where t.cliente_id = $1
order by t.id desc
limit 10)
`
)

type HandlerExtrato struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func (h *HandlerExtrato) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	extrato := Extrato{}
	extrato.Saldo.Limite = limite

	err = rows.Scan(&extrato.Saldo.Total, nil, nil, &extrato.Saldo.DataExtrato)
	if err != nil {
		http.Error(w, "erro ao obter informacao de saldo", http.StatusInternalServerError)
		return
	}

	extrato.UltimasTransacoes = extrato._c[:0]

	// iterate the remaining entries to get the transactions.
	for rows.Next() {
		tr := ExtratoTransacao{}
		rows.Scan(&tr.Valor, &tr.Descricao, &tr.Tipo, &tr.RealizadaEm)
		extrato.UltimasTransacoes = append(extrato.UltimasTransacoes, tr)
	}

	w.Header()[HeaderContentType] = []string{MimeApplicationJSON}
	w.WriteHeader(http.StatusOK)

	if err = json.NewEncoder(w).Encode(&extrato); err != nil {
		h.logger.Error("extrato: erro ao serializar a resposta", slog.Any("error", err))
	}
}
