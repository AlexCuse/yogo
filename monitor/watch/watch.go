package watch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexcuse/yogo/common/config"
	fib "github.com/gofiber/fiber/v2"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/google/uuid"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/gofiber/fiber/v2/middleware/cors"
)

type Watch struct {
	Symbol string `gorm:"primaryKey;autoIncrement:false" json:"symbol,omitempty"`
}

type Server struct {
	db        *gorm.DB
	log       *logrus.Logger
	appctx    context.Context
	app       *fib.App
	watchlist map[string]Watch
	pub       *kafka.Publisher
	cfg       config.Configuration
	iex       *iex.Client
}

func NewServer(cfg config.Configuration, appctx context.Context, db *gorm.DB, log *logrus.Logger, wml watermill.LoggerAdapter) Server {
	f := fib.New()
	f.Use(cors.New())

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
		db:        db,
		log:       log,
		appctx:    appctx,
		app:       f,
		watchlist: make(map[string]Watch),
		pub:       pub,
		cfg:       cfg,
		iex:       iecli,
	}

	err = server.loadScans()

	if err != nil {
		panic(err)
	}

	return server
}

func (server Server) loadScans() error {
	watching := make([]Watch, 0)

	result := server.db.Find(&watching)

	if result.RowsAffected == 0 {
		return nil
	}

	for _, watch := range watching {
		err := server.registerWatch(watch)
		if err != nil {
			return err
		}
	}

	return nil
}

func (server Server) registerWatch(w Watch) error {
	server.watchlist[strings.ToLower(w.Symbol)] = w
	return nil
}

func (server Server) Index(ctx *fib.Ctx) error {
	watching := make([]Watch, 0)

	result := server.db.Find(&watching)

	if result.Error != nil {
		return handleError(ctx, result.Error)
	}

	if result.RowsAffected == 0 {
		ctx.Status(404)
		return nil
	}

	return ctx.JSON(watching)
}

func (server Server) Save(ctx *fib.Ctx) error {
	w := Watch{}

	err := json.Unmarshal(ctx.Body(), &w)

	if err != nil {
		return handleError(ctx, err)
	}

	if r := server.db.WithContext(server.appctx).Clauses(clause.OnConflict{DoNothing: true}).Create(w); r.Error != nil {
		server.log.Errorf("unable to persist watch: %s", r.Error.Error())
	}

	if err != nil {
		return handleError(ctx, err)
	}

	err = server.registerWatch(w)

	if err != nil {
		return handleError(ctx, err)
	}

	return nil
}

func (server Server) Delete(ctx *fib.Ctx) error {
	sym := ctx.Query("symbol")

	if sym == "" {
		return nil
	}

	result := server.db.WithContext(server.appctx).Delete(&Watch{Symbol: sym})

	if result.RowsAffected == 0 {
		return errors.New(fmt.Sprintf("failed to delete `%s`", sym))
	} else if result.RowsAffected < 1 {
		return errors.New(fmt.Sprintf("deleted more than 1 row `%s`", sym))
	}

	delete(server.watchlist, strings.ToLower(sym))

	return nil
}

func (server Server) Run() error {
	server.app.Get("api/watch", server.Index)
	server.app.Post("api/watch", server.Save)
	server.app.Put("api/watch", server.Save)
	server.app.Delete("api/watch", server.Delete)
	go func(s Server) {
		s.background()
	}(server)

	return server.app.Listen(":50100")
}

func handleError(ctx *fib.Ctx, err error) error {
	ctx.Status(500)
	ctx.WriteString(err.Error())
	return err
}

func (server Server) background() {
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
		server.log.Infof("Skipping today's watch as %s was a market holiday.", lastHoliday.Date.String())
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
