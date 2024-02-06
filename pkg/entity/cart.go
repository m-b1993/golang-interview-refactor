package entity

import (
	"gorm.io/gorm"
)

type Status string

const (
	CartOpen   Status = "open"
	CartClosed Status = "closed"
)

type CartEntity struct {
	gorm.Model
	Total     float64
	SessionID string
	Status    Status `gorm:"type:enum('open', 'closed')"`
}
