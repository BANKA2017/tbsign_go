package _function

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"sync/atomic"

	"github.com/BANKA2017/tbsign_go/share"
)

var SlogLevel = slog.LevelError // slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError

func InitDefaultLogger() {
	handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level:     SlogLevel,
		AddSource: true,
	})

	logger := InitLoggerPresets(slog.New(handler))
	slog.SetDefault(logger)
}

func InitLoggerPresets(logger *slog.Logger) *slog.Logger {
	return logger.With("version", share.DynamicVersion).
		With(
			slog.Group("system",
				slog.String("os", runtime.GOOS),
				slog.String("arch", runtime.GOARCH),
			),
		)
}

func Fatal(msg string, v ...any) {
	slog.Error(msg, v...)
	os.Exit(1)
}

func FmtFatal(v ...any) {
	fmt.Println(v...)
	os.Exit(1)
}

// Telemetry
var TelemetryActive atomic.Bool

// var sentrySF singleflight.Group
//
//
// How to init sentry logger
// os.Setenv("tc_go_optin_telemetry", "1")
// ctx, cancel := context.WithCancel(context.Background())
// defer cancel()
// go func() {
// 	if err := _function.InitSentryLogger(ctx, ""); err != nil {
// 		slog.Error("sentry", "error", err)
// 	}
// }()
//
// func InitSentryLogger(ctx context.Context, dsn string) error {
// 	// env.SENTRY_DSN string option // official name
// 	// env.tc_go_optin_telemetry bool
//
// 	if !(utils.GetBoolEnv("tc_go_optin_telemetry") || GetOption("go_optin_telemetry") == "1") {
// 		return errors.New("telemetry disallowed")
// 	}
//
// 	if dsn == "" {
// 		dsn = utils.GetEnv("SENTRY_DSN", dsn)
// 		if dsn == "" {
// 			return errors.New("sentry dsn is empty")
// 		}
// 	}
//
// 	_, err, _ := sentrySF.Do("sentry", func() (any, error) {
// 		TelemetryActive.Store(true)
// 		defer TelemetryActive.Store(false)
// 		err := sentry.Init(sentry.ClientOptions{
// 			Dsn:        dsn,
// 			Debug:      true,
// 			EnableLogs: true,
// 			Release:    "tbsign_go." + share.DynamicVersion,
// 			Dist:       share.BuildRuntime,
//
// 			HTTPClient: DefaultClient,
// 		})
// 		if err != nil {
// 			slog.Error("sentry init error", "error", err.Error())
// 			return nil, err
// 		}
// 		slog.Info("sentry init success")
//
// 		// Flush buffered events before the program terminates.
// 		// flush here is no use here because ctx was canceled
// 		// sentry.Flush(2 * time.Second)
//
// 		sentryLogLevels := []slog.Level{slog.LevelInfo, slog.LevelWarn, slog.LevelError, sentryslog.LevelFatal}
// 		if SlogLevel == slog.LevelDebug {
// 			sentryLogLevels = append(sentryLogLevels, slog.LevelDebug)
// 		}
//
// 		ctx, cancel := context.WithTimeout(ctx, time.Hour)
// 		defer cancel()
//
// 		handler := sentryslog.Option{
// 			LogLevel:  sentryLogLevels,
// 			AddSource: true,
// 		}.NewSentryHandler(ctx)
//
// 		slog.SetDefault(InitLoggerPresets(slog.New(handler)))
//
// 		<-ctx.Done()
// 		InitDefaultLogger()
//
// 		return nil, nil
// 	})
// 	return err
// }
