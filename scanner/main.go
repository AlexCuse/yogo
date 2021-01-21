package main

import (
	"context"
	"github.com/alexcuse/yogo/common"
	"github.com/alexcuse/yogo/scanner/server"
)

func main() {
	cfg, log, wml := common.Bootstrap("configuration.toml")

	ctx := context.Background()

	server := server.NewServer(cfg, ctx, log, wml)

	log.Fatal(server.Run())
}
