package logger

import (
	"log/slog"
	"os"
	"strings"
	"sync"
)

var loggerOnce = sync.Once{}

func InitLogger(level string) {
	loggerOnce.Do(func() {
		logger := slog.New(
			slog.NewJSONHandler(os.Stdout,
				&slog.HandlerOptions{
					Level: func() slog.Leveler {
						switch strings.ToLower(level) {
						case "debug":
							return slog.LevelDebug
						case "info":
							return slog.LevelInfo
						case "warn":
							return slog.LevelWarn
						case "error":
							return slog.LevelError
						default:
							return slog.LevelInfo
						}
					}(),
				},
			),
		)
		slog.SetDefault(logger)
	})
}
