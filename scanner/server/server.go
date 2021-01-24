package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexcuse/yogo/common/config"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"time"
)

type Server struct {
	log          *logrus.Logger
	appctx       context.Context
	scans        []*Scan
	scansInvalid time.Time
	sub          *kafka.Subscriber
	pub          *kafka.Publisher
	cfg          config.Configuration
}

func NewServer(cfg config.Configuration, appctx context.Context, log *logrus.Logger, wml watermill.LoggerAdapter) Server {
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

	server := Server{
		log:    log,
		appctx: appctx,
		pub:    pub,
		sub:    sub,
		cfg:    cfg,
	}

	return server
}

func (server Server) Run() error {
	server.background()

	return nil
}

func (server Server) readScans() ([]*Scan, error) {
	if server.scans == nil || server.scansInvalid.Before(time.Now()) {
		resp, err := http.Get(fmt.Sprintf("http://signals:%d/api/signals", server.cfg.SignalPort))
		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		jsn, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return nil, err
		}

		scans := []*Scan{}

		err = json.Unmarshal(jsn, &scans)

		if err != nil {
			return nil, err
		}

		for _, s := range scans {
			s.Compile()
		}
		server.scansInvalid = time.Now().Add(15 * time.Minute)
		server.scans = scans
	}
	return server.scans, nil
}

func (server Server) background() {
	input, err := server.sub.Subscribe(server.appctx, server.cfg.ScanTopic)

	if err != nil {
		panic(err)
	}

	for {
		select {
		case msg := <-input:
			target := Target{}

			err := json.Unmarshal(msg.Payload, &target)

			if err != nil {
				server.log.Errorf("unable to unmarshal message: %s", err.Error())
				msg.Nack()
				continue
			}

			scans, err := server.readScans()

			if err != nil {
				server.log.Errorf("unable to read scans: %s", err.Error())
				msg.Nack()
				continue
			}

			success := true

			for _, scan := range scans {
				match, err := scan.Check(target)

				if err != nil {
					server.log.Errorf("problem processing scan %s: (%s)", scan.Name, err.Error())
					success = false
					continue
				} else if match {
					server.log.Infof("%s hit on %s: %+v", scan.Name, target.Quote.Symbol, target)

					rslt, err := json.Marshal(hit{
						RuleName:  scan.Name,
						Symbol:    target.Quote.Symbol,
						QuoteDate: time.Time(target.Quote.Date),
					})

					if err != nil {
						server.log.Errorf("problem sending match for %s: (%s)", scan.Name, err.Error())
						success = false
						continue
					}

					err = server.pub.Publish(server.cfg.HitTopic, message.NewMessage(uuid.New().String(), rslt))

					if err != nil {
						server.log.Errorf("problem sending match for %s: (%s)", scan.Name, err.Error())
						success = false
					}
				}
			}

			if success {
				msg.Ack()
			} else {
				msg.Nack()
			}
		}
	}
}
