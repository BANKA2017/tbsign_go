package _function

import (
	"log/slog"
	"os"

	"github.com/BANKA2017/tbsign_go/share"
	"github.com/kdnetwork/code-snippet/go/log"
)

var SlogLevel slog.LevelVar

func InitDefaultLogger() {
	SlogLevel.Set(slog.LevelInfo) // slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     &SlogLevel,
		AddSource: true,
	})

	logger := log.InitLoggerPresets(slog.New(handler).With("version", share.DynamicVersion))
	slog.SetDefault(logger)
}
