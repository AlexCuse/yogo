package signals

import (
	"encoding/json"

	"github.com/alexcuse/yogo/internal/pkg/db"
	iex "github.com/goinvest/iexcloud/v2"
)

type Signal struct {
	Name   string `gorm:"primaryKey;autoIncrement:false" json:"name,omitempty"`
	Source string `json:"source,omitempty"`
}

type SignalWithHitCount struct {
	Signal
	Count int `json:"count,omitempty"`
}

type SignalResult struct {
	Signal
	Active []SignalHit `json:"active"`
}

type SignalHit struct {
	Symbol      string   `json:"symbol"`
	QuoteDate   iex.Date `json:"quoteDate"`
	Open        float64  `json:"open"`
	Close       float64  `json:"close"`
	Volume      float64  `json:"volume"`
	CompanyName string   `json:"companyName"`
	MarketCap   float64  `json:"marketCap"`
	High52Wk    float64  `json:"high52Wk"`
	Low52Wk     float64  `json:"low52Wk"`
	Avg200Price float64  `json:"avg200rice"`
	Avg50Price  float64  `json:"avg50Price"`
	Avg10Vol    float64  `json:"avg10Vol"`
	Avg30Vol    float64  `json:"avg30Vol"`
	MA50        float64  `json:"ma50"`
	MA200       float64  `json:"ma200"`
	PE          float64  `json:"pe"`
	Beta        float64  `json:"beta"`
}

func (signal *SignalResult) AddHit(movement db.Movement, stats db.Stats) error {
	previousDay := iex.PreviousDay{}
	keyStats := iex.KeyStats{}

	err := json.Unmarshal(movement.Data, &previousDay)

	if err != nil {
		return err
	}

	err = json.Unmarshal(stats.Data, &keyStats)

	if err != nil {
		return err
	}

	hit := SignalHit{
		Symbol:      previousDay.Symbol,
		QuoteDate:   previousDay.Date,
		Open:        previousDay.Open,
		Close:       previousDay.Close,
		Volume:      previousDay.Volume,
		CompanyName: keyStats.Name,
		MarketCap:   keyStats.MarketCap,
		High52Wk:    keyStats.Week52High,
		Low52Wk:     keyStats.Week52Low,
		MA50:        keyStats.Day200MovingAvg,
		MA200:       keyStats.Day50MovingAvg,
		Avg10Vol:    keyStats.Avg10Volume,
		Avg30Vol:    keyStats.Avg30Volume,
		PE:          keyStats.PERatio,
		Beta:        keyStats.Beta,
	}

	signal.Active = append(signal.Active, hit)

	return nil
}
