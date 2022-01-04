package transfer

type TransferRequest struct {
	Origin      uint64
	Destination uint64 `json:"destination"`
	Amount      int64  `json:"amount"`
}
