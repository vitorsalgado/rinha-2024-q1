package mod

//easyjson:json
type Transacao struct {
	Descricao string `json:"descricao"`
	Tipo      string `json:"tipo"`
	Valor     int    `json:"valor"`
}

func (t *Transacao) IsCredit() bool { return t.Tipo == "c" }

//easyjson:json
type Resumo struct {
	Limite int `json:"limite"`
	Saldo  int `json:"saldo"`
}
