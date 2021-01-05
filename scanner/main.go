package main

import (
	"context"
	"fmt"
	iex "github.com/goinvest/iexcloud/v2"
	"github.com/alexcuse/yogo/scanner/config"
	"github.com/alexcuse/yogo/scanner/signals"
	"os"
)
func main() {
	tickers := os.Args[1:]

	cfg, err := config.Load("configuration.toml")

	if err != nil {
		panic(err)
	}

	sig, err := signals.Load("signals.toml")

	if err != nil {
		panic(err)
	}

	ct := iex.NewClient(cfg.Token, iex.WithBaseURL(cfg.BaseURL))

	q, err := iex.Client.BatchQuote(*ct, context.Background(), tickers)

	if err != nil {
		panic(err)
	}

	for ticker, quote := range q {
		fmt.Printf("----- %s -----\n", ticker)
		for _, s := range sig {
			if s.Check(quote){
				fmt.Printf("%s matched\n", s.Name)
			}
		}

		fmt.Printf("%+v\n", quote)
		fmt.Printf("----- END %s -----\n\n", ticker)
	}
}