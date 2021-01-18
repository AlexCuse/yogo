package main

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/sirupsen/logrus"
	"time"

	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/alexcuse/yogo/common"
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

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&Movement{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&Stats{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&Hit{})
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	movements, err := sub.Subscribe(ctx, cfg.QuoteTopic)
	if err != nil {
		panic(err)
	}

	//add waitgroup etc
	go processMovements(db, ctx, movements, log)

	stats, err := sub.Subscribe(ctx, cfg.StatsTopic)
	if err != nil {
		panic(err)
	}

	go processStats(db, ctx, stats, log)

	hits, err := sub.Subscribe(ctx, cfg.HitTopic)
	if err != nil {
		panic(err)
	}

	go processHits(db, ctx, hits, log)

	select {
	case <-ctx.Done():
		return
	}
}

func processMovements(db *gorm.DB, ctx context.Context, input <-chan *message.Message, log *logrus.Logger) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-input:
			movement := iex.PreviousDay{}

			if err := json.Unmarshal(msg.Payload, &movement); err != nil {
				log.Errorf("unable to unmarshal message: %s", err.Error())
				continue
			}

			if r := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(Movement{
				Symbol: movement.Symbol,
				Date:   time.Time(movement.Date),
				Data:   datatypes.JSON(msg.Payload),
			}); r.Error != nil {
				log.Errorf("unable to persist movement: %s", r.Error.Error())
				continue
			}

			msg.Ack()
		}
	}
}

func processStats(db *gorm.DB, ctx context.Context, input <-chan *message.Message, log *logrus.Logger) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-input:
			tickerStats := struct {
				Stats  iex.KeyStats
				Ticker string
			}{}

			if err := json.Unmarshal(msg.Payload, &tickerStats); err != nil {
				log.Errorf("unable to unmarshal message: %s", err.Error())
				continue
			}

			if r := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(Stats{
				Symbol: tickerStats.Ticker,
				Date:   time.Now(),
				Data:   datatypes.JSON(msg.Payload),
			}); r.Error != nil {
				log.Errorf("unable to persist stats: %s", r.Error.Error())
				continue
			}

			msg.Ack()
		}
	}
}

func processHits(db *gorm.DB, ctx context.Context, input <-chan *message.Message, log *logrus.Logger) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-input:
			hit := Hit{}

			if err := json.Unmarshal(msg.Payload, &hit); err != nil {
				log.Errorf("unable to unmarshal message: %s", err.Error())
				continue
			}

			if r := db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(hit); r.Error != nil {
				log.Errorf("unable to persist hit: %s", r.Error.Error())
				continue
			}

			msg.Ack()
		}
	}
}

type Movement struct {
	Symbol string    `gorm:"primaryKey;autoIncrement:false"`
	Date   time.Time `gorm:"primaryKey;autoIncrement:false"`
	Data   datatypes.JSON
}

type Stats struct {
	Symbol string    `gorm:"primaryKey;autoIncrement:false"`
	Date   time.Time `gorm:"primaryKey;autoIncrement:false"`
	Data   datatypes.JSON
}

type Hit struct {
	RuleName  string    `gorm:"primaryKey;autoIncrement:false"`
	Symbol    string    `gorm:"primaryKey;autoIncrement:false"`
	QuoteDate time.Time `gorm:"primaryKey;autoIncrement:false"`
}
