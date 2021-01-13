package main

import (
	"context"
	"encoding/json"
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

	ctx := context.Background()
	input, err := sub.Subscribe(ctx, cfg.QuoteTopic)
	if err != nil {
		panic(err)
	}

	for {
		select {
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

type Movement struct {
	Symbol string    `gorm:"primaryKey;autoIncrement:false"`
	Date   time.Time `gorm:"primaryKey;autoIncrement:false"`
	Data   datatypes.JSON
}
