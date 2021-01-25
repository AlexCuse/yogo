package stocktwits

import (
	"context"
	"strings"
	"time"

	"github.com/alexcuse/yogo/social-enricher"
	"github.com/rs/zerolog"
)

func NewSentimentCalculator() SentimentCalculator {
	return SentimentCalculator{}
}

type SentimentCalculator struct {
}

func (s *SentimentCalculator) Execute(ctx context.Context, twits *Twits) (*social.Sentiment, error) {
	result := social.Sentiment{
		Sybmol: twits.Symbol.Symbol,
	}

	result.Timestamp = time.Unix(twits.Cursor.Since, 0)
	for _, m := range twits.Messages {
		switch strings.ToLower(m.Entities.Sentiment.Basic) {
		case "bearish":
			result.Bearish++
		case "bullish":
			result.Bullish++
		}
	}

	return &result, nil
}

func (s *SentimentCalculator) Stream(ctx context.Context, twits chan *Twits, sentiment chan *social.Sentiment) {
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-twits:
			if !ok {
				return
			}
			log := zerolog.Ctx(ctx).With().Str("symbol", t.Symbol.Symbol).Logger()
			ctx = log.WithContext(ctx)

			r, err := s.Execute(ctx, t)
			if err != nil {
				log.Err(err).Msg("")
			}

			sentiment <- r
		}
	}
}
