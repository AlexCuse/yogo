package main

import (
	"context"
	"github.com/alexcuse/yogo/internal/pkg/configuration"
	"github.com/alexcuse/yogo/internal/pkg/logging"

	"github.com/alexcuse/yogo/internal/scanner"
)

func main() {
	ctx := context.Background()
	log := logging.Bootstrap()

	errHandler := func(err error) {
		if err != nil {
			log.Fatal().Err(err).Msg("panic")
			panic(err)
		}
	}
	cfg := &scanner.Configuration{}
	errHandler(configuration.Unmarshal(cfg))

	server, err := scanner.NewServer(cfg, ctx, log)
	if err != nil {
		log.Error().Err(err).Msgf("Failed to connect to broker: %s", cfg.BrokerURL)
	}
	errHandler(err)

	errHandler(server.Run())
}
