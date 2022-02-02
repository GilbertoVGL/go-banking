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

var (
	mockListAccount       func(context.Context, account.ListAccountQuery) (account.ListAccountsReponse, error)
	mockAddAccount        func(context.Context, account.NewAccountRequest) error
	mockLogin             func(context.Context, login.LoginRequest) (login.Account, error)
	mockGetAccountBalance func(context.Context, uint64) (account.BalanceResponse, error)
	mockGetTransfer       func(context.Context, uint64, transfer.ListTransferQuery) (transfer.ListTransferResponse, error)
)

func (mr *mockRepository) ListAccount(ctx context.Context, params account.ListAccountQuery) (account.ListAccountsReponse, error) {
	return mockListAccount(ctx, params)
}
func (mr *mockRepository) AddAccount(ctx context.Context, a account.NewAccountRequest) error {
	return mockAddAccount(ctx, a)
}
func (mr *mockRepository) GetAccountBySecretAndCPF(ctx context.Context, l login.LoginRequest) (login.Account, error) {
	return mockLogin(ctx, l)
}
func (mr *mockRepository) GetAccountBalance(ctx context.Context, a uint64) (account.BalanceResponse, error) {
	return mockGetAccountBalance(ctx, a)
}
func (mr *mockRepository) GetTransfers(ctx context.Context, a uint64, l transfer.ListTransferQuery) (transfer.ListTransferResponse, error) {
	return mockGetTransfer(ctx, a, l)
}

type mockService struct {
	r *mockRepository
}

func (ms *mockService) List(ctx context.Context, a account.ListAccountQuery) (account.ListAccountsReponse, error) {
	return ms.r.ListAccount(ctx, a)
}
func (ms *mockService) NewAccount(ctx context.Context, a account.NewAccountRequest) (account.NewAccountResponse, error) {
	var r account.NewAccountResponse
	err := ms.r.AddAccount(ctx, a)
	if err != nil {
		return r, err
	}
	r.Msg = "account created"
	return r, nil
}
func (ms *mockService) GetBalance(ctx context.Context, a uint64) (account.BalanceResponse, error) {
	return ms.r.GetAccountBalance(ctx, a)
}
func (ms *mockService) LoginUser(ctx context.Context, l login.LoginRequest) (login.LoginReponse, error) {
	account, err := ms.r.GetAccountBySecretAndCPF(ctx, l)
	return login.LoginReponse{Token: account.Cpf}, err
}
func (ms *mockService) GetTransfers(ctx context.Context, a uint64, l transfer.ListTransferQuery) (transfer.ListTransferResponse, error) {
	return ms.r.GetTransfers(ctx, a, l)
}
func (ms *mockService) DoTransfer(ctx context.Context, t transfer.TransferRequest) error {
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

		expected := login.Account{}
		var result login.Account
		json.NewDecoder(rr.Body).Decode(&result)

		if result != expected {
			t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
				result, expected)
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

func TestGetTransfer(t *testing.T) {
	path := url.URL{
		Path:     "/transfers",
		RawQuery: (&url.Values{"pageSize": []string{"10"}, "page": []string{"1"}}).Encode(),
	}
	r := &mockRepository{}
	s := mockService{r}

	t.Run("getTransfer is OK", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, path.String(), nil)
		if err != nil {
			t.Fatal(err)
		}

		mockGetTransfer = func(ctx context.Context, a uint64, l transfer.ListTransferQuery) (transfer.ListTransferResponse, error) {
			transfers := []transfer.ListTransfer{}
			return transfer.ListTransferResponse{
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
		path.RawQuery = (&url.Values{"pageSize": []string{"10"}, "page": []string{"1"}}).Encode()
		req, err := http.NewRequest(http.MethodGet, path.String(), nil)
		if err != nil {
			t.Fatal(err)
		}

		mockListAccount = func(ctx context.Context, q account.ListAccountQuery) (account.ListAccountsReponse, error) {
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
		result := strings.Trim(rr.Body.String(), " \r\n")

		if result != expected {
			t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
				result, expected)
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

		expected := strings.Trim(`{"error":"invalid argument: invalid query params: pageSize, page"}`, " \r\n")
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

		mockListAccount = func(ctx context.Context, q account.ListAccountQuery) (account.ListAccountsReponse, error) {
			return account.ListAccountsReponse{}, errors.New("unable to access database")
		}

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(listAccounts(&s))

		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v",
				status, http.StatusBadRequest)
		}

		expected := strings.Trim(`{"error":"unable to access database"}`, "\r\n")
		body := strings.Trim(rr.Body.String(), "\r\n")

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

		mockAddAccount = func(ctx context.Context, a account.NewAccountRequest) error {
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

		mockAddAccount = func(ctx context.Context, a account.NewAccountRequest) error {
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

		mockGetAccountBalance = func(ctx context.Context, i uint64) (account.BalanceResponse, error) {
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

		expected := account.BalanceResponse{}
		var result account.BalanceResponse
		json.NewDecoder(rr.Body).Decode(&result)

		if result != expected {
			t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
				result, expected)
		}
	})
}

func TestGetSelfBalance(t *testing.T) {
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

		mockGetAccountBalance = func(ctx context.Context, i uint64) (account.BalanceResponse, error) {
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

		expected := account.BalanceResponse{}
		var result account.BalanceResponse
		json.NewDecoder(rr.Body).Decode(&result)

		if result != expected {
			t.Errorf("handler returned unexpected body: \ngot \n\t%v\n want \n\t%v",
				result, expected)
		}
	})
}
