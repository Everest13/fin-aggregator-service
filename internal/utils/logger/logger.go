package logger

import (
	"go.uber.org/zap"
)

var log *zap.Logger

func Init() error {
	var err error
	log, err = zap.NewProduction()
	return err
}

func Sync() {
	log.Sync()
}

func Info(msg string, key, val string) {
	log.Info(msg, zap.String(key, val))
}

func Warn(msg string, key, val string) {
	log.Warn(msg, zap.String(key, val))
}

func Error(msg string, err error) {
	log.Error(msg, zap.Error(err))
}

func ErrorWithFields(msg string, err error, kvs ...interface{}) {
	fields := []zap.Field{zap.Error(err)}

	for i := 0; i < len(kvs)-1; i += 2 {
		key, ok := kvs[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, zap.Any(key, kvs[i+1]))
	}

	log.Error(msg, fields...)
}

func WarnWithFields(msg string, kvs ...interface{}) {
	fields := []zap.Field{}

	for i := 0; i < len(kvs)-1; i += 2 {
		key, ok := kvs[i].(string)
		if !ok {
			continue
		}
		fields = append(fields, zap.Any(key, kvs[i+1]))
	}

	log.Warn(msg, fields...)
}

func Fatal(msg string, err error) {
	log.Fatal(msg, zap.Error(err))
}
