package account

import "time"

type UserId uint64

type Account struct {
	Id         uint64
	Name       string
	Cpf        string
	Balance    int64
	Secret     string
	Active     bool
	Created_at time.Duration
	Updated_at time.Duration
}

type ListAccountsReponse struct {
	Total int64         `json:"total"`
	Page  int64         `json:"page"`
	Data  []ListAccount `json:"data"`
}

type ListAccount struct {
	Id      uint64 `json:"id"`
	Name    string `json:"name"`
	Cpf     string `json:"cpf"`
	Balance int64  `json:"balance"`
}

type ListAccountQuery struct {
	PageSize int
	Offset   int
}

type BalanceRequest struct {
	ID string `json:"id"`
}

type BalanceResponse struct {
	Balance int64 `json:"balance"`
}

type NewAccountRequest struct {
	Name    string `json:"name"`
	Cpf     string `json:"cpf"`
	Secret  string `json:"secret"`
	Balance int64  `json:"balance"`
}

type AccountResponse struct {
	Name    string `json:"name"`
	Cpf     string `json:"cpf"`
	Balance int64  `json:"balance"`
}

type NewAccountError struct {
}
