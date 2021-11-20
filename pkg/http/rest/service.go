package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/login"
	"github.com/GilbertoVGL/go-banking/pkg/middleware"
	"github.com/GilbertoVGL/go-banking/pkg/transfer"
)

func NewRouter(l login.Service, a account.Service, t transfer.Service) http.Handler {
	r := mux.NewRouter()

	transferRouter := r.PathPrefix("/transfers").Subrouter()
	transferRouter.HandleFunc("", doTransfer(t)).Methods("POST")
	transferRouter.HandleFunc("", getTransfer(t)).Methods("GET")
	transferRouter.Use(middleware.Auth)

	accountRouter := r.PathPrefix("/accounts").Subrouter()
	accountRouter.HandleFunc("", newAccount(a)).Methods("POST")
	accountRouter.HandleFunc("", listAccounts(a)).Methods("GET")
	accountRouter.HandleFunc("/{id}/balance", getBalance(a)).Methods("GET")
	accountRouter.Use(middleware.Auth)

	r.HandleFunc("/login", doLogin(l)).Methods("POST")

	return r
}

func doLogin(s login.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var newLogin login.LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&newLogin); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		loginResponse, err := s.LoginUser(newLogin)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, loginResponse)
	}
}

func doTransfer(s transfer.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var newTransfer transfer.TransferRequest

		if err := json.NewDecoder(r.Body).Decode(&newTransfer); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		s.DoTransfer(newTransfer)

		respondWithJSON(w, http.StatusCreated, newTransfer)
	}
}

func getTransfer(s transfer.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var newTransfer transfer.TransferRequest

		s.GetTransfers(newTransfer)

		respondWithJSON(w, http.StatusOK, newTransfer)
	}
}

func newAccount(s account.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var newAccount account.NewAccountRequest

		if err := json.NewDecoder(r.Body).Decode(&newAccount); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := s.NewAccount(newAccount); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		respondWithJSON(w, http.StatusCreated, map[string]string{"msg": "account created"})
	}
}

func listAccounts(s account.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var invalid []string
		limit, err := strconv.Atoi(r.FormValue("limit"))
		if err != nil {
			fmt.Println(err)
			invalid = append(invalid, "limit")
		}

		page, err := strconv.Atoi(r.FormValue("pageSize"))
		if err != nil {
			invalid = append(invalid, "pageSize")
		}

		offset, err := strconv.Atoi(r.FormValue("offset"))
		if err != nil {
			invalid = append(invalid, "offset")
		}

		if len(invalid) > 0 {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid query params: %s", strings.Join(invalid, ", ")))
			return
		}

		query := account.ListAccountQuery{
			Limit:    limit,
			PageSize: page,
			Offset:   offset,
		}

		listAccounts, err := s.List(query)

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, listAccounts)
	}
}

func getBalance(s account.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		respondWithJSON(w, http.StatusOK, account.BalanceResponse{})
	}
}

func respondWithError(w http.ResponseWriter, code int, message string) {
	respondWithJSON(w, code, map[string]string{"error": message})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
