package main

import (
	"context"
	"fmt"
	iex "github.com/goinvest/iexcloud/v2"
	config "github.com/alexcuse/yogo/scanner/config"
	"os"
)
func main() {
	tickers := os.Args[1:]

	cfg, err := config.Load("configuration.toml")

	if err != nil {
		panic(err)
	}

	ct := iex.NewClient(cfg.Token)

	q, err := iex.Client.BatchQuote(*ct, context.Background(), tickers)

	if err != nil {
		panic(err)
	}

	for ticker, quote := range q {
		fmt.Println(fmt.Sprintf("%s: %+v\n", ticker, quote))
	}
}