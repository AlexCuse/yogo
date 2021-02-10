package logging

import (
	"os"

	"github.com/rs/zerolog"
)

func Bootstrap() zerolog.Logger {
	return zerolog.New(os.Stdout).With().Logger()
}
