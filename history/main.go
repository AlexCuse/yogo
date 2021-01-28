package main

import (
	"context"
	"encoding/json"
	"os"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexdrl/zerowater"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"

	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/alexcuse/yogo/common/contracts/db"
	iex "github.com/goinvest/iexcloud/v2"
	"gorm.io/datatypes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type config struct {
	DSN            string
	BrokerURL      string
	QuoteTopic     string
	StatsTopic     string
	HitTopic       string
	SentimentTopic string
}

func main() {
	log := zerolog.New(os.Stdout).With().Logger()
	ctx := log.WithContext(context.Background())

	errHandler := func(err error) {
		if err != nil {
			log.Fatal().Err(err).Msg("panic")
			panic(err)
		}
	}

	cfg := &config{}
	viper.SetEnvPrefix("yogo")
	viper.AutomaticEnv()
	viper.AddConfigPath(".")
	viper.SetConfigName("configuration")
	err := viper.ReadInConfig()
	errHandler(err)
	errHandler(viper.Unmarshal(cfg))

	wml := zerowater.NewZerologLoggerAdapter(log)
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

	err = dbase.AutoMigrate(&sentiment{})
	if err != nil {
		panic(err)
	}

	sentiments, err := sub.Subscribe(ctx, cfg.SentimentTopic)
	if err != nil {
		panic(err)
	}
	go processSentiments(dbase, ctx, sentiments, log)

	select {
	case <-ctx.Done():
		return
	}
}

func processMovements(dbase *gorm.DB, ctx context.Context, input <-chan *message.Message, log zerolog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-input:
			movement := iex.PreviousDay{}

			if err := json.Unmarshal(msg.Payload, &movement); err != nil {
				log.Error().Err(err).Msg("unable to unmarshal message")
				msg.Nack()
				continue
			}

			if r := dbase.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(db.Movement{
				Symbol: movement.Symbol,
				Date:   time.Time(movement.Date),
				Data:   datatypes.JSON(msg.Payload),
			}); r.Error != nil {
				log.Error().Err(r.Error).Msg("unable to persist movement")
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}

func processStats(dbase *gorm.DB, ctx context.Context, input <-chan *message.Message, log zerolog.Logger) {
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
				log.Error().Err(err).Msg("unable to unmarshal message")
				msg.Nack()
				continue
			}

			jsn, err := json.Marshal(tickerStats.Stats)
			if err != nil {
				log.Error().Err(err).Msg("unable to marshal stats")
			}

			if r := dbase.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(db.Stats{
				Symbol:    tickerStats.Ticker,
				QuoteDate: tickerStats.QuoteDate,
				Data:      datatypes.JSON(jsn),
			}); r.Error != nil {
				log.Error().Err(r.Error).Msg("unable to persist movement")
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}

func processHits(dbase *gorm.DB, ctx context.Context, input <-chan *message.Message, log zerolog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-input:
			hit := db.Hit{}

			if err := json.Unmarshal(msg.Payload, &hit); err != nil {
				log.Error().Err(err).Msg("unable to unmarshal message")
				msg.Nack()
				continue
			}

			if r := dbase.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(hit); r.Error != nil {
				log.Error().Err(r.Error).Msg("unable to persist hit")
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}

func processSentiments(dbase *gorm.DB, ctx context.Context, input <-chan *message.Message, log zerolog.Logger) {
	for {
		select {
		case <-ctx.Done():
			return

		case msg := <-input:
			s := sentiment{}

			if err := json.Unmarshal(msg.Payload, &s); err != nil {
				log.Error().Err(err).Msg("unable to unmarshal message")
				msg.Nack()
				continue
			}

			if r := dbase.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(s); r.Error != nil {
				log.Error().Err(r.Error).Msg("unable to persist sentiment")
				msg.Nack()
				continue
			}

			msg.Ack()
		}
	}
}

type sentiment struct {
	Sybmol    string    `gorm:"primaryKey;autoIncrement:false" json:"sybmol,omitempty"`
	Src       string    `gorm:"primaryKey;autoIncrement:false" json:"src,omitempty"`
	Bearish   int       `json:"bearish,omitempty"`
	Bullish   int       `json:"bullish,omitempty"`
	Timestamp time.Time `gorm:"primaryKey;autoIncrement:false" json:"timestamp,omitempty"`
}
