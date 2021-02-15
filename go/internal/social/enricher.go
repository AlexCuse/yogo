package social

import (
	"context"
	"encoding/json"

	"github.com/ThreeDotsLabs/watermill/message"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func NewEnricher(dl DailySentimenter, pub message.Publisher) Enricher {
	return Enricher{
		dl:  dl,
		pub: pub,
	}
}

type Enricher struct {
	dl  DailySentimenter
	pub message.Publisher
}

func (e *Enricher) Execute(ctx context.Context, input <-chan *message.Message, socialTopic string) {
	for {
		msg := <-input
		log := zerolog.Ctx(ctx).With().Str("message_id", msg.UUID).Logger()
		movement := iex.PreviousDay{}

		err := json.Unmarshal(msg.Payload, &movement)
		if err != nil {
			log.Error().Err(err).RawJSON("payload", msg.Payload).Msg("could not unmarshal to movement")
			msg.Nack()
			continue
		}

		symbol := movement.Symbol
		log = log.With().Str("symbol", symbol).Logger()

		sent, err := e.dl.Get(movement.Symbol)
		if err != nil {
			log.Error().Err(err).RawJSON("payload", msg.Payload).Msg("could not unmarshal")
			msg.Nack()
			continue
		}

		m := make(map[string]interface{})
		err = json.Unmarshal([]byte(msg.Payload), &m)
		if err != nil {
			log.Error().Err(err).RawJSON("payload", msg.Payload).Msg("could not unmarshal")
			msg.Nack()
			continue
		}

		social := make(map[string]struct {
			Bearish int
			Bullish int
		})
		for feed, snapshots := range sent.Trend {
			agr := struct {
				Bearish int
				Bullish int
			}{}
			for _, s := range snapshots {
				agr.Bearish += s.Bearish
				agr.Bullish += s.Bullish
			}

			social[feed] = agr
		}

		m["social"] = social
		enrichedPayload, err := json.Marshal(m)
		if err != nil {
			log.Error().Err(err).Interface("message", m).Msgf("could marshal not to json for topic %s", socialTopic)
			msg.Nack()
			continue
		}

		err = e.pub.Publish(socialTopic, message.NewMessage(uuid.New().String(), enrichedPayload))
		if err != nil {
			log.Error().Err(err).RawJSON("payload", enrichedPayload).Msgf("could not publish on to %q", socialTopic)
			msg.Nack()
			continue
		}

		log.Debug().RawJSON("payload", enrichedPayload).Msgf("enriched %q", symbol)

		msg.Ack()
	}

}
