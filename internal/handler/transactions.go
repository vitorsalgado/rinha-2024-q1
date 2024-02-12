package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

type TransactionRes struct {
	Limit   int `json:"limite"`
	Balance int `json:"saldo"`
}

func ListTransactions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		http.Error(w, "405 method not allowed", http.StatusMethodNotAllowed)
		return
	}

	parts := strings.FieldsFunc(r.URL.Path, func(c rune) bool { return c == '/' })
	if len(parts) != 3 { // cliente/<id>/<operation>
		http.Error(w, "invalid", http.StatusUnprocessableEntity)
		return
	}

	w.Header().Add("content-type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)

	tr := TransactionRes{}
	json.NewEncoder(w).Encode(&tr)
}
