package mod

//easyjson:json
type Transacao struct {
	Descricao string `json:"descricao"`
	Tipo      string `json:"tipo"`
	Valor     int    `json:"valor"`
}

//easyjson:json
type Resumo struct {
	Limite int `json:"limite"`
	Saldo  int `json:"saldo"`
}
