package postgresdb

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/GilbertoVGL/go-banking/pkg/login"
)

type postgresDB struct {
	db *gorm.DB
}

func new() (*gorm.DB, error) {
	dsn := url.URL{
		User:     url.UserPassword(os.Getenv("DB_USER"), os.Getenv("DB_PW")),
		Scheme:   "postgres",
		Host:     fmt.Sprintf("%s:%s", os.Getenv("DB_HOST"), os.Getenv("DB_PORT")),
		Path:     os.Getenv("DB_NAME"),
		RawQuery: (&url.Values{"sslmode": []string{"disable"}}).Encode(),
	}

	db, err := gorm.Open(postgres.Open(dsn.String()), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()

	if err != nil {
		return nil, err
	}

	idle, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONN"))

	if err != nil {
		return nil, err
	}

	pool, err := strconv.Atoi(os.Getenv("DB_MAX_POOL"))

	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(idle)
	sqlDB.SetMaxOpenConns(pool)

	return db, nil
}

func (r *postgresDB) getDb() (*gorm.DB, error) {
	if r == nil || r.db == nil {
		db, err := new()
		if err != nil {
			log.Print("warning: unable to initialize db:", err)
			return nil, errors.New("unable to initialize db")
		}
		r = &postgresDB{db}
	}

	return r.db, nil
}

func NewRepository() (*postgresDB, error) {
	db, err := new()
	if err != nil {
		return &postgresDB{}, err
	}
	db.AutoMigrate(&Account{}, &Transfer{})
	return &postgresDB{db}, nil
}

func (r *postgresDB) Close() error {
	db, err := r.db.DB()
	if err != nil {
		return err
	}
	db.Close()
	return nil
}

func (r *postgresDB) Login(l login.LoginRequest) (bool, error) {
	db, err := r.getDb()

	if err != nil {
		return false, err
	}

	var account Account

	if err := db.Model(&account).Where("cpf = ?", l.Cpf).Where("secret = ?", l.Secret).First(&account).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, errors.New("wrong credentials")
		}

		return false, err
	}

	return false, nil
}

func (r *postgresDB) ListAccounts() ([]account.ListAccountsReponse, error) {
	var accountsResponse []account.ListAccountsReponse
	var accounts []Account
	var count int64

	db, err := r.getDb()

	if err != nil {
		return accountsResponse, err
	}

	if err := db.Model(&Account{}).Select("name", "cpf", "balance").Find(&accounts).Limit(5); err != nil {
		return accountsResponse, err.Error
	}

	if err := db.Model(&Account{}).Count(&count); err != nil {
		return accountsResponse, err.Error
	}

	fmt.Printf("%+v\n", accounts)

	return accountsResponse, nil
}

func (r *postgresDB) AddAccount(a account.NewAccountRequest) error {
	db, err := r.getDb()

	if err != nil {
		return err
	}

	var account Account

	account.Name = a.Name
	account.Cpf = a.Cpf
	account.Balance = a.Balance
	account.Secret = a.Secret
	account.Active = true

	stmt := db.Session(&gorm.Session{DryRun: true}).Create(&account).Statement
	fmt.Println("QUERY: ", stmt.SQL.String())
	fmt.Println("VARS: ", stmt.Vars)

	if err := db.Create(&account).Error; err != nil {
		if strings.Contains(err.Error(), "23505") {
			return errors.New("user already registered")
		}
		return err
	}

	return nil
}
