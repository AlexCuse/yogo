package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/alexcuse/yogo/common/config"
	fib "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Signal struct {
	Name   string `gorm:"primaryKey;autoIncrement:false" json:"name,omitempty"`
	Source string `json:"source,omitempty"`
}

type Server struct {
	db     *gorm.DB
	log    *logrus.Logger
	appctx context.Context
	app    *fib.App
	cfg    config.Configuration
}

func NewServer(cfg config.Configuration, appctx context.Context, db *gorm.DB, log *logrus.Logger, wml watermill.LoggerAdapter) Server {
	f := fib.New()
	f.Use(cors.New())

	server := Server{
		db:     db,
		log:    log,
		appctx: appctx,
		app:    f,
		cfg:    cfg,
	}

	return server
}

func (server Server) Run() error {
	server.app.Post("api/signal", server.Save)
	server.app.Put("api/signal", server.Save)
	server.app.Delete("api/signal", server.Delete)
	server.app.Get("api/signal", server.List)
	return server.app.Listen(fmt.Sprintf(":%d", server.cfg.SignalPort))
}

func handleError(ctx *fib.Ctx, err error) error {
	ctx.Status(500)
	ctx.WriteString(err.Error())
	return err
}

func (server Server) List(ctx *fib.Ctx) error {
	signals := make([]Signal, 0)

	result := server.db.Find(&signals)

	if result.Error != nil {
		return handleError(ctx, result.Error)
	}

	if result.RowsAffected == 0 {
		ctx.Status(404)
		return nil
	}

	return ctx.JSON(signals)
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

	return nil
}
