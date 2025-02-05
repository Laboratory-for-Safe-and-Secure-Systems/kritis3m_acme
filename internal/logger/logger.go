package logger

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

const (
	CtxKeyLogger = iota
)

// Logger is a zerolog-based logger implementing the custom and standard log interfaces.
type Logger struct {
	logger zerolog.Logger
	fields []keyValue
}

// keyValue is a loosely-typed key-value pair.
type keyValue struct {
	key   string
	value interface{}
}

// GetLogger retrieves the logger from the context.
func GetLogger(ctx context.Context) *Logger {
	return ctx.Value(CtxKeyLogger).(*Logger)
}

// New creates a new zerolog-based logger which writes to the specified writer.
// It uses zerolog.ConsoleWriter for human-readable output (old style output with colors).
func New(w io.Writer) *Logger {
	// Use zerolog.ConsoleWriter for human-readable output with colors.
	consoleWriter := zerolog.ConsoleWriter{
		Out:        w,
		TimeFormat: time.RFC3339Nano,
		NoColor:    false, // Enable colors
	}

	// Initialize zerolog with the console writer.
	logger := zerolog.New(consoleWriter).With().Timestamp().Logger()

	// Set the global level to debug to include all log levels.
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	return &Logger{
		logger: logger,
	}
}

// Debug uses fmt.Sprint to construct and log a message.
func (l *Logger) Debug(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.logw(zerolog.DebugLevel, msg)
}

// Debugf uses fmt.Sprintf to log a formatted message.
func (l *Logger) Debugf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logw(zerolog.DebugLevel, msg)
}

// Debugw logs a message with some additional context.
func (l *Logger) Debugw(msg string, keysAndValues ...interface{}) {
	l.logw(zerolog.DebugLevel, msg, keysAndValues...)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *Logger) Error(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.logw(zerolog.ErrorLevel, msg)
}

// Errorf uses fmt.Sprintf to log a formatted message.
func (l *Logger) Errorf(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logw(zerolog.ErrorLevel, msg)
}

// Errorw logs a message with some additional context.
func (l *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	l.logw(zerolog.ErrorLevel, msg, keysAndValues...)
}

// Info uses fmt.Sprint to construct and log a message.
func (l *Logger) Info(v ...interface{}) {
	msg := fmt.Sprint(v...)
	l.logw(zerolog.InfoLevel, msg)
}

// Infof uses fmt.Sprintf to log a formatted message.
func (l *Logger) Infof(format string, v ...interface{}) {
	msg := fmt.Sprintf(format, v...)
	l.logw(zerolog.InfoLevel, msg)
}

// Infow logs a message with some additional context.
func (l *Logger) Infow(msg string, keysAndValues ...interface{}) {
	l.logw(zerolog.InfoLevel, msg, keysAndValues...)
}

// With adds a variadic number of key-values pairs to the logging context.
// The first element of the pair is used as the field key and should be a string.
// Passing a non-string key or an orphaned key panics.
func (l *Logger) With(args ...interface{}) *Logger {
	if len(args)%2 != 0 {
		panic("number of arguments is not a multiple of 2")
	}

	newLogger := &Logger{
		logger: l.logger,
		fields: l.fields,
	}

	for i := 0; i < len(args); i += 2 {
		key, ok := args[i].(string)
		if !ok {
			panic(fmt.Sprintf("argument %d is not a string", i))
		}

		newLogger.fields = append(newLogger.fields, keyValue{key: key, value: args[i+1]})
	}

	return newLogger
}

// logw is the common implementation for all logging methods.
func (l *Logger) logw(level zerolog.Level, msg string, keysAndValues ...interface{}) {
	// Start a new event at the specified level.
	event := l.logger.WithLevel(level)

	// Get the caller information.
	skip := 2
	_, file, line, ok := runtime.Caller(skip) // Adjust skip level if needed.
	if ok {
		for strings.Contains(file, "runtime/") || strings.Contains(file, "zerolog/") || strings.Contains(file, "logger/") || strings.Contains(file, "log/") {
			_, file, line, ok = runtime.Caller(skip + 1)
			if !ok {
				break
			}
			skip++
		}
		shortFile := fmt.Sprintf("%s/%s:%d", filepath.Base(filepath.Dir(file)), filepath.Base(file), line)
		// Add the caller info to the event.
		event = event.Str("caller", shortFile)
	}

	// Add the stored fields.
	for _, kv := range l.fields {
		event = event.Interface(kv.key, kv.value)
	}

	// Process keysAndValues.
	if len(keysAndValues)%2 != 0 {
		panic("number of arguments is not a multiple of 2")
	}

	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			panic(fmt.Sprintf("argument %d is not a string", i))
		}
		value := keysAndValues[i+1]
		event = event.Interface(key, value)
	}

	// Log the message.
	event.Msg(msg)
}

// --- Standard Library Logging Interface Methods ---

// Print logs the given arguments at Info level.
func (l *Logger) Print(v ...interface{}) {
	l.Info(v...)
}

// Printf logs a formatted message at Info level.
func (l *Logger) Printf(format string, v ...interface{}) {
	l.Infof(format, v...)
}

// Println logs the given arguments (with spaces) at Info level.
func (l *Logger) Println(v ...interface{}) {
	// fmt.Sprintln adds a newline; we use Info to keep the output style unchanged.
	l.Info(fmt.Sprintln(v...))
}

// Fatal logs the given arguments at Error level and exits the program.
func (l *Logger) Fatal(v ...interface{}) {
	l.Error(v...)
	os.Exit(1)
}

// Fatalf logs a formatted message at Error level and exits the program.
func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Errorf(format, v...)
	os.Exit(1)
}

// Fatalln logs the given arguments (with spaces) at Error level and exits the program.
func (l *Logger) Fatalln(v ...interface{}) {
	l.Error(fmt.Sprintln(v...))
	os.Exit(1)
}

// Panic logs the given arguments at Error level and panics.
func (l *Logger) Panic(v ...interface{}) {
	s := fmt.Sprint(v...)
	l.Error(s)
	panic(s)
}

// Panicf logs a formatted message at Error level and panics.
func (l *Logger) Panicf(format string, v ...interface{}) {
	s := fmt.Sprintf(format, v...)
	l.Error(s)
	panic(s)
}

// Panicln logs the given arguments (with spaces) at Error level and panics.
func (l *Logger) Panicln(v ...interface{}) {
	s := fmt.Sprintln(v...)
	l.Error(s)
	panic(s)
}

// Write implements the io.Writer interface so that Logger can be used
// as an output for packages that expect an io.Writer (such as the standard log package).
func (l *Logger) Write(p []byte) (n int, err error) {
	// Log the byte slice as an Info-level message.
	l.Info(string(p))
	return len(p), nil
}
