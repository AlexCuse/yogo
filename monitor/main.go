package main

import (
	"context"
	"github.com/alexcuse/yogo/common"
	"github.com/alexcuse/yogo/monitor/server"
	"os"
)

func main() {
	cfg, log, wml := common.Bootstrap("configuration.toml")

	ctx := context.Background()

	server := server.NewServer(cfg, ctx, log, wml)

	server.Run()

	select {
	case <-ctx.Done():
		os.Exit(0)
	}
}
