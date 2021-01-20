package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/alexcuse/yogo/common/config"
	fib "github.com/gofiber/fiber/v2"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Watch struct {
	Symbol string `gorm:"primaryKey;autoIncrement:false"`
}

type Server struct {
	db     *gorm.DB
	log    *logrus.Logger
	appctx context.Context
	app    *fib.App
	cfg    config.Configuration
	iex    *iex.Client
}

func NewServer(cfg config.Configuration, appctx context.Context, db *gorm.DB, log *logrus.Logger) Server {
	f := fib.New()

	iecli := iex.NewClient(cfg.IEXToken, iex.WithBaseURL(cfg.IEXBaseURL))

	server := Server{
		db:     db,
		log:    log,
		appctx: appctx,
		app:    f,
		cfg:    cfg,
		iex:    iecli,
	}

	return server
}

func (server Server) List(ctx *fib.Ctx) error {
	watching := make([]Watch, 0)

	_ = server.db.Find(&watching)

	jsn, _ := json.Marshal(watching)

	ctx.Write(jsn)

	return nil
}

func (server Server) Save(ctx *fib.Ctx) error {
	w := Watch{}

	err := json.Unmarshal(ctx.Body(), &w)

	if err != nil {
		return handleError(ctx, err)
	}

	if r := server.db.WithContext(server.appctx).Clauses(clause.OnConflict{DoNothing: true}).Create(w); r.Error != nil {
		server.log.Errorf("unable to persist server: %s", r.Error.Error())
	}

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

	return nil
}

func (server Server) Run() error {
	server.app.Post("api/watch", server.Save)
	server.app.Put("api/watch", server.Save)
	server.app.Delete("api/watch", server.Delete)
	server.app.Get("api/watch", server.List)
	return server.app.Listen(fmt.Sprintf(":%d", server.cfg.WatchPort))
}

func handleError(ctx *fib.Ctx, err error) error {
	ctx.Status(500)
	ctx.WriteString(err.Error())
	return err
}
