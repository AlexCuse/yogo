package scanner

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexcuse/yogo/internal/pkg/messaging"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexdrl/zerowater"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type Configuration struct {
	BrokerURL  string
	SignalPort int
	ScanTopic  string
	HitTopic   string
}

type Server struct {
	log          zerolog.Logger
	appctx       context.Context
	scans        []*Scan
	scansInvalid time.Time
	sub          message.Subscriber
	pub          message.Publisher
	cfg          *Configuration
}

func NewServer(cfg *Configuration, appctx context.Context, log zerolog.Logger) (Server, error) {
	wml := zerowater.NewZerologLoggerAdapter(log)

	sub, err := messaging.NewSubscriber(cfg.BrokerURL, "scanner", wml)
	if err != nil {
		return Server{}, err
	}

	pub, err := messaging.NewPublisher(cfg.BrokerURL, "scanner", wml)
	if err != nil {
		return Server{}, err
	}

	server := Server{
		log:    log,
		appctx: appctx,
		pub:    pub,
		sub:    sub,
		cfg:    cfg,
	}

	return server, nil
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
			_ = s.Compile()
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
		msg := <-input
		target := Target{}

		err := json.Unmarshal(msg.Payload, &target)

		if err != nil {
			server.log.Error().Err(err).Msg("unable to unmarshal message")
			msg.Nack()
			continue
		}

		scans, err := server.readScans()

		if err != nil {
			server.log.Error().Err(err).Msg("unable to read scans")
			msg.Nack()
			continue
		}

		success := true

		for _, scan := range scans {
			match, err := scan.Check(target)

			log := server.log.With().Str("scan_name", scan.Name).Logger()

			if err != nil {
				log.Error().Err(err).Msg("problem processing scan")
				success = false
				continue
			} else if match {
				log.Info().Msgf("hit on %s: %+v", target.Quote.Symbol, target)

				rslt, err := json.Marshal(hit{
					RuleName:  scan.Name,
					Symbol:    target.Quote.Symbol,
					QuoteDate: time.Time(target.Quote.Date),
				})

				if err != nil {
					log.Error().Err(err).Msg("problem sending match")
					success = false
					continue
				}

				err = server.pub.Publish(server.cfg.HitTopic, message.NewMessage(uuid.New().String(), rslt))

				if err != nil {
					log.Error().Err(err).Msg("problem sending match")
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
