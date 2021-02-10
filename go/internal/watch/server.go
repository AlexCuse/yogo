package watch

import (
	"context"
	"encoding/json"
	"fmt"

	fib "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Configuration struct {
	IEXToken   string
	IEXBaseURL string
	WatchPort  int
	DSN        string
}

type Watch struct {
	Symbol string `gorm:"primaryKey;autoIncrement:false" json:"symbol,omitempty"`
}

type Server struct {
	db     *gorm.DB
	log    zerolog.Logger
	appctx context.Context
	app    *fib.App
	cfg    *Configuration
	iex    *iex.Client
}

func NewServer(cfg *Configuration, appctx context.Context, db *gorm.DB, log zerolog.Logger) Server {
	f := fib.New()
	f.Use(cors.New())

	iecli := iex.NewClient(cfg.IEXToken, iex.WithBaseURL(cfg.IEXBaseURL))

	db.AutoMigrate(&Watch{})

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
		server.log.Error().Err(err).Msg("unable to persist server")
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
		return fmt.Errorf("failed to delete `%s`", sym)
	} else if result.RowsAffected < 1 {
		return fmt.Errorf("deleted more than 1 row `%s`", sym)
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
	_, _ = ctx.WriteString(err.Error())
	return err
}
