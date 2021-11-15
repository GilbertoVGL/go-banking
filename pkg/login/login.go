package login

import "time"

type Account struct {
	Id         string
	Name       string
	Cpf        string
	Balance    int64
	Secret     string
	Active     bool
	Created_at time.Duration
	Updated_at time.Duration
}

type LoginRequest struct {
	Cpf    string `json:"cpf"`
	Secret string `json:"secret"`
}

type LoginReponse struct {
	Token string `json:"token"`
}
