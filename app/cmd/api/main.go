package main

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func main() {
	logger := zerolog.New(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: time.RFC3339}).
		With().Timestamp().Caller().Logger()

	server, err := NewServer()
	if err != nil {
		logger.Fatal().Msg(err.Error())
	}
	err = server.Run()
	logger.Fatal().Msg(err.Error())
}
