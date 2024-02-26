package mod

import "time"

//easyjson:json
type Extrato struct {
	Saldo             ExtratoSaldo       `json:"saldo"`
	UltimasTransacoes []ExtratoTransacao `json:"ultimas_transacoes"`
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
