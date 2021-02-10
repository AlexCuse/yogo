package main

import (
	"context"

	"github.com/alexcuse/yogo/internal/pkg/configuration"
	"github.com/alexcuse/yogo/internal/pkg/logging"
	"github.com/alexcuse/yogo/internal/signals"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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
	cfg := &signals.Configuration{}
	errHandler(configuration.Unmarshal(cfg))

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	errHandler(err)
	err = db.AutoMigrate(&signals.Signal{})
	errHandler(err)

	server := signals.NewServer(cfg, ctx, db, log)
	errHandler(server.Run())
}
