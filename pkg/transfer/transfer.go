package transfer

type TransferRequest struct {
	Destination string `json:"destination"`
	Amount      int64  `json:"amount"`
}
