package stdlog

import (
	"gameserver/pkg/log"
	"os"
	"testing"
)

type Discard int

func (d Discard) Write(p []byte) (n int, err error) { return len(p), nil }
func (d Discard) Close() (err error)                { return }

func TestLogger(t *testing.T) {
	logger := NewLogger(Writer(os.Stdout))
	defer logger.Close()

	log.Debug(logger).Print("log", "test debug")
	log.Info(logger).Print("log", "test info")
	log.Warn(logger).Print("log", "test warn")
	log.Error(logger).Print("log", "test error")
}

func BenchmarkLoggerPrint(b *testing.B) {
	b.SetParallelism(100)
	logger := NewLogger(Writer(Discard(0)))
	defer logger.Close()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Print("log", "test")
		}
	})
}

func BenchmarkLoggerHelperInfo(b *testing.B) {
	b.SetParallelism(100)
	logger := NewLogger(Writer(Discard(0)))
	defer logger.Close()
	h := log.NewHelper("test", logger)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Info("test")
		}
	})
}

func BenchmarkLoggerHelperInfof(b *testing.B) {
	b.SetParallelism(100)
	logger := NewLogger(Writer(Discard(0)))
	defer logger.Close()
	h := log.NewHelper("test", logger)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			h.Infof("log %s", "test")
		}
	})
}

func BenchmarkLoggerHelperInfow(b *testing.B) {
	b.SetParallelism(100)
	logger := NewLogger(Writer(Discard(0)))
	defer logger.Close()
	log := log.NewHelper("test", logger)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			log.Infow("log", "test")
		}
	})
}
