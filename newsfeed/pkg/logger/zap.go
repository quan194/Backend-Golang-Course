package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapLogger uses zap log
type zapLogger struct {
	logger *zap.Logger
}

func newZapLogger(cfg Config) (*zapLogger, error) {
	// log level
	var level zapcore.Level
	switch cfg.Level {
	case LevelDebug:
		level = zap.DebugLevel
	case LevelWarn:
		level = zap.WarnLevel
	case LevelError:
		level = zap.ErrorLevel
	default:
		level = zap.InfoLevel
	}

	// encode
	var encoderCfg zapcore.EncoderConfig
	if cfg.UseJSON {
		encoderCfg = zap.NewProductionEncoderConfig()
	} else {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	}
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder

	var enc zapcore.Encoder
	if cfg.UseJSON {
		enc = zapcore.NewJSONEncoder(encoderCfg)
	} else {
		enc = zapcore.NewConsoleEncoder(encoderCfg)
	}

	// output
	var ws zapcore.WriteSyncer
	if cfg.Output == "file" && cfg.Filename != "" {
		file, err := os.OpenFile(cfg.Filename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, fmt.Errorf("error opening log file %s: %v", cfg.Filename, err)
		}
		ws = zapcore.AddSync(file)
	} else {
		ws = zapcore.AddSync(os.Stdout)
	}

	// init
	core := zapcore.NewCore(enc, ws, level)

	// options
	opts := []zap.Option{zap.AddCaller(), zap.AddCallerSkip(2)}

	l := &zapLogger{
		logger: zap.New(core, opts...),
	}

	return l, nil
}

func (l *zapLogger) Debugf(template string, args ...interface{}) {
	l.logger.Sugar().Debugf(template, args...)
}

func (l *zapLogger) Debug(msg string, fields ...Field) {
	zapFields := toZapFields(fields)
	l.logger.Sugar().With(zapFields...).Debug(msg)
}

func (l *zapLogger) Infof(template string, args ...interface{}) {
	l.logger.Sugar().Infof(template, args...)
}

func (l *zapLogger) Info(msg string, fields ...Field) {
	zapFields := toZapFields(fields)
	l.logger.Sugar().With(zapFields...).Info(msg)
}

func (l *zapLogger) Warnf(template string, args ...interface{}) {
	l.logger.Sugar().Warnf(template, args...)
}

func (l *zapLogger) Warn(msg string, fields ...Field) {
	zapFields := toZapFields(fields)
	l.logger.Sugar().With(zapFields...).Warn(msg)
}

func (l *zapLogger) Errorf(template string, args ...interface{}) {
	l.logger.Sugar().Errorf(template, args...)
}

func (l *zapLogger) Error(msg string, fields ...Field) {
	zapFields := toZapFields(fields)
	l.logger.Sugar().With(zapFields...).Error(msg)
}

func toZapFields(fields []Field) []any {
	zapFields := make([]any, len(fields))
	for i, f := range fields {
		zapFields[i] = zap.Any(f.Key, f.Value)
	}
	return zapFields
}
