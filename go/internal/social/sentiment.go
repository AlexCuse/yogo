package social

import (
	"context"
	"encoding/json"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/google/uuid"
	"github.com/rs/zerolog"

	"github.com/philippgille/gokv"
)

type SentimentSnapshot struct {
	Sentiment
	Sybmol string `json:"sybmol,omitempty"`
	Src    string `json:"src,omitempty"`
}

func NewDailySentiment(s SentimentSnapshot) DailySentiment {
	d := DailySentiment{
		Symbol: s.Sybmol,
		Trend:  map[string][]Sentiment{},
	}

	d.Trend[s.Src] = []Sentiment{
		{
			Bearish:   s.Bearish,
			Bullish:   s.Bullish,
			Timestamp: s.Timestamp,
		},
	}

	return d
}

type Sentiment struct {
	Bearish   int       `json:"bearish,omitempty"`
	Bullish   int       `json:"bullish,omitempty"`
	Timestamp time.Time `json:"timestamp,omitempty"`
}
type DailySentiment struct {
	Symbol string `json:"symbol,omitempty"`
	Trend  map[string][]Sentiment
}

func (d *DailySentiment) Add(snapshot SentimentSnapshot) error {
	s := Sentiment{
		Bearish:   snapshot.Bearish,
		Bullish:   snapshot.Bullish,
		Timestamp: snapshot.Timestamp,
	}

	trend, ok := d.Trend[snapshot.Src]
	if !ok {
		d.Trend[snapshot.Src] = []Sentiment{
			s,
		}
		return nil
	}

	d.Trim(time.Now())

	d.Trend[snapshot.Src] = append(trend, s)

	return nil
}

func (d *DailySentiment) Trim(t time.Time) {
	for src, trend := range d.Trend {
		compacted := trend[:0]
		for _, ss := range trend {
			if t.Sub(ss.Timestamp).Hours() <= 24 {
				compacted = append(compacted, ss)
			}
		}
		d.Trend[src] = compacted
	}
}

type DailySentimenter interface {
	Get(symbol string) (DailySentiment, error)
}

func NewDailySentimentAggregator(store gokv.Store) (DailySentimentAggregator, error) {
	return DailySentimentAggregator{
		store: store,
	}, nil
}

type DailySentimentAggregator struct {
	store gokv.Store
}

func (t *DailySentimentAggregator) Get(symbol string) (DailySentiment, error) {
	s := DailySentiment{}
	_, err := t.store.Get(symbol, s)
	if err != nil {
		return s, err
	}

	return s, nil
}

func (t *DailySentimentAggregator) Add(ctx context.Context, snapshot SentimentSnapshot) (DailySentiment, error) {
	s := DailySentiment{}
	exists, err := t.store.Get(snapshot.Sybmol, &s)
	if err != nil {
		return s, err
	}
	if !exists {
		s = NewDailySentiment(snapshot)
	} else {
		if err = s.Add(snapshot); err != nil {
			return s, err
		}
	}

	return s, t.store.Set(snapshot.Sybmol, s)
}

func (t *DailySentimentAggregator) Stream(ctx context.Context, sentiments <-chan SentimentSnapshot) {
	for {
		select {
		case <-ctx.Done():
			return
		case s, ok := <-sentiments:
			if !ok {
				return
			}
			log := zerolog.Ctx(ctx).With().Str("symbol", s.Sybmol).Logger()
			ctx := log.WithContext(ctx)

			ds, err := t.Add(ctx, s)
			if err != nil {
				log.Err(err).Msg("")
			}

			log.Debug().Interface("daily", ds).Msg("daily sentiment updated")
		}
	}
}

func NewSentimentHistorian(pub message.Publisher) SentimentHistorian {
	return SentimentHistorian{
		pub: pub,
	}
}

type SentimentHistorian struct {
	pub message.Publisher
}

func (h *SentimentHistorian) Stream(ctx context.Context, sentiments <-chan SentimentSnapshot, topic string) {
	for {
		select {
		case <-ctx.Done():
			return
		case s, ok := <-sentiments:
			if !ok {
				return
			}
			log := zerolog.Ctx(ctx).With().Str("symbol", s.Sybmol).Logger()

			payload, err := json.Marshal(s)
			if err != nil {
				log.Error().Err(err).Interface("sentiment", s).Msgf("could not marshal to json")
				continue
			}

			err = h.pub.Publish(topic, message.NewMessage(uuid.New().String(), payload))
			if err != nil {
				log.Error().Err(err).RawJSON("payload", payload).Msgf("could not publish on to %q", topic)
				continue
			}

			log.Debug().RawJSON("payload", payload).Msg("published")
		}
	}
}
