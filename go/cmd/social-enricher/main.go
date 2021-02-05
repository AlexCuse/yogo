package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexcuse/yogo/internal/pkg/configuration"
	"github.com/alexcuse/yogo/internal/pkg/logging"

	"github.com/ThreeDotsLabs/watermill-kafka/v2/pkg/kafka"
	"github.com/alexcuse/yogo/internal/social"
	"github.com/alexcuse/yogo/internal/social/stocktwits"
	"github.com/alexdrl/zerowater"
	"github.com/go-resty/resty/v2"
	fib "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	gokv_syncmap "github.com/philippgille/gokv/syncmap"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
)

type config struct {
	DSN                string
	SocialCronSchedule string
	WatchAPI           string
	StocktwitsAPI      string
	BrokerURL          string
	QuoteTopic         string
	SocialTopic        string
	SentimentTopic     string
}

func main() {
	log := logging.Bootstrap()
	ctx := log.WithContext(context.Background())

	errHandler := func(err error) {
		if err != nil {
			log.Fatal().Err(err).Msg("panic")
			panic(err)
		}
	}

	cfg := &config{}
	errHandler(configuration.Unmarshal(cfg))

	wml := zerowater.NewZerologLoggerAdapter(log)
	sub, err := kafka.NewSubscriber(kafka.SubscriberConfig{
		Brokers:               []string{cfg.BrokerURL},
		Unmarshaler:           kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: kafka.DefaultSaramaSubscriberConfig(),
	}, wml)
	errHandler(err)

	pub, err := kafka.NewPublisher(kafka.PublisherConfig{
		Brokers:               []string{cfg.BrokerURL},
		Marshaler:             kafka.DefaultMarshaler{},
		OverwriteSaramaConfig: kafka.DefaultSaramaSyncPublisherConfig(),
	}, wml)
	errHandler(err)

	kvStore := gokv_syncmap.NewStore(gokv_syncmap.DefaultOptions)

	symbols := make(chan string, 100)
	defer close(symbols)

	sentimentSnapshots := make(chan social.SentimentSnapshot, 100)
	defer close(sentimentSnapshots)

	f := fib.New()
	f.Use(cors.New())

	wl := social.NewWatchList(resty.New().SetHostURL(cfg.WatchAPI))

	{
		twits := make(chan stocktwits.Twits, 100)
		defer close(twits)
		symbolApi := stocktwits.NewSymbolApi(resty.New().SetHostURL(cfg.StocktwitsAPI))
		go symbolApi.Stream(ctx, symbols, twits)

		sentimenter := stocktwits.NewSentimentCalculator()
		go sentimenter.Stream(ctx, twits, sentimentSnapshots)
	}

	{
		argSentimentSnapshots := make(chan social.SentimentSnapshot, 100)
		histSentimentSnapshots := make(chan social.SentimentSnapshot, 100)
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				case s, ok := <-sentimentSnapshots:
					if !ok {
						return
					}

					argSentimentSnapshots <- s
					histSentimentSnapshots <- s
				}
			}
		}()
		t, err := social.NewDailySentimentAggregator(kvStore)
		errHandler(err)
		go t.Stream(ctx, argSentimentSnapshots)

		enricher := social.NewEnricher(&t, pub)
		input, err := sub.Subscribe(context.Background(), cfg.QuoteTopic)
		errHandler(err)
		go enricher.Execute(ctx, input, cfg.SocialTopic)

		historian := social.NewSentimentHistorian(pub)
		go historian.Stream(ctx, histSentimentSnapshots, cfg.SentimentTopic)
	}

	{
		calc := social.NewCalculator(&wl, symbols)

		crn := cron.New()
		_, err = crn.AddFunc(cfg.SocialCronSchedule, func() {
			log := zerolog.Ctx(ctx).With().Time("cron", time.Now()).Logger()
			ctx := log.WithContext(ctx)

			err := calc.Start(ctx)
			if err != nil {
				log.Err(err).Msg("failed to start calc")
			}
		})
		errHandler(err)
		go crn.Run()

		err := calc.Start(ctx)
		errHandler(err)
	}

	termChan := make(chan os.Signal, 10)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	<-termChan

	ctx.Done()
}
