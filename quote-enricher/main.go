package main

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexcuse/yogo/common"
	"github.com/alexcuse/yogo/common/contracts/db"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"time"
)

func main() {
	cfg, log, wml := common.Bootstrap("configuration.toml")

	dbase, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})

	sub, err := kafka.NewSubscriber(kafka.SubscriberConfig{
		Brokers:               []string{cfg.BrokerURL},
		Unmarshaler:           kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: kafka.DefaultSaramaSubscriberConfig(),
	}, wml)

	if err != nil {
		panic(err)
	}

	pub, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:               []string{cfg.BrokerURL},
		Marshaler:             kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: kafka.DefaultSaramaSyncPublisherConfig(),
	}, wml)

	if err != nil {
		panic(err)
	}

	input, err := sub.Subscribe(context.Background(), cfg.QuoteTopic)

	if err != nil {
		panic(err)
	}

	for {
		select {
		case msg := <-input:
			movement := iex.PreviousDay{}

			err := json.Unmarshal(msg.Payload, &movement)

			if err != nil {
				log.Errorf("unable to unmarshal message: %s", err.Error())
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

			if result.RowsAffected == 1 && result.Error == nil {
				log.Debugf("got %s stats from DB: %s", movement.Symbol, movement.Date.String())
				err = json.Unmarshal(dbStats.Data, &keystats)

				if err != nil {
					log.Errorf("failed to unmarshal %s (%s) stats: %s", movement.Symbol, movement.Date.String(), err.Error())
				} else {
					found = true
				}
			}

			if !found {
				log.Debugf("fetching stats from IEX: %s / %s", movement.Symbol, movement.Date.String())
				iexClient := iex.NewClient(cfg.IEXToken, iex.WithBaseURL(cfg.IEXBaseURL))

				keystats, err = iexClient.KeyStats(context.Background(), movement.Symbol)

				if err != nil {
					log.Errorf("failed to get %s (%s) stats from IEX: %s", movement.Symbol, movement.Date.String(), err.Error())
				} else {
					found = true
				}
			}

			if !found {
				errString := "no error"
				if err != nil {
					errString = err.Error()
				}
				log.Errorf("Could not retrieve key stats: %s", errString)
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
				log.Errorf("Could not marshal key stats: %s", err.Error())
				msg.Nack()
				continue
			}

			err = pub.Publish(cfg.StatsTopic, message.NewMessage(uuid.New().String(), statsPl))

			if err != nil {
				log.Errorf("Could not publish stats: %s", err.Error())
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
				log.Errorf("Could not marshal enriched payload: %s", err.Error())
				msg.Nack()
				continue
			}

			err = pub.Publish(cfg.ScanTopic, message.NewMessage(msg.UUID, pl))

			if err != nil {
				log.Errorf("Could not publish enriched payload: %s", err.Error())
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}
