package rest

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
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
	accountRouter.HandleFunc("/balance", getSelfBalance(a)).Methods("GET")
	accountRouter.HandleFunc("/{id}/balance", getBalance(a)).Methods("GET")
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
			respondWithError(w, http.StatusBadRequest, err)
			return
		}

		loginResponse, err := s.LoginUser(newLogin)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, err)
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
			respondWithError(w, http.StatusBadRequest, err)
			return
		}

		if err := s.DoTransfer(newTransfer); err != nil {
			switch err.(type) {
			case *apperrors.ArgumentError:
				respondWithError(w, http.StatusBadRequest, err)
			case *apperrors.TransferRequestError:
				respondWithError(w, http.StatusBadRequest, err)
			default:
				respondWithError(w, http.StatusInternalServerError, err)
			}
			return
		}

		respondWithJSON(w, http.StatusCreated, newTransfer)
	}
}

func getTransfer(s transfer.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var invalid []string
		id := r.Context().Value(middleware.UserIdContextKey("userId")).(uint64)
		query := transfer.ListTransferQuery{
			PageSize: 15,
			Page:     0,
		}

		if r.FormValue("pageSize") != "" {
			page, err := strconv.Atoi(r.FormValue("pageSize"))
			if err != nil {
				invalid = append(invalid, "pageSize")
			} else {
				query.PageSize = page
			}
		}

		if r.FormValue("page") != "" {
			page, err := strconv.Atoi(r.FormValue("page"))
			if err != nil || page < 1 {
				invalid = append(invalid, "page")
			} else {
				query.Page = page - 1
			}
		}

		if len(invalid) > 0 {
			respondWithError(w, http.StatusBadRequest, errors.New(fmt.Sprintf("invalid query params: %s", strings.Join(invalid, ", "))))
			return
		}

		listTransfer, err := s.GetTransfers(id, query)

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		respondWithJSON(w, http.StatusOK, listTransfer)
	}
}

func newAccount(s account.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var newAccount account.NewAccountRequest

		if err := json.NewDecoder(r.Body).Decode(&newAccount); err != nil {
			respondWithError(w, http.StatusBadRequest, err)
			return
		}

		if err := s.NewAccount(newAccount); err != nil {
			respondWithError(w, http.StatusBadRequest, err)
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
			Page:     0,
		}

		if r.FormValue("pageSize") != "" {
			page, err := strconv.Atoi(r.FormValue("pageSize"))
			if err != nil {
				invalid = append(invalid, "pageSize")
			} else {
				query.PageSize = page
			}
		}

		if r.FormValue("page") != "" {
			page, err := strconv.Atoi(r.FormValue("page"))
			if err != nil || page < 1 {
				invalid = append(invalid, "page")
			} else {
				query.Page = page - 1
			}
		}

		if len(invalid) > 0 {
			respondWithError(w, http.StatusBadRequest, errors.New(fmt.Sprintf("invalid query params: %s", strings.Join(invalid, ", "))))
			return
		}

		listAccounts, err := s.List(query)

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		respondWithJSON(w, http.StatusOK, listAccounts)
	}
}

func getBalance(s account.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		userId, err := strconv.Atoi(id)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, errors.New("invalid id format"))
			return
		}

		balance, err := s.GetBalance(uint64(userId))

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		respondWithJSON(w, http.StatusOK, balance)
	}
}

func getSelfBalance(s account.Service) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value(middleware.UserIdContextKey("userId")).(uint64)
		balance, err := s.GetBalance(userId)

		if err != nil {
			respondWithError(w, http.StatusInternalServerError, err)
			return
		}

		respondWithJSON(w, http.StatusOK, balance)
	}
}

func respondWithError(w http.ResponseWriter, code int, err error) {
	respondWithJSON(w, code, map[string]string{"error": err.Error()})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
