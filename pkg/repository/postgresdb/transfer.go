package postgresdb

import (
	"gorm.io/gorm"
)

type Transfer struct {
	gorm.Model
	account_origin_id      uint    `gorm:"not null"`
	account_destination_id uint    `gorm:"not null"`
	amount                 float64 `gorm:"not null"`
}
