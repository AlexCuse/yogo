package monitor

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/alexcuse/yogo/internal/pkg/messaging"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexcuse/yogo/internal/pkg/db"
	"github.com/alexdrl/zerowater"
	fib "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Configuration struct {
	IEXToken            string
	IEXBaseURL          string
	BrokerURL           string
	QuoteTopic          string
	StatsTopic          string
	HitTopic            string
	ScanTopic           string
	LogLevel            string
	DSN                 string
	SignalPort          int
	WatchPort           int
	MonitorSource       string
	MonitorCronSchedule string
	MonitorPort         int
}

type Server struct {
	pub    message.Publisher
	db     *gorm.DB
	log    zerolog.Logger
	appctx context.Context
	app    *fib.App
	cfg    *Configuration
	iex    *iex.Client
}

func NewServer(cfg *Configuration, appctx context.Context, log zerolog.Logger) Server {
	f := fib.New()

	pub, err := messaging.NewPublisher(cfg.BrokerURL, "monitor", zerowater.NewZerologLoggerAdapter(log))
	if err != nil {
		log.Error().Msgf("Failed to connect to %s: %s", cfg.BrokerURL, err.Error())
		panic(err)
	}

	dbase, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = dbase.AutoMigrate(db.Asset{})
	if err != nil {
		panic(err)
	}

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
	go func(s Server) {
		crn := cron.New(cron.WithLocation(time.FixedZone("Eastern", -5*60*60)))

		_, _ = crn.AddFunc(s.cfg.MonitorCronSchedule, s.watch)

		crn.Run()
	}(server)

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
		server.log.Error().Err(err).Msg("iex.PreviousHoliday")
		return
	} else if lastHoliday.Date.String() == iex.Date(time.Now().AddDate(0, 0, -1)).String() {
		server.log.Info().Msgf("Skipping today's monitor run as %s was a market holiday.", lastHoliday.Date.String())
		return
	}

	_, err = server.getQuotes()
	if err != nil {
		server.log.Error().Err(err).Msg("server.getQuotes")
		return
	}
}

func (server Server) getQuotes() ([]iex.PreviousDay, error) {
	if strings.ToLower(server.cfg.MonitorSource) == "marketexplore" {
		ctx, cancel := context.WithTimeout(server.appctx, 5*time.Second)
		defer cancel()
		res, err := server.iex.PreviousDayMarket(ctx)

		if err == nil {
			for _, q := range res {
				server.quote(q)
				if r := server.db.WithContext(server.appctx).Clauses(clause.OnConflict{DoNothing: true}).Create(db.Asset{Symbol: q.Symbol}); r.Error != nil {
					server.log.Error().Err(err).Msg("unable to persist asset")
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
		server.log.Error().Err(err).Msg("unable to get previous trading day")
	}

	results := make([]iex.PreviousDay, 0)

	for _, w := range wl {
		log := server.log.With().Str("symbol", w.Symbol).Str("last_trade_date", lastTradeDate.Date.String()).Logger()

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
			log.Debug().Msg("got movement from DB")
			err = json.Unmarshal(dbMovement.Data, &pd)

			if err != nil {
				log.Error().Err(err).Msgf("Unable to unmarshal json %s", string(dbMovement.Data))
			} else {
				found = true
			}
		}

		if !found {
			log.Debug().Msg("fetching PreviousDay from IEX")
			ctx, cancel := context.WithTimeout(server.appctx, 500*time.Millisecond)
			pd, err = server.iex.PreviousDay(ctx, w.Symbol)
			if err != nil {
				log.Error().Err(err).Msg("failed to fetch from IEX")
			} else {
				found = true
			}

			cancel()
		}

		if found {
			server.quote(pd)
			results = append(results, pd)
		}
	}
	return results, nil
}

func (server Server) quote(t iex.PreviousDay) {
	jsn, err := json.Marshal(t)
	if err != nil {
		log.Error().Err(err).Msgf("unable to json.marshal %v", t)
		return
	}

	msg := message.NewMessage(uuid.New().String(), jsn)

	err = server.pub.Publish(server.cfg.QuoteTopic, msg)
	if err != nil {
		log.Error().Err(err).Msg("failed to publish")
	}
}
