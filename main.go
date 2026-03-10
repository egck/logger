package main

import (
	// "context"
	"context"
	"fmt"
	"log/slog"
	"os"
	"sync"

	// "sync"
	"time"

	"github.com/egck/glogger/logger"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	waitGroup := &sync.WaitGroup{}
	customLogger, err := logger.NewLogger(
		logger.WithJsonHandler(),
		logger.WithLevel(slog.LevelDebug),
		logger.WithLimiter(ctx, waitGroup, 2, 5),
	)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i := 0; i < 20; i++ {
		customLogger.Info("my error", "component", "main")
		if i%3 == 0 {
			customLogger.SetLevel(slog.LevelInfo)
			time.Sleep(1 * time.Second)
		}
	}

	cancel()
	waitGroup.Wait()
}
