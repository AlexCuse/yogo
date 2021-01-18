package main

import (
	"context"
	"github.com/alexcuse/yogo/common"
	"github.com/alexcuse/yogo/monitor/watch"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, log, wml := common.Bootstrap("configuration.toml")

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&watch.Watch{})

	if err != nil {
		panic(err)
	}

	server := watch.NewServer(cfg, context.Background(), db, log, wml)

	log.Fatal(server.Run())
}
