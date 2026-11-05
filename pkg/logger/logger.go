package logger

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(defaultLogger)
}

// Init 使用简单配置初始化日志系统
func Init(level string) {
	initWithValues(level, "both", "./logs", 100, 10, 30)
}

func initWithValues(level, output, logDir string, maxSize, maxBackups, maxAge int) {
	var logLevel slog.Level
	switch level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	var writers []io.Writer
	opts := &slog.HandlerOptions{Level: logLevel}

	if output == "stdout" || output == "both" {
		writers = append(writers, os.Stdout)
	}

	if output == "file" || output == "both" {
		if err := os.MkdirAll(logDir, 0755); err != nil {
			slog.Error("创建日志目录失败", "error", err)
			return
		}

		now := time.Now()
		logFile := filepath.Join(logDir,
			filepath.Base(os.Args[0])+"-"+now.Format("2006-01-02")+".log")

		lumberjackWriter := &lumberjack.Logger{
			Filename:   logFile,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			LocalTime:  true,
			Compress:   true,
		}
		writers = append(writers, lumberjackWriter)
	}

	if len(writers) > 0 {
		var writer io.Writer
		if len(writers) == 1 {
			writer = writers[0]
		} else {
			writer = io.MultiWriter(writers...)
		}
		defaultLogger = slog.New(slog.NewTextHandler(writer, opts))
		slog.SetDefault(defaultLogger)
	}
}

func Debug(msg string, args ...any) { slog.Debug(msg, args...) }
func Info(msg string, args ...any)  { slog.Info(msg, args...) }
func Warn(msg string, args ...any)  { slog.Warn(msg, args...) }
func Error(msg string, args ...any) { slog.Error(msg, args...) }
func Fatal(msg string, args ...any) {
	slog.Error(msg, args...)
	os.Exit(1)
}
