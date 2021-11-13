package postgresdb

import (
	"gorm.io/gorm"
)

type Account struct {
	gorm.Model
	Name    string  `gorm:"not null"`
	Cpf     string  `gorm:"not null; unique"`
	Secret  string  `gorm:"not null"`
	Balance float64 `gorm:"not null"`
	Active  bool    `gorm:"not null"`
}
