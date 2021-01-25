package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/alexcuse/yogo/social-enricher"
	"github.com/alexcuse/yogo/social-enricher/stocktwits"
	"github.com/go-resty/resty/v2"
	fib "github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/spf13/viper"
)

type config struct {
	DSN           string
	Schedule      string
	WatchAPI      string
	StocktwitsAPI string
}

func main() {
	log := zerolog.New(os.Stdout).With().Logger()
	ctx := log.WithContext(context.Background())

	errHandler := func(err error) {
		if err != nil {
			log.Fatal().Err(err).Msg("panic")
			panic(err)
		}
	}

	cfg := &config{}
	viper.SetEnvPrefix("yogo")
	viper.AutomaticEnv()
	viper.AddConfigPath(".")
	viper.SetConfigName("configuration")
	err := viper.ReadInConfig()
	errHandler(err)
	errHandler(viper.Unmarshal(cfg))

	symbols := make(chan string)
	defer close(symbols)

	sentiments := make(chan *social.Sentiment)
	defer close(sentiments)

	f := fib.New()
	f.Use(cors.New())

	wl := social.NewWatchList(resty.New().SetHostURL(cfg.WatchAPI))

	{
		twits := make(chan *stocktwits.Twits)
		defer close(twits)
		symbolApi := stocktwits.NewSymbolApi(resty.New().SetHostURL(cfg.StocktwitsAPI))
		go symbolApi.Stream(ctx, symbols, twits)

		sentimenter := stocktwits.NewSentimentCalculator()
		go sentimenter.Stream(ctx, twits, sentiments)
	}

	{
		calc := social.NewCalculator(&wl, symbols)

		crn := cron.New()
		_, err = crn.AddFunc("30 04 * * 2,3,4,5,6", func() {
			log := zerolog.Ctx(ctx).With().Time("cron", time.Now()).Logger()
			ctx := log.WithContext(ctx)

			err := calc.Start(ctx)
			if err == nil {
				log.Err(err).Msg("failed to start calc")
			}
		})
		errHandler(err)
		go crn.Run()

		calc.Start(ctx)
	}

	termChan := make(chan os.Signal, 10)
	signal.Notify(termChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log := zerolog.Ctx(ctx).With().Logger()
		for {
			select {
			case <-ctx.Done():
				return
			case s, ok := <-sentiments:
				if !ok {
					return
				}

				log.Info().Interface("sentiment", s).Msg("")

			}
		}
	}()

	<-termChan

	ctx.Done()
}
