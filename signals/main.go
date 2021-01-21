package main

import (
	"context"
	"github.com/alexcuse/yogo/common"
	"github.com/alexcuse/yogo/signals/server"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, log, wml := common.Bootstrap("configuration.toml")

	db, err := gorm.Open(postgres.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = db.AutoMigrate(&server.Signal{})

	if err != nil {
		panic(err)
	}

	ctx := context.Background()

	server := server.NewServer(cfg, ctx, db, log, wml)

	log.Fatal(server.Run())
}
