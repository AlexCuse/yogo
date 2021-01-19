package server

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexcuse/yogo/common/config"
	fib "github.com/gofiber/fiber/v2"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"time"
)

type Server struct {
	pub    message.Publisher
	log    *logrus.Logger
	appctx context.Context
	app    *fib.App
	cfg    config.Configuration
	iex    *iex.Client
}

func NewServer(cfg config.Configuration, appctx context.Context, log *logrus.Logger, wml watermill.LoggerAdapter) Server {
	f := fib.New()

	pub, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:               []string{cfg.BrokerURL},
		Marshaler:             kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: kafka.DefaultSaramaSyncPublisherConfig(),
	}, wml)

	if err != nil {
		panic(err)
	}

	iecli := iex.NewClient(cfg.IEXToken, iex.WithBaseURL(cfg.IEXBaseURL))

	server := Server{
		log:    log,
		appctx: appctx,
		app:    f,
		pub:    pub,
		cfg:    cfg,
		iex:    iecli,
	}

	return server
}
func (server Server) Run() {
	//always run on startup
	server.watch()

	crn := cron.New()

	//tues-sat 4:30 AM
	crn.AddFunc("30 04 * * 2,3,4,5,6", server.watch)

	crn.Run()
}

func (server Server) watch() {
	previousHolidays, err := server.iex.PreviousHoliday(server.appctx, 1)

	if err != nil {
		server.log.Error(err.Error())
	}

	lastHoliday := previousHolidays[0]

	server.log.Infof("%s == %s, %t", lastHoliday.Date.String(), iex.Date(time.Now().AddDate(0, 0, -1)).String(), lastHoliday.Date == iex.Date(time.Now().AddDate(0, 0, -1)))

	//if there is an error here let it try to run anyway
	if err == nil && lastHoliday.Date.String() == iex.Date(time.Now().AddDate(0, 0, -1)).String() {
		server.log.Infof("Skipping today's server as %s was a market holiday.", lastHoliday.Date.String())
		return
	}

	ctx, cancel := context.WithTimeout(server.appctx, 5*time.Second)
	defer cancel()

	market, err := server.iex.PreviousDayMarket(ctx)

	if err != nil {
		server.log.Error(err.Error())
		return
	}

	for _, t := range market {
		server.quote(t)
	}
}

func (server Server) quote(t iex.PreviousDay) {
	jsn, err := json.Marshal(t)

	if err != nil {
		server.log.Error(err.Error())
		return
	}

	msg := message.NewMessage(uuid.New().String(), jsn)

	err = server.pub.Publish(server.cfg.QuoteTopic, msg)

	if err != nil {
		server.log.Error(err.Error())
	}
}
