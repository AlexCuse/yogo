package main

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/alexcuse/yogo/common"
	"github.com/alexcuse/yogo/common/contracts/db"
	iex "github.com/goinvest/iexcloud/v2"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	cfg, log, wml := common.Bootstrap("configuration.toml")

	sub, err := kafka.NewSubscriber(kafka.SubscriberConfig{
		Brokers:               []string{cfg.BrokerURL},
		Unmarshaler:           kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: kafka.DefaultSaramaSubscriberConfig(),
	}, wml)
	if err != nil {
		panic(err)
	}

	dbase, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = dbase.AutoMigrate(&db.Movement{})
	if err != nil {
		panic(err)
	}

	err = dbase.AutoMigrate(&db.Stats{})
	if err != nil {
		panic(err)
	}

	err = dbase.AutoMigrate(&db.Hit{})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	movements, err := sub.Subscribe(ctx, cfg.QuoteTopic)
	if err != nil {
		panic(err)
	}

	//add waitgroup etc
	go processMovements(dbase, ctx, movements, log)

	stats, err := sub.Subscribe(ctx, cfg.StatsTopic)
	if err != nil {
		panic(err)
	}

	go processStats(dbase, ctx, stats, log)

	hits, err := sub.Subscribe(ctx, cfg.HitTopic)
	if err != nil {
		panic(err)
	}

	go processHits(dbase, ctx, hits, log)

	select {
	case <-ctx.Done():
		return
	}
}

func processMovements(dbase *gorm.DB, ctx context.Context, input <-chan *message.Message, log *logrus.Logger) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-input:
			movement := iex.PreviousDay{}

			if err := json.Unmarshal(msg.Payload, &movement); err != nil {
				log.Errorf("unable to unmarshal message: %s", err.Error())
				msg.Nack()
				continue
			}

			if r := dbase.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(db.Movement{
				Symbol: movement.Symbol,
				Date:   time.Time(movement.Date),
				Data:   datatypes.JSON(msg.Payload),
			}); r.Error != nil {
				log.Errorf("unable to persist movement: %s", r.Error.Error())
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}

func processStats(dbase *gorm.DB, ctx context.Context, input <-chan *message.Message, log *logrus.Logger) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-input:
			tickerStats := struct {
				Stats     iex.KeyStats
				Ticker    string
				QuoteDate time.Time
			}{}

			if err := json.Unmarshal(msg.Payload, &tickerStats); err != nil {
				log.Errorf("unable to unmarshal message: %s", err.Error())
				msg.Nack()
				continue
			}

			jsn, err := json.Marshal(tickerStats.Stats)

			if err != nil {
				log.Errorf("unable to marshal stats")
			}

			if r := dbase.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(db.Stats{
				Symbol:    tickerStats.Ticker,
				QuoteDate: tickerStats.QuoteDate,
				Data:      datatypes.JSON(jsn),
			}); r.Error != nil {
				log.Errorf("unable to persist stats: %s", r.Error.Error())
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}

func processHits(dbase *gorm.DB, ctx context.Context, input <-chan *message.Message, log *logrus.Logger) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-input:
			hit := db.Hit{}

			if err := json.Unmarshal(msg.Payload, &hit); err != nil {
				log.Errorf("unable to unmarshal message: %s", err.Error())
				msg.Nack()
				continue
			}

			if r := dbase.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(hit); r.Error != nil {
				log.Errorf("unable to persist hit: %s", r.Error.Error())
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}
