package login

type LoginRequest struct {
	Cpf    string `json:"cpf"`
	Secret string `json:"secret"`
}

type LoginReponse struct {
	Token string `json:"token"`
}
