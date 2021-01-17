package common

import (
	"github.com/ThreeDotsLabs/watermill"
	"github.com/alexcuse/yogo/common/config"
	"github.com/sirupsen/logrus"
	"os"
)

func Bootstrap(configFile string) (config.Configuration, *logrus.Logger, watermill.LoggerAdapter) {
	cfg, err := config.Load(configFile)
	if err != nil {
		panic(err)
	}

	log := logrus.New()
	log.SetOutput(os.Stdout)

	if cfg.LogLevel != "" {
		if level, err := logrus.ParseLevel(cfg.LogLevel); err == nil {
			log.SetLevel(level)
		}
	}

	wml := watermill.NewStdLoggerWithOut(log.Out, true, false)

	return cfg, log, wml
}
