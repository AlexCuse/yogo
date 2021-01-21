package main

import (
	"context"
	"github.com/alexcuse/yogo/common"
	"github.com/alexcuse/yogo/watch/server"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, log, _ := common.Bootstrap("configuration.toml")

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	s := server.NewServer(cfg, context.Background(), db, log)

	log.Fatal(s.Run())
}
