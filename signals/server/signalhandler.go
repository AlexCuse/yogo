package server

import (
	"encoding/json"
	"errors"
	"fmt"
	fib "github.com/gofiber/fiber/v2"
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
		h.log.Errorf("unable to persist signal: %s", r.Error.Error())
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
		return errors.New(fmt.Sprintf("failed to delete `%s`", name))
	} else if result.RowsAffected < 1 {
		return errors.New(fmt.Sprintf("deleted more than 1 row `%s`", name))
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

	result := h.db.Select("signals.*, hits.symbol").Table(
		"hits",
	).Joins(
		"left join signals on hits.rule_name = signals.name",
	).Where(
		"hits.quote_date = (?) and signals.name = ?",
		h.db.Select("Max(quote_date)").Table("hits"),
		name,
	).Scan(&res)

	if result.Error != nil {
		return h.handleError(ctx, result.Error)
	}

	tickers := make([]string, 0)

	tickerResult := h.db.Select("distinct symbol").Table(
		"hits",
	).Where(
		"quote_date = (?) and rule_name = ?",
		h.db.Select("Max(quote_date)").Table("hits"),
		name,
	).Scan(&tickers)

	if tickerResult.Error != nil {
		return h.handleError(ctx, result.Error)
	}

	return ctx.JSON(SignalDetail{res, tickers})
}
