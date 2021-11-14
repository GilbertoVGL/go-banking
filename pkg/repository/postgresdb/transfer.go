package postgresdb

type Transfer struct {
	account_origin_id      uint
	account_destination_id uint
	amount                 float64
}
