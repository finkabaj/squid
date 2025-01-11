package logger

import (
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

	if os.Getenv("ENV") == "development" {
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
