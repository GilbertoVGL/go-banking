package transfer

import "time"

type TransferRequest struct {
	Origin      uint64
	Destination uint64 `json:"destination"`
	Amount      int64  `json:"amount"`
}

type ListTransferQuery struct {
	PageSize int
	Page     int
}

type ListTransferReponse struct {
	Total int64          `json:"total"`
	Page  int64          `json:"page"`
	Data  []ListTransfer `json:"data"`
}

type ListTransfer struct {
	Amount          uint64    `json:"amount"`
	CreatedAt       time.Time `json:"transferDate"`
	DestinationName string    `json:"destinationName"`
	DestinationCpf  string    `json:"destinationCpf"`
	OriginName      string    `json:"originName"`
	OriginCpf       string    `json:"originCpf"`
}
