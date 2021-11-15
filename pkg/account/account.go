package account

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

type ListAccountsReponse struct {
	Total int64         `json:"total"`
	Page  int64         `json:"page"`
	Data  []ListAccount `json:"data"`
}

type ListAccount struct {
	Name    string `json:"name"`
	Cpf     string `json:"cpf"`
	Balance int64  `json:"balance"`
}

type ListAccountQuery struct {
	Limit    int
	PageSize int
	Offset   int
}

type BalanceRequest struct {
	ID string `json:"id"`
}

type BalanceResponce struct {
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
