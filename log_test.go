package log

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"sync/atomic"
	"testing"
)

func TestHooks(t *testing.T) {
	count := &atomic.Int64{}
	log := New(
		WithHooks(func(e zapcore.Entry) error {
			fmt.Println("count:", count.Add(1), "msg:", e.Message)
			return nil
		}))
	log.Infoln("info msg")
}

func TestNamed(t *testing.T) {
	log := New()
	log.Named("LOGX").Infoln("info msg")
}
func TestField(t *testing.T) {
	log := New(
		WithLevel(DebugLevel),
		WithField(F{"name": "robin"}),
	)
	log.Debugw("error msg", "key1", "value", "key2", "value2")
}

func TestEncoder(t *testing.T) {
	log := New(
		WithLevel(InfoLevel),
		WithConsoleSeparator("\t"),
		WithSampleEncoder(Console),
	)
	log.Infoln("info msg")
	log = New(
		WithSampleEncoder(JSON),
	)
	log.Errorw("error msg", "key1", "value", "key2", "value2")
}
