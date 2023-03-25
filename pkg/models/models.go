package models

import (
	"time"

	"gorm.io/gorm"
)

type Order struct {
	ID            uint64 `gorm:"primaryKey"`
	PositionID    string `gorm:"index"`
	Symbol        string
	ClientOrderID string
	Price         string
	Amount        string
	Side          string
	UpdateTime    int64
}

type Position struct {
	ID            string `gorm:"primaryKey"`
	Liquidity     string
	TickLower     int
	TickUpper     int
	Symbol0       string
	Amount0       string
	Decimals0     int
	HedgedAmount0 string
	Amount1       string
	Symbol1       string
	Decimals1     int
	HedgedAmount1 string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func AutoMigrate(db *gorm.DB, reset bool) error {
	if reset {
		err := db.Migrator().DropTable(&Position{})
		if err != nil {
			return err
		}
	}

	return db.AutoMigrate(&Position{})
}
