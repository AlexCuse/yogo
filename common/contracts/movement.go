package contracts

import (
	iex "github.com/goinvest/iexcloud/v2"
)

func PreviousDayMovement(previous iex.PreviousDay) Movement {
	return Movement{
		Symbol: previous.Symbol,
		Open:   previous.Open,
		Close:  previous.Close,
		Volume: previous.Volume,
		Date:   Date(previous.Date),
	}
}

type Movement struct {
	Symbol string
	Open   float64
	Close  float64
	Volume int
	Date   Date
}
