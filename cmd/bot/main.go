package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	logger, _ := zap.NewProduction(
		zap.IncreaseLevel(zapcore.DebugLevel),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	defer func() { _ = logger.Sync() }()

	if err := runBot(ctx, mts, logger.Named("bot")); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%+v\n", err)
		os.Exit(2)
	}
}
