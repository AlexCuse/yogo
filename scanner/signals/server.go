package signals

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexcuse/yogo/common/config"
	fib "github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
	"time"
)

type Server struct {
	db     *gorm.DB
	log    *logrus.Logger
	appctx context.Context
	app    *fib.App
	scans  map[string]*Scan
	sub    *kafka.Subscriber
	pub    *kafka.Publisher
	cfg    config.Configuration
}

func NewServer(cfg config.Configuration, appctx context.Context, db *gorm.DB, log *logrus.Logger, wml watermill.LoggerAdapter) Server {
	f := fib.New()

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
		db:     db,
		log:    log,
		appctx: appctx,
		app:    f,
		scans:  make(map[string]*Scan),
		pub:    pub,
		sub:    sub,
		cfg:    cfg,
	}

	err = server.loadScans()

	if err != nil {
		panic(err)
	}

	return server
}

func (server Server) loadScans() error {
	signals := make([]Signal, 0)

	result := server.db.Find(&signals)

	if result.RowsAffected == 0 {
		return nil
	}

	for _, sig := range signals {
		err := server.registerScan(sig)
		if err != nil {
			return err
		}
	}

	return nil
}

func (server Server) registerScan(sig Signal) error {
	scan, err := NewScan(sig)

	if err != nil {
		return err
	}

	name := strings.ToLower(scan.Name)

	server.scans[name] = scan
	return nil
}

func (server Server) Save(ctx *fib.Ctx) error {
	s := Signal{}

	err := json.Unmarshal(ctx.Body(), &s)

	if err != nil {
		return handleError(ctx, err)
	}

	if r := server.db.WithContext(server.appctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(s); r.Error != nil {
		server.log.Errorf("unable to persist signal: %s", r.Error.Error())
	}

	if err != nil {
		return handleError(ctx, err)
	}

	err = server.registerScan(s)

	if err != nil {
		return handleError(ctx, err)
	}

	return nil
}

func (server Server) Delete(ctx *fib.Ctx) error {
	name := ctx.Query("name")

	if name == "" {
		return nil
	}

	result := server.db.WithContext(server.appctx).Delete(&Signal{Name: name})

	if result.RowsAffected == 0 {
		return errors.New(fmt.Sprintf("failed to delete `%s`", name))
	} else if result.RowsAffected < 1 {
		return errors.New(fmt.Sprintf("deleted more than 1 row `%s`", name))
	}

	delete(server.scans, strings.ToLower(name))

	return nil
}

func (server Server) Run() error {
	server.app.Post("api/signal", server.Save)
	server.app.Put("api/signal", server.Save)
	server.app.Delete("api/signal", server.Delete)

	go func(s Server) {
		s.background()
	}(server)

	return server.app.Listen(fmt.Sprintf(":%d", server.cfg.SignalPort))
}

func handleError(ctx *fib.Ctx, err error) error {
	ctx.Status(500)
	ctx.WriteString(err.Error())
	return err
}

func (server Server) readScans() []*Scan {
	scans := make([]*Scan, 0)

	for _, s := range server.scans {
		scans = append(scans, s)
	}

	return scans
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
				continue
			}

			for _, scan := range server.readScans() {
				match, err := scan.Check(target)

				if err != nil {
					server.log.Errorf("problem processing scan %s: (%s)", scan.Name, err.Error())
				} else if match {
					server.log.Infof("%s hit on %s: %+v", scan.Name, target.Quote.Symbol, target)

					rslt, err := json.Marshal(hit{
						RuleName:  scan.Name,
						Symbol:    target.Quote.Symbol,
						QuoteDate: time.Time(target.Quote.Date),
					})

					if err != nil {
						server.log.Errorf("problem sending match for %s: (%s)", scan.Name, err.Error())
						continue
					}

					err = server.pub.Publish(server.cfg.HitTopic, message.NewMessage(uuid.New().String(), rslt))

					if err != nil {
						server.log.Errorf("problem sending match for %s: (%s)", scan.Name, err.Error())
					}
				}
			}

			msg.Ack()
		}
	}
}
