// Copyright 2018 The Nakama Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func SetupLogging(config Config) (*zap.Logger, *zap.Logger) {
	consoleLogger := NewJSONLogger(os.Stdout, config.GetLogger().Level)
	fileLogger := NewJSONFileLogger(consoleLogger, config.GetLogger().File, config.GetLogger().Level)

	if fileLogger != nil {
		multiLogger := NewMultiLogger(consoleLogger, fileLogger)

		if config.GetLogger().Stdout {
			zap.RedirectStdLog(multiLogger)
			return multiLogger, multiLogger
		} else {
			zap.RedirectStdLog(fileLogger)
			return fileLogger, multiLogger
		}
	}

	zap.RedirectStdLog(consoleLogger)
	return consoleLogger, consoleLogger
}

func NewJSONFileLogger(consoleLogger *zap.Logger, fpath string, level string) *zap.Logger {
	if len(fpath) == 0 {
		return nil
	}

	output, err := os.OpenFile(fpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		consoleLogger.Fatal("Could not create log file", zap.Error(err))
		return nil
	}

	return NewJSONLogger(output, level)
}

func NewMultiLogger(loggers ...*zap.Logger) *zap.Logger {
	cores := make([]zapcore.Core, 0, len(loggers))
	for _, logger := range loggers {
		cores = append(cores, logger.Core())
	}

	teeCore := zapcore.NewTee(cores...)
	options := []zap.Option{zap.AddStacktrace(zap.ErrorLevel)}
	return zap.New(teeCore, options...)
}

func NewJSONLogger(output *os.File, level string) *zap.Logger {
	jsonEncoder := zapcore.NewJSONEncoder(zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	})

	zapLevel := zapcore.InfoLevel
	switch string(level) {
	case "debug", "DEBUG":
		zapLevel = zapcore.DebugLevel
	case "info", "INFO":
		zapLevel = zapcore.InfoLevel
	case "warn", "WARN":
		zapLevel = zapcore.WarnLevel
	case "error", "ERROR":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	core := zapcore.NewCore(jsonEncoder, zapcore.Lock(output), zapLevel)
	options := []zap.Option{zap.AddStacktrace(zap.ErrorLevel)}
	return zap.New(core, options...)
}
