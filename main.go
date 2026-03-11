package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"
	"time"

	"github.com/egck/glogger/logger"
)

func main() {
	// Usage of the logger with a limiter
	usageWithLimiter()

	// Usage of the logger without a limiter
	usageWithoutLimiter()

	// Usage of the logger changing dynamically its level
	dynamicLevelUpdate()
}

func usageWithLimiter() {
	// Init vars
	ctx, cancel := context.WithCancel(context.Background())
	waitGroup := &sync.WaitGroup{}
	flushPeriod := 5
	firstN := 2

	// Init logger
	customLogger, err := logger.NewLogger(
		logger.WithJsonHandler(),
		logger.WithLevel(slog.LevelInfo),
		logger.WithLimiter(ctx, waitGroup, firstN, flushPeriod),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Simulate a log overflow
	for i := 0; i < 200; i++ {
		customLogger.Error("my error", "component", "main")                                                      // will be printed and register
		customLogger.Debug("my debug", "component", "main", "details", "these are useful details for debugging") // won't be either registered or printed
	}

	time.Sleep(time.Duration(flushPeriod+1) * time.Second)

	cancel()
	waitGroup.Wait()
}

func usageWithoutLimiter() {
	// Init logger
	customLogger, err := logger.NewLogger(
		logger.WithJsonHandler(),
		logger.WithLevel(slog.LevelInfo),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Error log printed
	customLogger.Error("my error", "component", "main")

	// Debug log not printed
	customLogger.Debug("my debug", "component", "main", "details", "these are useful details for debugging")
}

func dynamicLevelUpdate() {
	// Init logger
	customLogger, err := logger.NewLogger(
		logger.WithJsonHandler(),
		logger.WithLevel(slog.LevelInfo),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	// Error log printed
	customLogger.Error("my error", "component", "main")

	// Debug log not printed
	customLogger.Debug("my debug", "component", "main", "details", "these are useful details for debugging")
	
	customLogger.SetLevel(slog.LevelDebug)
	// Debug log printed
	customLogger.Debug("my debug", "component", "main", "details", "these are useful details for debugging")

}
