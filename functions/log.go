package _function

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"

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

// // Telemetry
// var sentrySF singleflight.Group
//
// func InitSentryLogger(ctx context.Context, dsn string) error {
// 	// env.tc_go_optin_sentry_dsn string option
// 	// env.tc_go_optin_telemetry bool
//
// 	if !utils.GetBoolEnv("tc_go_optin_telemetry") && GetOption("go_optin_telemetry") != "1" {
// 		return errors.New("telemetry disallowed")
// 	}
//
// 	if dsn == "" {
// 		dsn = utils.GetEnv("tc_go_optin_sentry_dsn", "")
// 		if dsn == "" {
// 			return errors.New("sentry dsn is empty")
// 		}
// 	}
//
// 	_, err, _ := sentrySF.Do("sentry", func() (any, error) {
// 		err := sentry.Init(sentry.ClientOptions{
// 			Dsn:              dsn,
// 			Debug:            true,
// 			SendDefaultPII:   true,
// 			EnableLogs:       true,
// 			EnableTracing:    true,
// 			TracesSampleRate: 1.0,
// 		})
// 		if err != nil {
// 			slog.Error("sentry init error", "error", err.Error())
// 			return nil, err
// 		}
// 		// Flush buffered events before the program terminates.
// 		defer sentry.Flush(2 * time.Second)
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
//
// 		return nil, nil
// 	})
// 	return err
// }
