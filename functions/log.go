package _function

import (
	"fmt"
	"log/slog"
	"os"
)

var SlogLevel = slog.LevelError // slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError

func InitLogger() {
	slog.SetDefault(slog.New(
		slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     SlogLevel,
			AddSource: true,
		}),
	))
}

func Fatal(msg string, v ...any) {
	slog.Error(msg, v...)
	os.Exit(1)
}

func FmtFatal(v ...any) {
	fmt.Println(v...)
	os.Exit(1)
}
