package rest

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/GilbertoVGL/go-banking/pkg/account"
)

type mockRepository struct{}

func (mr *mockRepository) ListAccount() (account.ListAccountsReponse, error) {
	accounts := []account.ListAccount{}
	return account.ListAccountsReponse{
		Total: 0,
		Page:  0,
		Data:  accounts,
	}, nil
}
func (mr *mockRepository) AddAccount(a account.NewAccountRequest) error {
	return nil
}

type mockService struct {
	r *mockRepository
}

func (ms *mockService) List(a account.ListAccountQuery) (account.ListAccountsReponse, error) {
	return ms.r.ListAccount()
}
func (ms *mockService) NewAccount(a account.NewAccountRequest) error {
	return nil
}
func (ms *mockService) GetBalance(a account.BalanceRequest) {

}

func TestListAccountsIsOk(t *testing.T) {
	path := url.URL{
		Path:     "/accounts",
		RawQuery: (&url.Values{"limit": []string{"10"}, "pageSize": []string{"10"}, "offset": []string{"0"}}).Encode(),
	}

	req, err := http.NewRequest("GET", path.String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	r := &mockRepository{}
	s := mockService{r}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(listAccounts(&s))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := strings.Trim(`{"total":0,"page":0,"data":[]}`, " \r\n")
	body := strings.Trim(rr.Body.String(), " \r\n")

	if body != expected {
		t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
			body, expected)
	}
}
