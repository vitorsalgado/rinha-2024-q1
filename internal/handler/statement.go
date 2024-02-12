package handler

import (
	"encoding/json"
	"net/http"
)

type Statement struct {
	Balance          Balance       `json:"saldo"`
	LastTransactions []Transaction `json:"ultimas_transacoes"`
}

type Balance struct {
	Total int    `json:"total,omitempty"`
	Date  string `json:"data_extrato,omitempty"`
	Limit int    `json:"limite,omitempty"`
}

type Transaction struct {
	Type        string `json:"tipo,omitempty"`
	Value       int    `json:"valor,omitempty"`
	Description string `json:"descricao,omitempty"`
	CreatedAt   string `json:"realizada_em,omitempty"`
}

func BankStatement(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Add("content-type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	tr := Statement{}
	json.NewEncoder(w).Encode(&tr)
}
