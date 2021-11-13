package main

import (
	"log"
	"os"
	"strconv"

	"github.com/GilbertoVGL/go-banking/pkg/config"
	"github.com/GilbertoVGL/go-banking/pkg/server"
)

func main() {
	log.Println("Started Banking API")

	if err := config.Load(".env"); err != nil {
		log.Fatal(err)
	}

	port, err := strconv.Atoi(os.Getenv("APP_PORT"))

	if err != nil {
		log.Fatalln(err)
	}

	s, err := server.New(port)

	if err != nil {
		log.Fatalln(err)
	}

	log.Fatal(s.ListenAndServe())
}
