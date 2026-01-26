package logging

import (
	"log/slog"
	"os"
	"runtime"

	"github.com/Shoowa/vamos/config"
)

func configure(cfg *config.Config) *slog.HandlerOptions {
	logLevel := &slog.LevelVar{}
	if cfg.Logger.Debug == true {
		logLevel.Set(slog.LevelDebug)
	} else {
		logLevel.Set(slog.LevelWarn)
	}

	opts := &slog.HandlerOptions{Level: logLevel}
	return opts
}

// CreateLogger provides a structured JSON logger configured with a few fields
// displaying the version of the programming language, and a version of the
// application. It emits either debug or warn data.
func CreateLogger(cfg *config.Config) *slog.Logger {
	goVersion := slog.String("lang", runtime.Version())
	appVersion := slog.String("app", config.AppVersion)
	group := slog.Group("version", goVersion, appVersion)

	opts := configure(cfg)
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler).With(group)
	slog.SetDefault(logger)

	logger.Info("Logger", "level", opts.Level.Level())
	return logger
}
