package signals

import (
	"encoding/json"
	"fmt"

	"github.com/alexcuse/yogo/internal/pkg/db"
	fib "github.com/gofiber/fiber/v2"
	iex "github.com/goinvest/iexcloud/v2"
	"gorm.io/gorm/clause"
)

type SignalHandler interface {
	List(ctx *fib.Ctx) error
	Delete(ctx *fib.Ctx) error
	Save(ctx *fib.Ctx) error
	Current(ctx *fib.Ctx) error
	CurrentByName(ctx *fib.Ctx) error
}

type signalHandler struct {
	Server
}

func NewSignalHandler(server Server) SignalHandler {
	return signalHandler{server}
}

func (h signalHandler) List(ctx *fib.Ctx) error {
	signals := make([]Signal, 0)

	result := h.db.Find(&signals)

	if result.Error != nil {
		return h.handleError(ctx, result.Error)
	}

	return ctx.JSON(signals)
}

func (h signalHandler) Save(ctx *fib.Ctx) error {
	s := Signal{}

	err := json.Unmarshal(ctx.Body(), &s)

	if err != nil {
		return h.handleError(ctx, err)
	}

	if r := h.db.WithContext(h.appctx).Clauses(clause.OnConflict{UpdateAll: true}).Create(s); r.Error != nil {
		h.log.Error().Err(err).Msg("unable to persist signal")
	}

	if err != nil {
		return h.handleError(ctx, err)
	}

	if err != nil {
		return h.handleError(ctx, err)
	}

	return nil
}

func (h signalHandler) Delete(ctx *fib.Ctx) error {
	name := ctx.Query("name")

	if name == "" {
		return nil
	}

	result := h.db.WithContext(h.appctx).Delete(&Signal{Name: name})

	if result.RowsAffected == 0 {
		return fmt.Errorf("failed to delete `%s`", name)
	} else if result.RowsAffected < 1 {
		return fmt.Errorf(fmt.Sprintf("deleted more than 1 row `%s`", name))
	}

	return nil
}

func (h signalHandler) Current(ctx *fib.Ctx) error {
	signals := make([]SignalWithHitCount, 0)

	result := h.db.Select("signals.*, count(hits.*) as Count").Table(
		"hits",
	).Joins(
		"left join signals on hits.rule_name = signals.name",
	).Group(
		"signals.name, signals.source",
	).Where(
		"quote_date = (?)",
		h.db.Select("Max(quote_date)").Table("hits"),
	).Scan(&signals)

	if result.Error != nil {
		return h.handleError(ctx, result.Error)
	}

	if result.RowsAffected == 0 {
		ctx.Status(404)
		return nil
	}

	return ctx.JSON(signals)
}

func (h signalHandler) CurrentByName(ctx *fib.Ctx) error {
	name := ctx.Params("name")

	res := Signal{}

	lastMarketDay := struct {
		Date iex.Date
	}{}

	dateResult := h.db.Select(`max("date") "date"`).Table("movements").Scan(&lastMarketDay)

	if dateResult.Error != nil {
		return h.handleError(ctx, dateResult.Error)
	}

	signalResult := h.db.Select("signals.*, hits.symbol").Table(
		"hits",
	).Joins(
		"left join signals on hits.rule_name = signals.name",
	).Where(
		"hits.quote_date = ? and signals.name = ?",
		lastMarketDay.Date,
		name,
	).Scan(&res)

	if signalResult.Error != nil {
		return h.handleError(ctx, signalResult.Error)
	}

	movements := make([]db.Movement, 0)

	movementResult := h.db.Select("movements.*").Table(
		"movements",
	).Joins(`inner join hits on hits.symbol = movements.symbol and hits.quote_date = movements."date"`).Where(
		`hits.quote_date = ? and hits.rule_name = ?`,
		lastMarketDay.Date,
		name,
	).Order(
		"movements.symbol",
	).Scan(&movements)

	if movementResult.Error != nil {
		return h.handleError(ctx, movementResult.Error)
	}

	stats := make([]db.Stats, 0)

	statsResult := h.db.Select("*").Table(
		"stats",
	).Joins("inner join hits on hits.symbol = stats.symbol and hits.quote_date = stats.quote_date").Where(
		`hits.quote_date = ? and hits.rule_name = ?`,
		lastMarketDay.Date,
		name,
	).Order(
		"stats.symbol",
	).Scan(&stats)

	if statsResult.Error != nil {
		return h.handleError(ctx, statsResult.Error)
	}

	output := SignalResult{res, []SignalHit{}}

	for _, mvmt := range movements {
		for _, stats := range stats {
			if mvmt.Symbol == stats.Symbol {
				_ = output.AddHit(mvmt, stats)
			}
		}
	}

	return ctx.JSON(output)
}
