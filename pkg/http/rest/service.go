package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/http/rest/middleware"
	"github.com/GilbertoVGL/go-banking/pkg/login"
	"github.com/GilbertoVGL/go-banking/pkg/transfer"
)

func NewRouter(l login.Service, a account.Service, t transfer.Service) http.Handler {
	r := mux.NewRouter()

	// Open routes \/
	r.HandleFunc("/", healthCheck).Methods("GET")
	r.HandleFunc("/login", doLogin(l)).Methods("POST")
	r.HandleFunc("/accounts", newAccount(a)).Methods("POST")

	// Needs auth \/
	transferRouter := r.PathPrefix("/transfers").Subrouter()
	transferRouter.HandleFunc("", doTransfer(t)).Methods("POST")
	transferRouter.HandleFunc("", getTransfer(t)).Methods("GET")
	transferRouter.Use(middleware.Auth)

	accountRouter := r.PathPrefix("/accounts").Subrouter()
	accountRouter.HandleFunc("", listAccounts(a)).Methods("GET")
	accountRouter.HandleFunc("/balance", getBalance(a)).Methods("GET")
	accountRouter.Use(middleware.Auth)

	return r
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	type isOk struct {
		Ok bool `json:"ok"`
	}

	respondWithJSON(w, http.StatusOK, isOk{true})
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
		newTransfer.Origin = r.Context().Value(middleware.UserIdContextKey("userId")).(uint64)

		if err := json.NewDecoder(r.Body).Decode(&newTransfer); err != nil {
			respondWithError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := s.DoTransfer(newTransfer); err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

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
		query := account.ListAccountQuery{
			PageSize: 15,
			Offset:   0,
		}

		if r.FormValue("pageSize") != "" {
			page, err := strconv.Atoi(r.FormValue("pageSize"))
			if err != nil {
				invalid = append(invalid, "pageSize")
			} else {
				query.PageSize = page
			}
		}

		if r.FormValue("offset") != "" {
			offset, err := strconv.Atoi(r.FormValue("offset"))
			if err != nil {
				invalid = append(invalid, "offset")
			} else {
				query.Offset = offset - 1
			}
		}

		if len(invalid) > 0 {
			respondWithError(w, http.StatusBadRequest, fmt.Sprintf("invalid query params: %s", strings.Join(invalid, ", ")))
			return
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
		userId := r.Context().Value(middleware.UserIdContextKey("userId")).(uint64)
		balance, err := s.GetBalance(userId)

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err.Error())
			return
		}

		respondWithJSON(w, http.StatusOK, balance)
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
