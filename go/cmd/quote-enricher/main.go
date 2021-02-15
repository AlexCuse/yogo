package main

import (
	"context"
	"encoding/json"
	"github.com/alexcuse/yogo/internal/pkg/messaging"
	"time"

	"github.com/alexcuse/yogo/internal/pkg/configuration"
	"github.com/alexcuse/yogo/internal/pkg/db"
	"github.com/alexcuse/yogo/internal/pkg/logging"
	"github.com/alexdrl/zerowater"

	"github.com/ThreeDotsLabs/watermill/message"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Config struct {
	DSN        string
	BrokerURL  string
	QuoteTopic string
	IEXToken   string
	IEXBaseURL string
	StatsTopic string
	ScanTopic  string
}

func main() {
	log := logging.Bootstrap()

	errHandler := func(err error) {
		if err != nil {
			log.Fatal().Err(err).Msg("panic")
			panic(err)
		}
	}

	cfg := Config{}
	errHandler(configuration.Unmarshal(&cfg))

	dbase, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	errHandler(err)

	wml := zerowater.NewZerologLoggerAdapter(log)
	sub, err := messaging.NewSubscriber(cfg.BrokerURL, "quote-enricher", wml)

	if err != nil {
		panic(err)
	}

	pub, err := messaging.NewPublisher(cfg.BrokerURL, "quote-enricher", wml)

	if err != nil {
		panic(err)
	}

	input, err := sub.Subscribe(context.Background(), cfg.QuoteTopic)

	if err != nil {
		panic(err)
	}

	for {
		msg := <-input
		movement := iex.PreviousDay{}

		err := json.Unmarshal(msg.Payload, &movement)

		if err != nil {
			log.Error().Err(err).Msg("unable to unmarshal message")
			continue
		}

		dbStats := db.Stats{}

		result := dbase.Select("stats.*").Table(
			"stats",
		).Where(
			"symbol = ? and stats.quote_date = ?",
			movement.Symbol,
			time.Time(movement.Date),
		).Scan(&dbStats)

		keystats := iex.KeyStats{}
		found := false

		log := log.With().Str("symbol", movement.Symbol).Str("movement_date", movement.Date.String()).Logger()

		if result.RowsAffected == 1 && result.Error == nil {
			log.Debug().Msg("got stats from DB")
			err = json.Unmarshal(dbStats.Data, &keystats)

			if err != nil {
				log.Error().Err(err).Msg("failed to unmarshal")
			} else {
				found = true
			}
		}

		if !found {
			log.Debug().Msg("fetching stats from IEX")
			iexClient := iex.NewClient(cfg.IEXToken, iex.WithBaseURL(cfg.IEXBaseURL))

			keystats, err = iexClient.KeyStats(context.Background(), movement.Symbol)

			if err != nil {
				log.Error().Err(err).Msg("failed to get stats from IEX")
			} else {
				found = true
			}
		}

		if !found {
			log.Error().Msg("Could not retrieve key stats")
			msg.Nack()
			continue
		}

		statsPl, err := json.Marshal(struct {
			Stats     iex.KeyStats
			Ticker    string
			QuoteDate time.Time
		}{
			Stats:     keystats,
			Ticker:    movement.Symbol,
			QuoteDate: time.Time(movement.Date),
		})
		if err != nil {
			log.Error().Err(err).Msg("Could not marshal key stats")
			msg.Nack()
			continue
		}

		err = pub.Publish(cfg.StatsTopic, message.NewMessage(uuid.New().String(), statsPl))

		if err != nil {
			log.Error().Err(err).Msg("Could not publish stats")
			msg.Nack()
			continue
		}

		scannable := struct {
			Quote iex.PreviousDay `json:"quote"`
			Stats iex.KeyStats    `json:"stats"`
		}{
			Quote: movement,
			Stats: keystats,
		}

		pl, err := json.Marshal(scannable)

		if err != nil {
			log.Error().Err(err).Msg("Could not marshal enriched payload:")
			msg.Nack()
			continue
		}

		err = pub.Publish(cfg.ScanTopic, message.NewMessage(msg.UUID, pl))

		if err != nil {
			log.Error().Err(err).Msg("Could not publish enriched payload")
			msg.Nack()
			continue
		}

		msg.Ack()
	}
}
