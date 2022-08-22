package models

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
	ID        string `gorm:"primaryKey"`
	Symbol0   string
	Amount0   string
	Decimals0 int
	Amount1   string
	Symbol1   string
	Decimals1 int
	Status    string `gorm:"index"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
