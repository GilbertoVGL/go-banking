package account

type ListAccount struct {
	Name    string `json:"name"`
	Cpf     string `json:"cpf"`
	Balance int64  `json:"balance"`
}

type ListAccountsReponse struct {
	Total int64 `json:"total"`
	Page  int64 `json:"page"`
	Data  []ListAccount
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
