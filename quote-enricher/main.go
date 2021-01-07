package main

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexcuse/yogo/common/config"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

func main() {
	cfg, err := config.Load("configuration.toml")
	if err != nil {
		panic(err)
	}

	log := logrus.New()
	if cfg.LogLevel != "" {
		if level, err := logrus.ParseLevel(cfg.LogLevel); err == nil {
			log.SetLevel(level)
		}
	}

	wml := watermill.NewStdLoggerWithOut(log.Out, true, false)

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

			iexClient := iex.NewClient(cfg.IEXToken, iex.WithBaseURL(cfg.IEXBaseURL))

			keystats, err := iexClient.KeyStats(context.Background(), movement.Symbol)

			if err != nil {
				log.Errorf("Could not retrieve key stats: %s", err.Error())
				continue
			}

			statsPl, err := json.Marshal(keystats)

			if err != nil {
				log.Errorf("Could not marshal key stats: %s", err.Error())
			}

			err = pub.Publish(cfg.StatsTopic, message.NewMessage(uuid.New().String(), statsPl))

			if err != nil {
				log.Errorf("Could not publish stats: %s", err.Error())
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
			}

			err = pub.Publish(cfg.ScanTopic, message.NewMessage(msg.UUID, pl))

			if err != nil {
				log.Errorf("Could not publish enriched payload: %s", err.Error())
				continue
			}

			msg.Ack()
		}
	}
}
