package postgresdb

type Account struct {
	Name    string
	Cpf     string
	Secret  string
	Balance float64
	Active  bool
}
