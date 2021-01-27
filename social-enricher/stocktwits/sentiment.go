package stocktwits

import (
	"context"
	"strings"

	"github.com/alexcuse/yogo/social-enricher"
	"github.com/rs/zerolog"
)

func NewSentimentCalculator() SentimentCalculator {
	return SentimentCalculator{}
}

type SentimentCalculator struct {
}

func (s *SentimentCalculator) Execute(ctx context.Context, twits Twits) (social.SentimentSnapshot, error) {
	result := social.SentimentSnapshot{
		Sybmol: twits.Symbol.Symbol,
		Src:    "stocktwits",
	}

	if len(twits.Messages) == 0 {
		return result, nil
	}

	result.Timestamp = twits.Messages[0].CreatedAt
	for _, m := range twits.Messages {
		switch strings.ToLower(m.Entities.Sentiment.Basic) {
		case "bearish":
			result.Bearish++
		case "bullish":
			result.Bullish++
		}
	}

	return result, nil
}

func (s *SentimentCalculator) Stream(ctx context.Context, twits <-chan Twits, sentiment chan social.SentimentSnapshot) {
	for {
		select {
		case <-ctx.Done():
			return
		case t, ok := <-twits:
			if !ok {
				return
			}
			log := zerolog.Ctx(ctx).With().Str("symbol", t.Symbol.Symbol).Logger()
			ctx := log.WithContext(ctx)

			r, err := s.Execute(ctx, t)
			if err != nil {
				log.Err(err).Msg("")
			}

			sentiment <- r

			log.Debug().Msg("snapshot")
		}
	}
}
