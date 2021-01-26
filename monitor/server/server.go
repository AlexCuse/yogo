package server

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexcuse/yogo/common/config"
	"github.com/alexcuse/yogo/common/contracts/db"
	fib "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type Server struct {
	pub    message.Publisher
	db     *gorm.DB
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

	dbase, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	dbase.AutoMigrate(db.Asset{})

	iecli := iex.NewClient(cfg.IEXToken, iex.WithBaseURL(cfg.IEXBaseURL))

	server := Server{
		log:    log,
		db:     dbase,
		appctx: appctx,
		app:    f,
		pub:    pub,
		cfg:    cfg,
		iex:    iecli,
	}

	return server
}
func (server Server) Run() error {
	crn := cron.New()

	crn.AddFunc(server.cfg.MonitorCronSchedule, server.watch)

	go crn.Run()

	f := fib.New()
	f.Use(cors.New())

	f.Post("/api/start", server.webStartWatch)

	return f.Listen(fmt.Sprintf(":%d", server.cfg.MonitorPort))
}

func (server Server) webStartWatch(ctx *fib.Ctx) error {
	go server.watch()

	return ctx.SendString("OK")
}

func (server Server) watch() {
	lastHoliday, err := server.iex.PreviousHoliday(server.appctx)

	if err != nil {
		server.log.Error(err.Error())
		return
	} else if lastHoliday.Date.String() == iex.Date(time.Now().AddDate(0, 0, -1)).String() {
		server.log.Infof("Skipping today's monitor run as %s was a market holiday.", lastHoliday.Date.String())
		return
	}

	market, err := server.getQuotes()

	if err != nil {
		server.log.Error(err.Error())
		return
	}

	for _, t := range market {
		server.quote(t)
	}
}

func (server Server) getQuotes() ([]iex.PreviousDay, error) {
	if strings.ToLower(server.cfg.MonitorSource) == "marketexplore" {
		ctx, cancel := context.WithTimeout(server.appctx, 5*time.Second)
		defer cancel()
		res, err := server.iex.PreviousDayMarket(ctx)

		if err == nil {
			for _, q := range res {
				if r := server.db.WithContext(server.appctx).Clauses(clause.OnConflict{DoNothing: true}).Create(db.Asset{Symbol: q.Symbol}); r.Error != nil {
					server.log.Errorf("unable to persist asset: %s", r.Error.Error())
				}
			}
		}

		return res, err
	}

	wl := make([]struct{ Symbol string }, 0)

	if strings.ToLower(server.cfg.MonitorSource) == "market" { //"watchlist" is default
		result := server.db.Select("assets.*").Table(
			"assets",
		).Scan(&wl)

		if result.Error != nil {
			return nil, result.Error
		}
	} else {
		watchlistUrl := fmt.Sprintf("http://watch:%d/api/watch", server.cfg.WatchPort)
		resp, err := http.Get(watchlistUrl)

		if err != nil {
			return nil, err
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(body, &wl)

		if err != nil {
			return nil, err
		}
	}

	//TODO: fix and use iex.PreviousTradingDay
	lastTradeDate, err := server.iex.PreviousTradingDay(server.appctx)

	if err != nil {
		server.log.Errorf("unable to get previous trading day: %s", err.Error())
	}

	results := make([]iex.PreviousDay, 0)

	for _, w := range wl {
		dbMovement := db.Movement{}

		result := server.db.Select("*").Table(
			"movements",
		).Where(
			`symbol = ? and "date" = ?`,
			w.Symbol,
			time.Time(lastTradeDate.Date),
		).Scan(&dbMovement)

		var pd iex.PreviousDay
		var found = false

		if result.RowsAffected == 1 && result.Error == nil {
			server.log.Debugf("got %s movement from DB: %s", w.Symbol, lastTradeDate.Date.String())
			err = json.Unmarshal(dbMovement.Data, &pd)

			if err != nil {
				server.log.Errorf("Unable to unmarshal json for %s(%s): %s (%s)", w.Symbol, lastTradeDate.Date.String(), err.Error(), string(dbMovement.Data))
			} else {
				found = true
			}
		}

		if !found {
			server.log.Debugf("fetching PreviousDay from IEX: %s", w.Symbol)
			ctx, cancel := context.WithTimeout(server.appctx, 500*time.Millisecond)
			pd, err = server.iex.PreviousDay(ctx, w.Symbol)

			if err != nil {
				server.log.Error(err)
			} else {
				found = true
			}

			cancel()
		}

		if found {
			results = append(results, pd)
		}
	}
	return results, nil
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
