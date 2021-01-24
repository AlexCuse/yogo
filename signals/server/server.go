package server

import (
	"context"
	"fmt"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/alexcuse/yogo/common/config"
	fib "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
)

type Server struct {
	db     *gorm.DB
	log    *logrus.Logger
	appctx context.Context
	app    *fib.App
	cfg    config.Configuration
	sig    SignalHandler
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

	server.sig = NewSignalHandler(server)

	return server
}

func (server Server) Run() error {
	server.app.Post("api/signal", server.sig.Save)
	server.app.Put("api/signal", server.sig.Save)
	server.app.Delete("api/signal", server.sig.Delete)
	server.app.Get("api/signal", server.sig.List)
	server.app.Get("api/signal/current", server.sig.Current)
	server.app.Get("api/signal/currentbyname", server.sig.CurrentByName)
	return server.app.Listen(fmt.Sprintf(":%d", server.cfg.SignalPort))
}

func (server Server) handleError(ctx *fib.Ctx, err error) error {
	ctx.Status(500)
	ctx.WriteString(err.Error())
	return err
}
