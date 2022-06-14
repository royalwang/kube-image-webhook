package logr

import (
	"context"
	"fmt"
	"github.com/go-logr/logr"
	"github.com/slok/kubewebhook/v2/pkg/log"
)

type logger struct {
	log logr.Logger
}

func NewLogger(log logr.Logger) *logger {
	return &logger{
		log: log,
	}
}

func (l logger) Infof(format string, args ...interface{}) {
	l.log.Info(fmt.Sprintf(format, args...))
}

func (l logger) Warningf(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func (l logger) Errorf(format string, args ...interface{}) {
	l.Infof(format, args...)
}

func (l logger) Debugf(format string, args ...interface{}) {
	l.log.V(1).Info(fmt.Sprintf(format, args...))
}

func (l logger) WithValues(values map[string]interface{}) log.Logger {
	var kv []any
	for k, v := range values {
		kv = append(kv, k)
		kv = append(kv, fmt.Sprintf("%v", v))
	}
	newLogger := l.log.WithValues(kv...)
	return NewLogger(newLogger)
}

func (l logger) WithCtxValues(ctx context.Context) log.Logger {
	return l.WithValues(log.ValuesFromCtx(ctx))
}

func (l logger) SetValuesOnCtx(parent context.Context, values map[string]interface{}) context.Context {
	return log.CtxWithValues(parent, values)
}
