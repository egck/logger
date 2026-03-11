package logger

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"sync"
	"time"
)

type Log struct {
	// Data part
	message    string
	attributes []any

	printHandler            func(msg string, args ...any) // Handler to print log (debug, info, ...)
	occurrenceNumber        int // Number of times the log has been detected
	registerDatetime string // Datetime of the log registering
}

type Limiter struct {
	firstN      int // Number of occurrences that will be printed for the same log
	flushPeriod int // Flush period of known logs

	context   context.Context // Context to properly stop the limiter goroutine
	waitGroup *sync.WaitGroup // waitGroup to sync with the main process
}

type Logger struct {
	logs map[string]*Log // Logs in memory
	lock sync.Mutex

	// Options
	logLevel *slog.LevelVar
	slog     *slog.Logger
	limiter  *Limiter
}

type Option func(*Logger)

func (logger *Logger) run() {
	ticker := time.NewTicker(time.Duration(logger.limiter.flushPeriod) * time.Second)
	for {
		select {
		case <-ticker.C:
			logger.flush()
		case <-logger.limiter.context.Done():
			logger.Info("stopped successfully", "component", "logger")
			logger.limiter.waitGroup.Done()
			return
		}
	}
}

func (logger *Logger) flush() {
	// Guarantee the access to in memory logs
	logger.lock.Lock()
	defer logger.lock.Unlock()

	// format every log in memory
	for _, log := range logger.logs {

		// generate slog attributes
		attributes := append(
			log.attributes,
			slog.Int("occurrences", log.occurrenceNumber),
			slog.String("since", log.registerDatetime),
		)

		// Print the summary of the log
		log.printHandler(log.message, attributes...)
	}

	// reset memory
	logger.logs = map[string]*Log{}
}

func (logger *Logger) register(printHandler func(msg string, args ...any), message string, args ...any) *Log {
	// Register the log in the memory
	if _, exists := logger.logs[message]; !exists {
		logger.logs[message] = &Log{
			message:                 message,
			printHandler:            printHandler,
			registerDatetime: time.Now().Format("2006-01-02T15:04:05.999999999Z07:00"),
			occurrenceNumber:        0,
			attributes:              args,
		}
	}

	// Increment occurrence number
	logger.logs[message].occurrenceNumber++

	return logger.logs[message]
}

func (logger *Logger) print(log *Log) {
	if logger.limiter == nil || log.occurrenceNumber <= logger.limiter.firstN {
		log.printHandler(log.message, log.attributes...)
	}
}

// Public functions

func NewLogger(opts ...Option) (*Logger, error) {
	// Init logger
	logger := &Logger{
		logLevel: &slog.LevelVar{},
		logs:     map[string]*Log{},
	}

	// Apply options
	for _, opt := range opts {
		opt(logger)
	}

	if logger.slog == nil {
		return nil, errors.New("logger must be init with a text or json handler")
	}

	// Run limiter goroutine if needed
	if logger.limiter != nil {
		logger.limiter.waitGroup.Add(1)
		go logger.run()
	}

	return logger, nil
}

func WithLimiter(ctx context.Context, waitGroup *sync.WaitGroup, firstN int, flushPeriod int) Option {
	return func(logger *Logger) {
		logger.limiter = &Limiter{
			firstN:      firstN,
			flushPeriod: flushPeriod,
			context:     ctx,
			waitGroup:   waitGroup,
		}
	}
}

func WithLevel(level slog.Level) Option {
	return func(logger *Logger) {
		logger.logLevel.Set(level)
	}
}

func WithJsonHandler() Option {
	return func(logger *Logger) {
		logger.slog = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logger.logLevel}))
	}
}

func WithTextHandler() Option {
	return func(logger *Logger) {
		logger.slog = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: logger.logLevel}))
	}
}

func (logger *Logger) SetLevel(level slog.Level) {
	logger.logLevel.Set(level)
}

func (logger *Logger) Debug(message string, args ...any) {
	if logger.logLevel.Level() <= slog.LevelDebug {
		logger.lock.Lock()
		defer logger.lock.Unlock()

		log := logger.register(logger.slog.Debug, message, args...)
		logger.print(log)
	}
}

func (logger *Logger) Info(message string, args ...any) {
	if logger.logLevel.Level() <= slog.LevelInfo {
		logger.lock.Lock()
		defer logger.lock.Unlock()

		log := logger.register(logger.slog.Info, message, args...)
		logger.print(log)
	}
}

func (logger *Logger) Warn(message string, args ...any) {
	if logger.logLevel.Level() <= slog.LevelWarn {
		logger.lock.Lock()
		defer logger.lock.Unlock()

		log := logger.register(logger.slog.Warn, message, args...)
		logger.print(log)
	}
}

func (logger *Logger) Error(message string, args ...any) {
	if logger.logLevel.Level() <= slog.LevelError {
		logger.lock.Lock()
		defer logger.lock.Unlock()

		log := logger.register(logger.slog.Error, message, args...)
		logger.print(log)
	}
}
