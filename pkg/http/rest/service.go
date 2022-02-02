package rest

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
	"github.com/GilbertoVGL/go-banking/pkg/config"
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
	r.Use(middleware.ReqTimeout)

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

	return http.TimeoutHandler(r, config.ServerReadTimeout, "Timeout")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	type isOk struct {
		Ok bool `json:"ok"`
	}

	respondWithJSON(w, http.StatusOK, isOk{true})
}

func doLogin(s login.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newLogin login.LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&newLogin); err != nil {
			respondWithError(w, http.StatusBadRequest, apperrors.NewArgumentError(err.Error()))
			return
		}

		loginCh := make(chan login.LoginReponse)
		errorCh := make(chan error)

		go func() {
			login, err := s.LoginUser(r.Context(), newLogin)
			if err != nil {
				errorCh <- err
				return
			}
			loginCh <- login
		}()

		select {
		case loginResponse := <-loginCh:
			respondWithJSON(w, http.StatusOK, loginResponse)
		case err := <-errorCh:
			respondWithError(w, http.StatusBadRequest, err)
		case <-r.Context().Done():
			respondWithError(w, http.StatusRequestTimeout, apperrors.NewInternalServerError("request timeout"))
		}
	}
}

func doTransfer(s transfer.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newTransfer transfer.TransferRequest
		newTransfer.Origin = r.Context().Value(middleware.UserIdContextKey("userId")).(uint64)

		if err := json.NewDecoder(r.Body).Decode(&newTransfer); err != nil {
			respondWithError(w, http.StatusBadRequest, apperrors.NewArgumentError(err.Error()))
			return
		}

		transferCh := make(chan bool)
		errCh := make(chan error)

		go func() {
			if err := s.DoTransfer(r.Context(), newTransfer); err != nil {
				errCh <- err
				return
			}

			transferCh <- true
		}()

		select {
		case err := <-errCh:
			switch err.(type) {
			case *apperrors.ArgumentError:
				respondWithError(w, http.StatusBadRequest, err)
			case *apperrors.TransferRequestError:
				respondWithError(w, http.StatusBadRequest, err)
			default:
				respondWithError(w, http.StatusInternalServerError, err)
			}
			return
		case <-transferCh:
			respondWithJSON(w, http.StatusCreated, newTransfer)
		case <-r.Context().Done():
			respondWithError(w, http.StatusRequestTimeout, apperrors.NewInternalServerError("request timeout"))
		}
	}
}

func getTransfer(s transfer.Service) http.HandlerFunc {
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
			respondWithError(w, http.StatusBadRequest, apperrors.NewArgumentError("invalid query params", strings.Join(invalid, ", ")))
			return
		}

		listTransferCh := make(chan transfer.ListTransferResponse)
		errCh := make(chan error)

		go func() {
			listTransfer, err := s.GetTransfers(r.Context(), id, query)

			if err != nil {
				errCh <- err
				return
			}
			listTransferCh <- listTransfer
		}()

		select {
		case listTransfer := <-listTransferCh:
			respondWithJSON(w, http.StatusOK, listTransfer)
		case err := <-errCh:
			respondWithError(w, http.StatusInternalServerError, err)
		case <-r.Context().Done():
			respondWithError(w, http.StatusRequestTimeout, apperrors.NewInternalServerError("request timeout"))
		}
	}
}

func newAccount(s account.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newAccount account.NewAccountRequest

		if err := json.NewDecoder(r.Body).Decode(&newAccount); err != nil {
			respondWithError(w, http.StatusBadRequest, apperrors.NewArgumentError(err.Error()))
			return
		}

		newAccountCh := make(chan account.NewAccountResponse)
		errCh := make(chan error)

		go func() {
			message, err := s.NewAccount(r.Context(), newAccount)
			if err != nil {
				errCh <- err
				return
			}
			newAccountCh <- message
		}()

		select {
		case message := <-newAccountCh:
			respondWithJSON(w, http.StatusCreated, message)
		case err := <-errCh:
			respondWithError(w, http.StatusBadRequest, err)
		case <-r.Context().Done():
			respondWithError(w, http.StatusRequestTimeout, apperrors.NewInternalServerError("request timeout"))
		}

	}
}

func listAccounts(s account.Service) http.HandlerFunc {
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
			respondWithError(w, http.StatusBadRequest, apperrors.NewArgumentError("invalid query params", strings.Join(invalid, ", ")))
			return
		}

		accountListCh := make(chan account.ListAccountsReponse)
		errCh := make(chan error)

		go func() {
			listAccounts, err := s.List(r.Context(), query)
			if err != nil {
				errCh <- err
				return
			}
			accountListCh <- listAccounts
		}()

		select {
		case listAccounts := <-accountListCh:
			respondWithJSON(w, http.StatusOK, listAccounts)
		case err := <-errCh:
			respondWithError(w, http.StatusBadRequest, err)
		case <-r.Context().Done():
			respondWithError(w, http.StatusRequestTimeout, apperrors.NewInternalServerError("request timeout"))
		}
	}
}

func getBalance(s account.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		userId, err := strconv.Atoi(id)

		if err != nil {
			respondWithError(w, http.StatusBadRequest, apperrors.NewArgumentError("invalid id format"))
			return
		}

		balanceCh := make(chan account.BalanceResponse)
		errCh := make(chan error)

		go func() {
			balance, err := s.GetBalance(r.Context(), uint64(userId))
			if err != nil {
				errCh <- err
				return
			}
			balanceCh <- balance
		}()

		select {
		case balance := <-balanceCh:
			respondWithJSON(w, http.StatusOK, balance)
		case err := <-errCh:
			respondWithError(w, http.StatusInternalServerError, err)
		case <-r.Context().Done():
			respondWithError(w, http.StatusRequestTimeout, apperrors.NewInternalServerError("request timeout"))
		}
	}
}

func getSelfBalance(s account.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value(middleware.UserIdContextKey("userId")).(uint64)

		balanceCh := make(chan account.BalanceResponse)
		errCh := make(chan error)

		go func() {
			balance, err := s.GetBalance(r.Context(), userId)
			if err != nil {
				errCh <- err
				return
			}
			balanceCh <- balance
		}()

		select {
		case balance := <-balanceCh:
			respondWithJSON(w, http.StatusOK, balance)
		case err := <-errCh:
			respondWithError(w, http.StatusInternalServerError, err)
		case <-r.Context().Done():
			respondWithError(w, http.StatusRequestTimeout, apperrors.NewInternalServerError("request timeout"))
		}
	}
}

func respondWithError(w http.ResponseWriter, code int, err error) {
	respondWithJSON(w, code, apperrors.RestError{Err: err.Error()})
}

func respondWithJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}
