package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/http/rest/middleware"
	"github.com/GilbertoVGL/go-banking/pkg/login"
	"github.com/GilbertoVGL/go-banking/pkg/transfer"
	"github.com/gorilla/mux"
)

type mockRepository struct{}

var mockListAccount func() (account.ListAccountsReponse, error)
var mockAddAccount func(account.NewAccountRequest) error
var mockLogin func(context.Context, login.LoginRequest) (login.Account, error)
var mockGetAccountBalance func(uint64) (account.BalanceResponse, error)
var mockGetTransfer func(uint64, transfer.ListTransferQuery) (transfer.ListTransferReponse, error)

func (mr *mockRepository) ListAccount(params account.ListAccountQuery) (account.ListAccountsReponse, error) {
	return mockListAccount()
}
func (mr *mockRepository) AddAccount(a account.NewAccountRequest) error {
	return mockAddAccount(a)
}
func (mr *mockRepository) GetAccountBySecretAndCPF(ctx context.Context, l login.LoginRequest) (login.Account, error) {
	return mockLogin(ctx, l)
}
func (mr *mockRepository) GetAccountBalance(a uint64) (account.BalanceResponse, error) {
	return mockGetAccountBalance(a)
}
func (mr *mockRepository) GetTransfers(a uint64, l transfer.ListTransferQuery) (transfer.ListTransferReponse, error) {
	return mockGetTransfer(a, l)
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
func (ms *mockService) GetBalance(a uint64) (account.BalanceResponse, error) {
	return ms.r.GetAccountBalance(a)
}
func (ms *mockService) LoginUser(ctx context.Context, l login.LoginRequest) (login.LoginReponse, error) {
	account, err := ms.r.GetAccountBySecretAndCPF(ctx, l)
	return login.LoginReponse{Token: account.Cpf}, err
}
func (ms *mockService) GetTransfers(a uint64, l transfer.ListTransferQuery) (transfer.ListTransferReponse, error) {
	return ms.r.GetTransfers(a, l)
}
func (ms *mockService) DoTransfer(transfer.TransferRequest) error {
	return nil
}

func TestDoLogin(t *testing.T) {
	path := url.URL{
		Path: "/login",
	}
	l := login.LoginRequest{
		Cpf:    "472.081.640-10",
		Secret: "secret_pass",
	}
	r := &mockRepository{}
	s := mockService{r}
	jsonPayload, _ := json.Marshal(l)
	payload := bytes.NewBuffer(jsonPayload)

	t.Run("doLogin is OK", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, path.String(), payload)
		if err != nil {
			t.Fatal(err)
		}

		mockLogin = func(ctx context.Context, l login.LoginRequest) (login.Account, error) {
			return login.Account{}, nil
		}

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
	})

	t.Run("doLogin service Error", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, path.String(), payload)
		if err != nil {
			t.Fatal(err)
		}

		mockLogin = func(ctx context.Context, l login.LoginRequest) (login.Account, error) {
			return login.Account{}, errors.New("bad_test")
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(doLogin(&s))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
}

func TestDoTransfer(t *testing.T) {
	path := url.URL{
		Path: "/transfers",
	}
	var d uint64 = 1
	var a int64 = 2
	q := transfer.TransferRequest{
		Destination: &d,
		Amount:      &a,
	}
	r := &mockRepository{}
	s := mockService{r}
	jsonPayload, _ := json.Marshal(q)
	payload := bytes.NewBuffer(jsonPayload)

	t.Run("doTransfer is OK", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodPost, path.String(), payload)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(doTransfer(&s))
		ctx := req.Context()
		ctx = context.WithValue(ctx, middleware.UserIdContextKey("userId"), uint64(1))
		ro := req.Clone(ctx)

		handler.ServeHTTP(rr, ro)

		if status := rr.Code; status != http.StatusCreated {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
}

func TestGetTransferIsOk(t *testing.T) {
	path := url.URL{
		Path:     "/transfers",
		RawQuery: (&url.Values{"pageSize": []string{"10"}, "page": []string{"0"}}).Encode(),
	}
	r := &mockRepository{}
	s := mockService{r}

	t.Run("getTransfer is OK", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, path.String(), nil)
		if err != nil {
			t.Fatal(err)
		}

		mockGetTransfer = func(a uint64, l transfer.ListTransferQuery) (transfer.ListTransferReponse, error) {
			transfers := []transfer.ListTransfer{}
			return transfer.ListTransferReponse{
				Total: 0,
				Page:  0,
				Data:  transfers,
			}, nil
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(getTransfer(&s))
		ctx := req.Context()
		ctx = context.WithValue(ctx, middleware.UserIdContextKey("userId"), uint64(1))
		ro := req.Clone(ctx)

		handler.ServeHTTP(rr, ro)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}
	})
}

func TestListAccounts(t *testing.T) {
	path := url.URL{
		Path: "/accounts",
	}
	r := &mockRepository{}
	s := mockService{r}

	t.Run("listAccounts is OK", func(t *testing.T) {
		path.RawQuery = (&url.Values{"pageSize": []string{"10"}, "page": []string{"0"}}).Encode()
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
	})

	t.Run("listAccounts invalid query", func(t *testing.T) {
		path.RawQuery = (&url.Values{"pageSize": []string{"bad"}, "page": []string{"params"}}).Encode()
		req, err := http.NewRequest(http.MethodGet, path.String(), nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(listAccounts(&s))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusOK)
		}

		expected := strings.Trim(`{"error":"invalid query params: pageSize, page"}`, " \r\n")
		body := strings.Trim(rr.Body.String(), " \r\n")

		if body != expected {
			t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
				body, expected)
		}
	})

	t.Run("listAccounts service error", func(t *testing.T) {
		path.RawQuery = (&url.Values{"pageSize": []string{"10"}, "page": []string{"10"}}).Encode()
		req, err := http.NewRequest(http.MethodGet, path.String(), nil)
		if err != nil {
			t.Fatal(err)
		}

		mockListAccount = func() (account.ListAccountsReponse, error) {
			return account.ListAccountsReponse{}, errors.New("unable to access database")
		}

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
	})
}

func TestNewAccount(t *testing.T) {
	path := url.URL{
		Path: "/accounts",
	}
	a := account.NewAccountRequest{
		Name:   "Mané",
		Cpf:    "610.781.580-53",
		Secret: "secret_pass",
	}
	r := &mockRepository{}
	s := mockService{r}

	t.Run("newAccount is OK", func(t *testing.T) {
		jsonPayload, _ := json.Marshal(a)
		payload := bytes.NewBuffer(jsonPayload)
		req, err := http.NewRequest(http.MethodPost, path.String(), payload)
		if err != nil {
			t.Fatal(err)
		}

		mockAddAccount = func(a account.NewAccountRequest) error {
			return nil
		}

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
	})

	t.Run("newAccount service error", func(t *testing.T) {
		a.Cpf = "999.666.999-66"
		jsonPayload, _ := json.Marshal(a)
		payload := bytes.NewBuffer(jsonPayload)

		req, err := http.NewRequest(http.MethodPost, path.String(), payload)
		if err != nil {
			t.Fatal(err)
		}

		mockAddAccount = func(a account.NewAccountRequest) error {
			return errors.New("invalid input: cpf")
		}

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
	})
}

func TestGetBalance(t *testing.T) {
	path := url.URL{
		Path: "/accounts/2/balance",
	}
	r := &mockRepository{}
	s := mockService{r}

	t.Run("getBalance is OK", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, path.String(), nil)
		if err != nil {
			t.Fatal(err)
		}

		mockGetAccountBalance = func(uint64) (account.BalanceResponse, error) {
			return account.BalanceResponse{Balance: 0}, nil
		}

		rr := httptest.NewRecorder()
		router := mux.NewRouter()
		router.HandleFunc("/accounts/{id}/balance", getBalance(&s))
		router.ServeHTTP(rr, req)

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
	})
}

func TestGetSelfBalanceIsOk(t *testing.T) {
	path := url.URL{
		Path: "/accounts/balance",
	}
	r := &mockRepository{}
	s := mockService{r}

	t.Run("getSelfBalance is OK", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, path.String(), nil)
		if err != nil {
			t.Fatal(err)
		}

		mockGetAccountBalance = func(uint64) (account.BalanceResponse, error) {
			return account.BalanceResponse{Balance: 0}, nil
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(getSelfBalance(&s))
		ctx := req.Context()
		ctx = context.WithValue(ctx, middleware.UserIdContextKey("userId"), uint64(1))
		ro := req.Clone(ctx)

		handler.ServeHTTP(rr, ro)

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
	})
}
