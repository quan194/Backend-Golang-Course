package logger

var (
	globalLogger Logger
)

const (
	TypeZap = "zap"

	LevelDebug = "debug"
	LevelInfo  = "info"
	LevelWarn  = "warn"
	LevelError = "error"

	OutputConsole = "console"
	OutputFile    = "file"
)

type Logger interface {
	Debugf(template string, args ...interface{})
	Debug(msg string, fields ...Field)

	Infof(template string, args ...interface{})
	Info(msg string, fields ...Field)

	Warnf(template string, args ...interface{})
	Warn(msg string, fields ...Field)

	Errorf(template string, args ...interface{})
	Error(msg string, fields ...Field)
}

type (
	Config struct {
		Type     string // zap
		Level    string // debug, info, warn, error
		Output   string // console, file
		Filename string // file path if Output=file
		UseJSON  bool   // structured json vs console text
	}

	Field struct {
		Key   string
		Value interface{}
	}
)

func init() {
	// set log with default config
	SetLogger(Config{
		Type:    TypeZap,
		Level:   LevelInfo,
		Output:  OutputConsole,
		UseJSON: false,
	})
}

func SetLogger(cfg Config) {
	switch cfg.Type {
	case TypeZap:
		l, err := newZapLogger(cfg)
		if err != nil {
			Error("logger init failed", E(err))
			return
		}
		globalLogger = l
	default:
		// do nothing, use default logger
	}
}

func Debugf(template string, args ...interface{}) {
	globalLogger.Debugf(template, args...)
}

func Debug(msg string, fields ...Field) {
	globalLogger.Debug(msg, fields...)
}

func Infof(template string, args ...interface{}) {
	globalLogger.Infof(template, args...)
}

func Info(msg string, fields ...Field) {
	globalLogger.Info(msg, fields...)
}

func Warnf(template string, args ...interface{}) {
	globalLogger.Warnf(template, args...)
}

func Warn(msg string, fields ...Field) {
	globalLogger.Warn(msg, fields...)
}

func Errorf(template string, args ...interface{}) {
	globalLogger.Errorf(template, args...)
}

func Error(msg string, fields ...Field) {
	globalLogger.Error(msg, fields...)
}

func F(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

func E(err error) Field {
	return F("error", err)
}
