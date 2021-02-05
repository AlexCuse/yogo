package main

import (
	"context"

	"github.com/alexcuse/yogo/internal/monitor"
	"github.com/alexcuse/yogo/internal/pkg/configuration"
	"github.com/alexcuse/yogo/internal/pkg/logging"
)

func main() {
	log := logging.Bootstrap()
	errHandler := func(err error) {
		if err != nil {
			log.Fatal().Err(err).Msg("panic")
			panic(err)
		}
	}

	cfg := &monitor.Configuration{}
	errHandler(configuration.Unmarshal(cfg))

	ctx := context.Background()

	server := monitor.NewServer(cfg, ctx, log)

	errHandler(server.Run())

}
