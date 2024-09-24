package main

import (
	"log/slog"
	"os"
	"strings"

	"github.com/dusted-go/logging/prettylog"
	slogformatter "github.com/samber/slog-formatter"
)

func InitLogging(lvl slog.Level) {
	logLvl := func() slog.Level {
		return lvl
	}()
	w := os.Stderr

	funcHandler := slogformatter.NewFormatterHandler(
		slogformatter.FormatByType(func(s []string) slog.Value {
			return slog.StringValue(strings.Join(s, ","))
		}),
	)

	plHandler := prettylog.New(
		&slog.HandlerOptions{
			Level:       logLvl,
			AddSource:   false,
			ReplaceAttr: nil,
		},
		prettylog.WithDestinationWriter(w),
	)

	formatHandler := funcHandler(plHandler)

	logger := slog.New(formatHandler)
	slog.SetDefault(logger)
}
