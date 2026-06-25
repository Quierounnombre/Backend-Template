package main

import (
	"log/slog"
	"os"
	"io"
	"time"
	"gopkg.in/natefinch/lumberjack.v2"
)

func modifyAttr(groups []string, a slog.Attr) slog.Attr {
	if a.Key == slog.TimeKey {
		a.Value = slog.StringValue(a.Value.Time().Format(time.DateTime))
	}
	return a
}

func writter_loging(s *Settings) io.Writer {
	rotator := &lumberjack.Logger {
		Filename:		s.Logger.Path,
		MaxSize:		s.Logger.MaxSize,
		MaxBackups:		s.Logger.MaxBackups,
		MaxAge:			s.Logger.MaxAge,
		Compress:		s.Logger.Compress,
	}
	w := io.MultiWriter(rotator, os.Stdout)
	return (w)
}

func set_logger(s *Settings) {
	var logLevels = map[string]slog.Level {
		"debug": slog.LevelDebug,
		"info":  slog.LevelInfo,
		"warn":  slog.LevelWarn,
		"error": slog.LevelError,
	}

	level, ok := logLevels[s.Logger.Level]
	if !ok {
		level = slog.LevelInfo
	}
	w := writter_loging(s)
	opts := slog.HandlerOptions {
		Level: level,
		AddSource: s.Logger.Source,
		ReplaceAttr: modifyAttr,
	}
	handler := slog.NewJSONHandler(w, &opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
