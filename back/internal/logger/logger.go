package logger

import (
	"github.com/finkabaj/squid/back/internal/config"
	"io"
	"os"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger
var loggerInitalized bool

func InitLogger(fs *os.File) {
	if loggerInitalized {
		return
	}

	var target io.Writer

	if config.Data.Env == "development" {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: "2006-01-02 15:04:05",
		}
		target = zerolog.MultiLevelWriter(consoleWriter, fs)
	} else {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		target = fs
	}

	Logger = zerolog.New(target).With().Timestamp().Logger()

	loggerInitalized = true
}
