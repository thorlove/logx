package log

import (
	"fmt"
	"go.uber.org/zap/zapcore"
	"sync/atomic"
	"testing"
)

func TestLogger(t *testing.T) {
	count := &atomic.Int64{}
	log := New(
		WithLevel(zapcore.InfoLevel),
		WithField(F{"name": "robin"}),
		WithHooks(func(e zapcore.Entry) error {
			fmt.Println("count:", count.Add(1), "msg:", e.Message)
			return nil
		})).Named("LOGX")
	log.Errorw("error msg", "key1", "value", "key2", "value2")
	log.Infoln("info msg")
	log.Debugln("debug msg")
}
