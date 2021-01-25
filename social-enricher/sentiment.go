package social

import (
	"time"
)

type Sentiment struct {
	Sybmol    string
	Bearish   int
	Bullish   int
	Timestamp time.Time
}
