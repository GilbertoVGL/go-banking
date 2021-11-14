package postgresdb

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/login"
)

type postgresDB struct {
	db *pgxpool.Pool
}

func new() (*pgxpool.Pool, error) {
	dsn := url.URL{
		User:     url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PW")),
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		Path:     os.Getenv("DB_NAME"),
		RawQuery: (&url.Values{"sslmode": []string{"disable"}}).Encode(),
	}

	config, err := pgxpool.ParseConfig(dsn.String())

	if err != nil {
		return nil, err
	}

	pool, err := strconv.Atoi(os.Getenv("DB_MAX_POOL"))

	if err != nil {
		return nil, err
	}

	config.MaxConns = int32(pool)

	db, err := pgxpool.ConnectConfig(context.Background(), config)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func (r *postgresDB) getConn() (*pgxpool.Conn, error) {
	if r == nil || r.db == nil {
		db, err := new()
		if err != nil {
			log.Print("warning: unable to initialize db:", err)
			return nil, errors.New("unable to initialize db")
		}
		r = &postgresDB{db}
	}

	conn, err := r.db.Acquire(context.Background())

	if err != nil {
		return conn, err
	}

	return conn, nil
}

func NewRepository() (*postgresDB, error) {
	db, err := new()
	if err != nil {
		return &postgresDB{}, err
	}

	return &postgresDB{db}, nil
}

func (r *postgresDB) Close() {
	r.db.Close()
}

func (r *postgresDB) Login(l login.LoginRequest) (bool, error) {
	conn, err := r.getConn()
	defer conn.Release()

	if err != nil {
		return false, err
	}

	var account Account
	query := fmt.Sprintf("SELECT * FROM accounts WHERE cpf = %s AND secret = %s;", l.Cpf, l.Secret)

	if err := conn.QueryRow(context.Background(), query).Scan(&account); err != nil {
		return false, err
	}

	return false, nil
}

func (r *postgresDB) ListAccounts() ([]account.ListAccountsReponse, error) {
	var accountsResponse []account.ListAccountsReponse
	var accounts []Account
	var count int

	conn, err := r.getConn()

	if err != nil {
		return accountsResponse, err
	}

	query := fmt.Sprintf("select name, cpf, balance from accounts limit 5;")
	if err := conn.QueryRow(context.Background(), query).Scan(&accounts); err != nil {
		return nil, err
	}

	countQuery := fmt.Sprintf("select count(*) from accounts;")

	if err := conn.QueryRow(context.Background(), countQuery).Scan(&count); err != nil {
		return nil, err
	}

	fmt.Printf("%+v\n", accounts)

	return accountsResponse, nil
}

func (r *postgresDB) AddAccount(a account.NewAccountRequest) error {
	conn, err := r.getConn()

	fmt.Println(conn)

	if err != nil {
		return err
	}

	var account Account

	account.Name = a.Name
	account.Cpf = a.Cpf
	account.Balance = a.Balance
	account.Secret = a.Secret
	account.Active = true

	return nil
}
