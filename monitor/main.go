package main

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexcuse/yogo/common"
	"github.com/gofiber/fiber/v2"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"github.com/sirupsen/logrus"
)

type input struct {
	Tickers []string
}

func main() {
	cfg, log, wml := common.Bootstrap("configuration.toml")

	ct := iex.NewClient(cfg.IEXToken, iex.WithBaseURL(cfg.IEXBaseURL))

	pub, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:               []string{cfg.BrokerURL},
		Marshaler:             kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: kafka.DefaultSaramaSyncPublisherConfig(),
	}, wml)

	if err != nil {
		panic(err)
	}

	fib := fiber.New()

	l := listener{
		log:        log,
		iexClient:  ct,
		publisher:  pub,
		quoteTopic: cfg.QuoteTopic,
	}

	fib.Add("POST", "/previousday", l.HandlePreviousDay)

	log.Fatal(fib.Listen(":50100"))
}

type listener struct {
	quoteTopic string
	log        *logrus.Logger
	iexClient  *iex.Client
	publisher  message.Publisher
}

func (l listener) HandlePreviousDay(ctx *fiber.Ctx) error {
	ipt := input{}

	err := json.Unmarshal(ctx.Request().Body(), &ipt)

	if err != nil {
		ctx.Status(500)
		ctx.WriteString(err.Error())
		return err
	}

	processingErrs := &multierror.Error{}

	quotes := make([]iex.PreviousDay, 0)

	for _, t := range ipt.Tickers {
		q, err := l.iexClient.PreviousDay(context.Background(), t)

		if err != nil {
			processingErrs = multierror.Append(processingErrs, err)
			continue
		}

		quotes = append(quotes, q)

		jsn, err := json.Marshal(q)

		if err != nil {
			processingErrs = multierror.Append(processingErrs, err)
			continue
		}

		msg := message.NewMessage(uuid.New().String(), jsn)

		err = l.publisher.Publish(l.quoteTopic, msg)

		if err != nil {
			processingErrs = multierror.Append(processingErrs, err)
			continue
		}
	}

	finalError := processingErrs.ErrorOrNil()

	if finalError != nil {
		l.log.Error(finalError.Error())
		ctx.Status(500)
		ctx.WriteString(finalError.Error())
		return finalError
	}

	resp, _ := json.Marshal(quotes)

	ctx.Status(200)
	ctx.Write(resp)

	return nil
}
