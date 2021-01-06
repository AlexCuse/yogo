package main

import (
	"context"
	"encoding/json"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/alexcuse/yogo/common/config"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/google/uuid"
	"github.com/hashicorp/go-multierror"
	"log"
	"net/http"
	"os"
)

type input struct {
	Tickers []string
}

func main() {
	log := log.New(os.Stdout, "quoter: ", log.LstdFlags)

	cfg, err := config.Load("configuration.toml")

	if err != nil {
		panic(err)
	}

	ct := iex.NewClient(cfg.IEXToken, iex.WithBaseURL(cfg.IEXBaseURL))

	pub, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:               []string{cfg.BrokerURL},
		Marshaler:             kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: kafka.DefaultSaramaSyncPublisherConfig(),
	}, &watermill.StdLoggerAdapter{ErrorLogger: log, InfoLogger: log})

	if err != nil {
		panic(err)
	}

	err = http.ListenAndServe(":50100", listener{
		log:        log,
		iexClient:  ct,
		publisher:  pub,
		quoteTopic: cfg.QuoteTopic,
	})

	panic(err)
}

type listener struct {
	quoteTopic string
	log        *log.Logger
	iexClient  *iex.Client
	publisher  message.Publisher
}

func (l listener) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	decoder := json.NewDecoder(request.Body)

	ipt := input{}

	err := decoder.Decode(&ipt)

	if err != nil {
		writer.WriteHeader(500)
		writer.Write([]byte(err.Error()))
		return
	}

	processingErrs := &multierror.Error{}

	for _, t := range ipt.Tickers {
		q, err := l.iexClient.PreviousDay(context.Background(), t)

		if err != nil {
			processingErrs = multierror.Append(processingErrs, err)
			continue
		}

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
		writer.WriteHeader(500)
		writer.Write([]byte(finalError.Error()))
		return
	}

	writer.WriteHeader(200)
	writer.Write([]byte("OK"))
}
