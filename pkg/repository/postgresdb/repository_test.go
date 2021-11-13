package postgresdb

import (
	"log"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/GilbertoVGL/go-banking/pkg/account"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func NewMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	sqlDb, mock, err := sqlmock.New()
	if err != nil {
		log.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn: sqlDb,
	}), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		t.Fatalf("new repository error: %s", err.Error())
	}

	return db, mock
}

func TestNewAccountIsOk(t *testing.T) {
	db, mock := NewMock(t)

	repository := postgresDB{db}

	defer func() {
		repository.Close()
	}()

	acc := account.NewAccountRequest{}
	acc.Balance = 10
	acc.Cpf = "980.421.840-26"
	acc.Name = "test"
	acc.Secret = "123"

	mock.ExpectBegin()

	err := repository.AddAccount(acc)
	assert.NoError(t, err)
}
