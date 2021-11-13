package transfer

type Service interface {
	GetTransfers(TransferRequest)
	DoTransfer(TransferRequest)
}

type Repository interface {
}

type service struct {
	r Repository
}

func New(r Repository) *service {
	return &service{r}
}

func (s *service) GetTransfers(TransferRequest) {

}

func (s *service) DoTransfer(TransferRequest) {

}
