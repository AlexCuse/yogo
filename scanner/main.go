package main

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/alexcuse/yogo/common/config"
	"github.com/alexcuse/yogo/scanner/signals"
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

	sig, err := signals.Load("signals.toml")

	if err != nil {
		panic(err)
	}

	sub, err := kafka.NewSubscriber(kafka.SubscriberConfig{
		Brokers:               []string{cfg.BrokerURL},
		Unmarshaler:           kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: kafka.DefaultSaramaSubscriberConfig(),
	}, wml)

	if err != nil {
		panic(err)
	}

	/*
		pub, err := kafka.NewPublisher(kafka.PublisherConfig{
			Brokers:               []string{cfg.BrokerURL},
			Marshaler:             kafka.DefaultMarshaler{},
			OverwriteSaramaConfig: kafka.DefaultSaramaSyncPublisherConfig(),
		}, wml)
	*/

	if err != nil {
		panic(err)
	}

	input, err := sub.Subscribe(context.Background(), cfg.ScanTopic)

	if err != nil {
		panic(err)
	}

	for {
		select {
		case msg := <-input:
			target := signals.Target{}

			err := json.Unmarshal(msg.Payload, &target)

			if err != nil {
				log.Errorf("unable to unmarshal message: %s", err.Error())
				continue
			}

			for _, signal := range sig {
				hit, err := signal.Check(target)

				if err != nil {
					log.Errorf(err.Error())
				} else if hit {
					//its a match do some shit
					log.Infof("%s hit on %s: %+v", signal.Name, target.Quote.Symbol, target)
				}
			}

			msg.Ack()
		}
	}
}
