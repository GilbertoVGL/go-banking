package rest

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/login"
	"github.com/GilbertoVGL/go-banking/pkg/transfer"
)

type mockRepository struct{}

var mockListAccount func() (account.ListAccountsReponse, error)
var mockAddAccount func(account.NewAccountRequest) error
var mockLogin func(login.LoginRequest) error

func (mr *mockRepository) ListAccount(params account.ListAccountQuery) (account.ListAccountsReponse, error) {
	return mockListAccount()
}
func (mr *mockRepository) AddAccount(a account.NewAccountRequest) error {
	return mockAddAccount(a)
}
func (mr *mockRepository) GetAccountBySecretAndCPF(l login.LoginRequest) error {
	return mockLogin(l)
}

type mockService struct {
	r *mockRepository
}

func (ms *mockService) List(a account.ListAccountQuery) (account.ListAccountsReponse, error) {
	return ms.r.ListAccount(a)
}
func (ms *mockService) NewAccount(a account.NewAccountRequest) error {
	return ms.r.AddAccount(a)
}
func (ms *mockService) GetBalance(a account.BalanceRequest) {

}
func (ms *mockService) LoginUser(l login.LoginRequest) (login.LoginReponse, error) {
	return login.LoginReponse{}, ms.r.GetAccountBySecretAndCPF(l)
}
func (ms *mockService) GetTransfers(transfer.TransferRequest) {
}
func (ms *mockService) DoTransfer(transfer.TransferRequest) {
}

func TestDoLoginIsOk(t *testing.T) {
	path := url.URL{
		Path: "/login",
	}
	payload := bytes.NewBuffer([]byte(`{"cpf":"472.081.640-10","secret":"secret_pass"}`))

	req, err := http.NewRequest(http.MethodPost, path.String(), payload)
	if err != nil {
		t.Fatal(err)
	}

	mockLogin = func(l login.LoginRequest) error {
		return nil
	}

	r := &mockRepository{}
	s := mockService{r}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(doLogin(&s))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := strings.Trim(`{"token":""}`, " \r\n")
	body := strings.Trim(rr.Body.String(), " \r\n")

	if body != expected {
		t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
			body, expected)
	}
}

func TestDoLoginServiceError(t *testing.T) {
	path := url.URL{
		Path: "/login",
	}
	payload := bytes.NewBuffer([]byte(`{"cpf":"472.081.640-10","secret":"secret_pass"}`))

	req, err := http.NewRequest(http.MethodPost, path.String(), payload)
	if err != nil {
		t.Fatal(err)
	}

	mockLogin = func(l login.LoginRequest) error {
		return errors.New("bad_test")
	}

	r := &mockRepository{}
	s := mockService{r}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(doLogin(&s))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestDoTransferIsOk(t *testing.T) {
	path := url.URL{
		Path: "/transfers",
	}
	payload := bytes.NewBuffer([]byte(`{"destination": "1", "amount": 2}`))

	req, err := http.NewRequest(http.MethodPost, path.String(), payload)
	if err != nil {
		t.Fatal(err)
	}

	r := &mockRepository{}
	s := mockService{r}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(doTransfer(&s))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestGetTransferIsOk(t *testing.T) {
	path := url.URL{
		Path: "/transfers",
	}

	req, err := http.NewRequest(http.MethodGet, path.String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	r := &mockRepository{}
	s := mockService{r}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getTransfer(&s))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func TestListAccountsIsOk(t *testing.T) {
	path := url.URL{
		Path:     "/accounts",
		RawQuery: (&url.Values{"limit": []string{"10"}, "pageSize": []string{"10"}, "offset": []string{"0"}}).Encode(),
	}

	req, err := http.NewRequest(http.MethodGet, path.String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	mockListAccount = func() (account.ListAccountsReponse, error) {
		accounts := []account.ListAccount{}
		return account.ListAccountsReponse{
			Total: 0,
			Page:  0,
			Data:  accounts,
		}, nil
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

func TestListAccountsInvalidQuery(t *testing.T) {
	path := url.URL{
		Path:     "/accounts",
		RawQuery: (&url.Values{"limit": []string{"really"}, "pageSize": []string{"bad"}, "offset": []string{"params"}}).Encode(),
	}

	req, err := http.NewRequest(http.MethodGet, path.String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	r := &mockRepository{}
	s := mockService{r}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(listAccounts(&s))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := strings.Trim(`{"error":"invalid query params: limit, pageSize, offset"}`, " \r\n")
	body := strings.Trim(rr.Body.String(), " \r\n")

	if body != expected {
		t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
			body, expected)
	}
}

func TestListAccountsServiceReturnError(t *testing.T) {
	path := url.URL{
		Path:     "/accounts",
		RawQuery: (&url.Values{"limit": []string{"10"}, "pageSize": []string{"10"}, "offset": []string{"10"}}).Encode(),
	}

	req, err := http.NewRequest(http.MethodGet, path.String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	mockListAccount = func() (account.ListAccountsReponse, error) {
		return account.ListAccountsReponse{}, errors.New("unable to access database")
	}

	r := &mockRepository{}
	s := mockService{r}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(listAccounts(&s))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusInternalServerError {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := strings.Trim(`{"error":"unable to access database"}`, " TestListAccountsIsOk\r\n")
	body := strings.Trim(rr.Body.String(), " \r\n")

	if body != expected {
		t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
			body, expected)
	}
}

func TestNewAccountIsOk(t *testing.T) {
	path := url.URL{
		Path: "/accounts",
	}
	payload := bytes.NewBuffer([]byte(`{"name":"Mané","cpf":"610.781.580-53","secret":"secret_pass"}`))

	req, err := http.NewRequest(http.MethodPost, path.String(), payload)
	if err != nil {
		t.Fatal(err)
	}

	mockAddAccount = func(a account.NewAccountRequest) error {
		return nil
	}

	r := &mockRepository{}
	s := mockService{r}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(newAccount(&s))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := strings.Trim(`{"msg":"account created"}`, " \r\n")
	body := strings.Trim(rr.Body.String(), " \r\n")

	if body != expected {
		t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
			body, expected)
	}
}

func TestNewAccountServiceError(t *testing.T) {
	path := url.URL{
		Path: "/accounts",
	}
	payload := bytes.NewBuffer([]byte(`{"name":"Mané","cpf":"999.666.999-66","secret":"secret_pass"}`))

	req, err := http.NewRequest(http.MethodPost, path.String(), payload)
	if err != nil {
		t.Fatal(err)
	}

	mockAddAccount = func(a account.NewAccountRequest) error {
		return errors.New("invalid input: cpf")
	}

	r := &mockRepository{}
	s := mockService{r}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(newAccount(&s))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := strings.Trim(`{"error":"invalid input: cpf"}`, " \r\n")
	body := strings.Trim(rr.Body.String(), " \r\n")

	if body != expected {
		t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
			body, expected)
	}
}

func TestGetBalanceIsOk(t *testing.T) {
	path := url.URL{
		Path: "/accounts/123/balance",
	}

	req, err := http.NewRequest(http.MethodGet, path.String(), nil)
	if err != nil {
		t.Fatal(err)
	}

	r := &mockRepository{}
	s := mockService{r}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(getBalance(&s))

	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	expected := strings.Trim(`{"balance":0}`, " \r\n")
	body := strings.Trim(rr.Body.String(), " \r\n")

	if body != expected {
		t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
			body, expected)
	}
}
