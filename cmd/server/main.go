package main

import (
	"os"
	"strconv"

	"github.com/GilbertoVGL/go-banking/pkg/config"
	"github.com/GilbertoVGL/go-banking/pkg/logger"
	"github.com/GilbertoVGL/go-banking/pkg/server"
)

func main() {
	logger.New(os.Stdout)
	logger.Log.Info("Started Banking API")

	if err := config.Load(".env"); err != nil {
		logger.Log.Fatal(err)
	}

	port, err := strconv.Atoi(os.Getenv("APP_PORT"))
	logger.Log.Debug("APP Port:", port)

	if err != nil {
		logger.Log.Fatal(err)
	}

	s, err := server.New(port)

	if err != nil {
		logger.Log.Fatal(err)
	}

	logger.Log.Fatal(s.ListenAndServe())
}
