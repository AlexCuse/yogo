package db

import (
	"gorm.io/datatypes"
	"time"
)

type Asset struct {
	Symbol string `gorm:"primaryKey;autoIncrement:false" json:"symbol,omitempty"`
}

type Movement struct {
	Symbol string    `gorm:"primaryKey;autoIncrement:false"`
	Date   time.Time `gorm:"primaryKey;autoIncrement:false;type:date"`
	Data   datatypes.JSON
}

type Stats struct {
	Symbol    string    `gorm:"primaryKey;autoIncrement:false"`
	QuoteDate time.Time `gorm:"primaryKey;autoIncrement:false;type:date"`
	Data      datatypes.JSON
}

type Hit struct {
	RuleName  string    `gorm:"primaryKey;autoIncrement:false"`
	Symbol    string    `gorm:"primaryKey;autoIncrement:false"`
	QuoteDate time.Time `gorm:"primaryKey;autoIncrement:false;type:date"`
}
