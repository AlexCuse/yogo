package signals

import (
	"context"
	"fmt"

	fib "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

type Configuration struct {
	SignalPort int
	DSN        string
}

type Server struct {
	db     *gorm.DB
	log    zerolog.Logger
	appctx context.Context
	app    *fib.App
	cfg    *Configuration
	sig    SignalHandler
}

func NewServer(cfg *Configuration, appctx context.Context, db *gorm.DB, log zerolog.Logger) Server {
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
	server.app.Get("api/signals", server.sig.List)
	server.app.Get("api/signals/current", server.sig.Current)
	server.app.Get("api/signal/:name/current", server.sig.CurrentByName)
	return server.app.Listen(fmt.Sprintf(":%d", server.cfg.SignalPort))
}

func (server Server) handleError(ctx *fib.Ctx, err error) error {
	ctx.Status(500)
	_, _ = ctx.WriteString(err.Error())
	return err
}
