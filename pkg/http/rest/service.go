package rest

import (
	"encoding/json"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/apperrors"
	"github.com/GilbertoVGL/go-banking/pkg/config"
	"github.com/GilbertoVGL/go-banking/pkg/http/rest/middleware"
	"github.com/GilbertoVGL/go-banking/pkg/logger"
	"github.com/GilbertoVGL/go-banking/pkg/login"
	"github.com/GilbertoVGL/go-banking/pkg/transfer"
)

func NewRouter(l login.Service, a account.Service, t transfer.Service) http.Handler {
	r := mux.NewRouter()

	// Open routes \/
	r.HandleFunc("/", healthCheck).Methods("GET").Name("Health Check")
	r.HandleFunc("/login", doLogin(l)).Methods("POST").Name("Login")
	r.HandleFunc("/accounts", newAccount(a)).Methods("POST").Name("Create account")
	r.Use(middleware.ReqTimeout)

	// Needs auth \/
	transferRouter := r.PathPrefix("/transfers").Subrouter()
	transferRouter.HandleFunc("", doTransfer(t)).Methods("POST").Name("Create transfer")
	transferRouter.HandleFunc("", listTransfer(t)).Methods("GET").Name("Read transfer")
	transferRouter.Use(middleware.Auth)

	accountRouter := r.PathPrefix("/accounts").Subrouter()
	accountRouter.HandleFunc("", listAccounts(a)).Methods("GET").Name("List accounts")
	accountRouter.HandleFunc("/balance", getSelfBalance(a)).Methods("GET").Name("Get current user balance")
	accountRouter.HandleFunc("/{id}/balance", getBalance(a)).Methods("GET").Name("Get some user balance")
	accountRouter.Use(middleware.Auth)

	headersOk := handlers.AllowedHeaders([]string{"Origin", "Content-Type", "Authorization"})
	originsOk := handlers.AllowedOrigins([]string{os.Getenv("ORIGIN_ALLOWED")})
	methodsOk := handlers.AllowedMethods([]string{"GET", "POST", "OPTIONS"})

	walkRoutes(r)

	return http.TimeoutHandler(handlers.CORS(originsOk, headersOk, methodsOk)(r), config.ServerReadTimeout, "Timeout")
}

func walkRoutes(r *mux.Router) {
	logger.Log.Info("********************************")
	logger.Log.Info("API Routes:")
	r.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		name := route.GetName()
		tpl, _ := route.GetPathTemplate()
		met, _ := route.GetMethods()
		if len(met) > 0 {
			logger.Log.Info(name, ":", met, "-", tpl)
		}
		return nil
	})
	logger.Log.Info("*******************************")
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	type isOk struct {
		Ok bool `json:"ok"`
	}

	ok := isOk{true}
	logger.Log.Info("Healthy:", ok.Ok)

	respondWithJSON(w, http.StatusOK, ok)
}

func doLogin(s login.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newLogin login.LoginRequest

		if err := json.NewDecoder(r.Body).Decode(&newLogin); err != nil {
			logger.Log.Error("Error while decoding do login body", err)
			respondWithError(w, http.StatusBadRequest, apperrors.NewArgumentError(err.Error()))
			return
		}

		logger.Log.Debug("Trying to login:", newLogin.Cpf)

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
			logger.Log.Debug("User succesfully logged in:", newLogin.Cpf)
			respondWithJSON(w, http.StatusOK, loginResponse)
		case err := <-errorCh:
			logger.Log.Error(err)
			respondWithError(w, http.StatusBadRequest, err)
		case <-r.Context().Done():
			err := apperrors.NewInternalServerError("request timeout")
			logger.Log.Error("Do login", err)
			respondWithError(w, http.StatusRequestTimeout, err)
		}
	}
}

func doTransfer(s transfer.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newTransfer transfer.TransferRequest
		newTransfer.Origin = r.Context().Value(middleware.UserIdContextKey("userId")).(uint64)

		if err := json.NewDecoder(r.Body).Decode(&newTransfer); err != nil {
			logger.Log.Error("Error while decoding do transfer body", err)
			respondWithError(w, http.StatusBadRequest, apperrors.NewArgumentError(err.Error()))
			return
		}

		logger.Log.Debug("Trying to do transfer from", newTransfer.Origin, "to", newTransfer.Destination, "of value", newTransfer.Amount)

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
		case <-transferCh:
			logger.Log.Debug("Transfer successfully made from account", newTransfer.Origin, "to", *newTransfer.Destination, "of value", *newTransfer.Amount)
			respondWithJSON(w, http.StatusCreated, newTransfer)
		case err := <-errCh:
			logger.Log.Error("Do Transfer error", err)
			switch err.(type) {
			case *apperrors.ArgumentError, *apperrors.TransferRequestError:
				respondWithError(w, http.StatusBadRequest, err)
			default:
				respondWithError(w, http.StatusInternalServerError, err)
			}
			return
		case <-r.Context().Done():
			err := apperrors.NewInternalServerError("request timeout")
			logger.Log.Error("Do transfer", err)
			respondWithError(w, http.StatusRequestTimeout, err)
		}
	}
}

func listTransfer(s transfer.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var invalid []string
		id := r.Context().Value(middleware.UserIdContextKey("userId")).(uint64)
		query := transfer.ListTransferQuery{
			PageSize: 15,
			Page:     0,
		}

		logger.Log.Debug("List transfers from user", id)

		if v := r.FormValue("pageSize"); v != "" {
			page, err := strconv.Atoi(v)

			if err != nil {
				logger.Log.Debug("List transfers invalid pageSize", v)
				invalid = append(invalid, "pageSize")
			} else {
				query.PageSize = page
			}
		}

		if v := r.FormValue("page"); v != "" {
			page, err := strconv.Atoi(v)

			if err != nil || page < 1 {
				logger.Log.Debug("List transfers invalid page", v)
				invalid = append(invalid, "page")
			} else {
				query.Page = page - 1
			}
		}

		if len(invalid) > 0 {
			err := apperrors.NewArgumentError("invalid query params", strings.Join(invalid, ", "))
			logger.Log.Error("List transfers invalid params", err)
			respondWithError(w, http.StatusBadRequest, err)
			return
		}

		transferListCh := make(chan transfer.ListTransferResponse)
		errCh := make(chan error)

		go func() {
			transferList, err := s.GetTransfers(r.Context(), id, query)

			if err != nil {
				errCh <- err
				return
			}
			transferListCh <- transferList
		}()

		select {
		case transferList := <-transferListCh:
			logger.Log.Debug("Successfully listed transfers", transferList)
			respondWithJSON(w, http.StatusOK, transferList)
		case err := <-errCh:
			logger.Log.Error("List transfer error", err)
			respondWithError(w, http.StatusInternalServerError, err)
		case <-r.Context().Done():
			err := apperrors.NewInternalServerError("request timeout")
			logger.Log.Error("List transfer", err)
			respondWithError(w, http.StatusRequestTimeout, err)
		}
	}
}

func newAccount(s account.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var newAccount account.NewAccountRequest

		if err := json.NewDecoder(r.Body).Decode(&newAccount); err != nil {
			logger.Log.Error("Error while decoding new account body", err)
			respondWithError(w, http.StatusBadRequest, apperrors.NewArgumentError(err.Error()))
			return
		}

		logger.Log.Debug("Trying to create new account for", newAccount.Cpf)

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
			logger.Log.Debug("New account created with balance", newAccount.Balance, "CPF", newAccount.Cpf, "and name", newAccount.Name)
			respondWithJSON(w, http.StatusCreated, message)
		case err := <-errCh:
			logger.Log.Error("New account error", err)
			respondWithError(w, http.StatusBadRequest, err)
		case <-r.Context().Done():
			err := apperrors.NewInternalServerError("request timeout")
			logger.Log.Error("New account", err)
			respondWithError(w, http.StatusRequestTimeout, err)
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

		if v := r.FormValue("pageSize"); v != "" {
			page, err := strconv.Atoi(v)

			if err != nil {
				logger.Log.Debug("List accounts invalid pageSize", v)
				invalid = append(invalid, "pageSize")
			} else {
				query.PageSize = page
			}
		}

		if v := r.FormValue("page"); v != "" {
			page, err := strconv.Atoi(v)

			if err != nil || page < 1 {
				logger.Log.Debug("List accounts invalid page", v)
				invalid = append(invalid, "page")
			} else {
				query.Page = page - 1
			}
		}

		if len(invalid) > 0 {
			err := apperrors.NewArgumentError("invalid query params", strings.Join(invalid, ", "))
			logger.Log.Error("List transfers invalid params", err)
			respondWithError(w, http.StatusBadRequest, err)
			return
		}

		accountsListCh := make(chan account.ListAccountsReponse)
		errCh := make(chan error)

		go func() {
			accountsList, err := s.List(r.Context(), query)
			if err != nil {
				errCh <- err
				return
			}
			accountsListCh <- accountsList
		}()

		select {
		case accountsList := <-accountsListCh:
			logger.Log.Debug("Successfully listed accounts", accountsList)
			respondWithJSON(w, http.StatusOK, accountsList)
		case err := <-errCh:
			logger.Log.Error("List accounts error", err)
			respondWithError(w, http.StatusBadRequest, err)
		case <-r.Context().Done():
			err := apperrors.NewInternalServerError("request timeout")
			logger.Log.Error("List transfer", err)
			respondWithError(w, http.StatusRequestTimeout, err)
		}
	}
}

func getBalance(s account.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		userId, err := strconv.Atoi(id)

		if err != nil {
			err := apperrors.NewArgumentError("invalid id format")
			logger.Log.Error("Error while decoding get balance user id", err)
			respondWithError(w, http.StatusBadRequest, err)
			return
		}

		logger.Log.Debug("Trying to get balance from", userId)

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
			logger.Log.Debug("Got balance from", userId, "balance", balance)
			respondWithJSON(w, http.StatusOK, balance)
		case err := <-errCh:
			logger.Log.Error("Get balance error", err)
			respondWithError(w, http.StatusInternalServerError, err)
		case <-r.Context().Done():
			err := apperrors.NewInternalServerError("request timeout")
			logger.Log.Error("Get balance", err)
			respondWithError(w, http.StatusRequestTimeout, err)
		}
	}
}

func getSelfBalance(s account.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userId := r.Context().Value(middleware.UserIdContextKey("userId")).(uint64)

		logger.Log.Debug("Trying to get self balance from", userId)

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
			logger.Log.Debug("Got self balance from", userId, "balance", balance)
			respondWithJSON(w, http.StatusOK, balance)
		case err := <-errCh:
			logger.Log.Error("Get self balance error", err)
			respondWithError(w, http.StatusInternalServerError, err)
		case <-r.Context().Done():
			err := apperrors.NewInternalServerError("request timeout")
			logger.Log.Error("Get self balance", err)
			respondWithError(w, http.StatusRequestTimeout, err)
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
