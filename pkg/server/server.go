package server

import (
	"fmt"
	"log"
	"net/http"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/http/rest"
	"github.com/GilbertoVGL/go-banking/pkg/login"
	"github.com/GilbertoVGL/go-banking/pkg/repository/postgresdb"
	"github.com/GilbertoVGL/go-banking/pkg/transfer"
)

func New(port int) (*http.Server, error) {
	db, err := postgresdb.NewRepository()
	if err != nil {
		log.Print("warning: unable to connect to database at startup: ", err)
	}

	l := login.New(db)
	a := account.New(db)
	t := transfer.New(db)

	r := rest.NewRouter(l, a, t)

	addr := fmt.Sprintf("localhost:%d", port)

	return &http.Server{
		Addr:    addr,
		Handler: r,
	}, nil
}
