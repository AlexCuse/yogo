package social

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type SentimentSnapshot struct {
	Sybmol    string
	Bearish   int
	Bullish   int
	Timestamp time.Time
	Src       string
}

func NewDailySentimentTracker(db *gorm.DB) (DailySentimentTracker, error) {
	if err := db.AutoMigrate(&DailySentiment{}); err != nil {
		return DailySentimentTracker{}, err
	}

	return DailySentimentTracker{
		db: db,
	}, nil
}

type DailySentimentTracker struct {
	db *gorm.DB
}

func (s *DailySentimentTracker) Add(ctx context.Context, sentiment *SentimentSnapshot) error {
	r := s.db.WithContext(ctx).Exec("update daily_sentiments set bearish = bearish + ?, bullish = bullish + ? where symbol = ? and date = ?",
		sentiment.Bearish, sentiment.Bullish, sentiment.Sybmol, sentiment.Timestamp)

	if r.RowsAffected == 0 {
		r = s.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(DailySentiment{
			Symbol:  sentiment.Sybmol,
			Date:    sentiment.Timestamp,
			Bearish: sentiment.Bearish,
			Bullish: sentiment.Bullish,
		})
	}

	return r.Error
}

func (t *DailySentimentTracker) Stream(ctx context.Context, sentiments chan *SentimentSnapshot) {
	for {
		select {
		case <-ctx.Done():
			return
		case s, ok := <-sentiments:
			if !ok {
				return
			}
			log := zerolog.Ctx(ctx).With().Str("symbol", s.Sybmol).Logger()
			ctx = log.WithContext(ctx)

			err := t.Add(ctx, s)
			if err != nil {
				log.Err(err).Msg("")
			}
		}
	}
}

type DailySentiment struct {
	Symbol  string    `gorm:"primaryKey;autoIncrement:false"`
	Date    time.Time `gorm:"primaryKey;autoIncrement:false;type:date"`
	Bearish int
	Bullish int
	Src     string
}
