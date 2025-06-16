package logger

import (
	"os"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a wrapper around zap.Logger
type Logger struct {
	*zap.Logger
}

// Config holds logger configuration
type Config struct {
	LogFile     string
	LogToFile   bool
	LogToStdout bool
	MaxSize     int
	MaxBackups  int
	MaxAge      int
	Compress    bool
}

// DefaultConfig returns default logger configuration
func DefaultConfig() Config {
	return Config{
		LogFile:     "logs/app.log",
		LogToFile:   true,
		LogToStdout: true,
		MaxSize:     100, // MB
		MaxBackups:  3,   // files
		MaxAge:      28,  // days
		Compress:    true,
	}
}

// New creates a new logger instance
func New(config Config) *Logger {
	// Create encoder config
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	// Create core for file output if enabled
	var fileCore zapcore.Core
	if config.LogToFile {
		fileWriter := zapcore.AddSync(&lumberjack.Logger{
			Filename:   config.LogFile,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		})
		fileCore = zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			fileWriter,
			zapcore.InfoLevel,
		)
	}

	// Create core for stdout output if enabled
	var stdoutCore zapcore.Core
	if config.LogToStdout {
		stdoutWriter := zapcore.AddSync(os.Stdout)
		stdoutCore = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			stdoutWriter,
			zapcore.InfoLevel,
		)
	}

	// Combine cores
	var core zapcore.Core
	switch {
	case config.LogToFile && config.LogToStdout:
		core = zapcore.NewTee(fileCore, stdoutCore)
	case config.LogToFile:
		core = fileCore
	case config.LogToStdout:
		core = stdoutCore
	default:
		// If no output is configured, use stdout as fallback
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			zapcore.InfoLevel,
		)
	}

	// Create logger
	logger := zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	return &Logger{logger}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...zapcore.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...zapcore.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...zapcore.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error logs an error message with error and additional fields
func (l *Logger) Error(msg string, err error, fields ...zapcore.Field) {
	allFields := append([]zapcore.Field{zap.Error(err)}, fields...)
	l.Logger.Error(msg, allFields...)
}

// Fatal logs a fatal message
func (l *Logger) Fatal(msg string, fields ...zapcore.Field) {
	l.Logger.Fatal(msg, fields...)
}
