package bbr

import (
	"context"
	"fmt"
	"gameserver/pkg/rate/limit"
	"gameserver/pkg/stat/metric"
	"math"
	"math/rand"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func confForTest() *Config {
	return &Config{
		Window:       time.Second,
		WinBucket:    10,
		CPUThreshold: 800,
	}
}

func warmup(bbr *BBR, count int) {
	for i := 0; i < count; i++ {
		done, err := bbr.Allow(context.TODO())
		time.Sleep(time.Millisecond * 1)
		if err == nil {
			done(limit.DoneInfo{Op: limit.Success})
		}
	}
}

func TestBBR(t *testing.T) {
	cfg := &Config{
		Window:       time.Second * 5,
		WinBucket:    50,
		CPUThreshold: 100,
	}
	limiter := newLimiter(cfg)
	var wg sync.WaitGroup
	var drop int64
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 300; i++ {
				f, err := limiter.Allow(context.TODO())
				if err != nil {
					atomic.AddInt64(&drop, 1)
				} else {
					count := rand.Intn(100)
					time.Sleep(time.Millisecond * time.Duration(count))
					f(limit.DoneInfo{Op: limit.Success})
				}
			}
		}()
	}
	wg.Wait()
	fmt.Println("drop: ", drop)
}

func TestBBRMaxPass(t *testing.T) {
	bucketDuration := time.Millisecond * 100
	passStat := metric.NewRollingCounter(metric.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	for i := 1; i <= 10; i++ {
		passStat.Add(int64(i * 100))
		time.Sleep(bucketDuration)
	}
	bbr := &BBR{
		passStat: passStat,
	}
	assert.Equal(t, int64(1000), bbr.maxPASS())

	// default max pass is equal to 1.
	passStat = metric.NewRollingCounter(metric.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	bbr = &BBR{
		passStat: passStat,
	}
	assert.Equal(t, int64(1), bbr.maxPASS())
}

func TestBBRMinRt(t *testing.T) {
	bucketDuration := time.Millisecond * 100
	rtStat := metric.NewRollingCounter(metric.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	for i := 0; i < 10; i++ {
		for j := i*10 + 1; j <= i*10+10; j++ {
			rtStat.Add(int64(j))
		}
		if i != 9 {
			time.Sleep(bucketDuration)
		}
	}
	bbr := &BBR{
		rtStat: rtStat,
	}
	assert.Equal(t, int64(6), bbr.minRT())

	// default max min rt is equal to maxFloat64.
	bucketDuration = time.Millisecond * 100
	rtStat = metric.NewRollingCounter(metric.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	bbr = &BBR{
		rtStat: rtStat,
	}
	assert.Equal(t, int64(math.Ceil(math.MaxFloat64)), bbr.minRT())
}

func TestBBRMaxQps(t *testing.T) {
	bucketDuration := time.Millisecond * 100
	passStat := metric.NewRollingCounter(metric.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	rtStat := metric.NewRollingCounter(metric.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	for i := 0; i < 10; i++ {
		passStat.Add(int64((i + 1) * 100))
		for j := i*10 + 1; j <= i*10+10; j++ {
			rtStat.Add(int64(j))
		}
		if i != 9 {
			time.Sleep(bucketDuration)
		}
	}
	bbr := &BBR{
		passStat:        passStat,
		rtStat:          rtStat,
		winBucketPerSec: 10,
	}
	assert.Equal(t, int64(60), bbr.maxFlight())
}

func TestBBRShouldDrop(t *testing.T) {
	var cpu int64
	cpuGetter := func() int64 {
		return cpu
	}
	bucketDuration := time.Millisecond * 100
	passStat := metric.NewRollingCounter(metric.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	rtStat := metric.NewRollingCounter(metric.RollingCounterOpts{Size: 10, BucketDuration: bucketDuration})
	for i := 0; i < 10; i++ {
		passStat.Add(int64((i + 1) * 100))
		for j := i*10 + 1; j <= i*10+10; j++ {
			rtStat.Add(int64(j))
		}
		if i != 9 {
			time.Sleep(bucketDuration)
		}
	}
	bbr := &BBR{
		cpu:             cpuGetter,
		passStat:        passStat,
		rtStat:          rtStat,
		winBucketPerSec: 10,
		conf:            defaultConf,
	}
	// cpu >=  800, inflight < maxQps
	cpu = 800
	bbr.inFlight = 50
	assert.Equal(t, false, bbr.shouldDrop())

	// cpu >=  800, inflight > maxQps
	cpu = 800
	bbr.inFlight = 80
	assert.Equal(t, true, bbr.shouldDrop())

	// cpu < 800, inflight > maxQps, cold duration
	cpu = 700
	bbr.inFlight = 80
	assert.Equal(t, true, bbr.shouldDrop())

	// cpu < 800, inflight > maxQps
	time.Sleep(2 * time.Second)
	cpu = 700
	bbr.inFlight = 80
	assert.Equal(t, false, bbr.shouldDrop())
}

func TestGroup(t *testing.T) {
	cfg := &Config{
		Window:       time.Second * 5,
		WinBucket:    50,
		CPUThreshold: 100,
	}
	group := NewGroup(cfg)
	t.Run("get", func(t *testing.T) {
		limiter := group.Get("test")
		assert.NotNil(t, limiter)
	})
}

func BenchmarkBBRAllowUnderLowLoad(b *testing.B) {
	bbr := newLimiter(confForTest()).(*BBR)
	bbr.cpu = func() int64 {
		return 500
	}
	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		done, err := bbr.Allow(context.TODO())
		if err == nil {
			done(limit.DoneInfo{Op: limit.Success})
		}
	}
}

func BenchmarkBBRAllowUnderHighLoad(b *testing.B) {
	bbr := newLimiter(confForTest()).(*BBR)
	bbr.cpu = func() int64 {
		return 900
	}
	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		if i%10000 == 0 {
			maxFlight := bbr.maxFlight()
			if maxFlight != 0 {
				bbr.inFlight = rand.Int63n(bbr.maxFlight() * 2)
			}
		}
		done, err := bbr.Allow(context.TODO())
		if err == nil {
			done(limit.DoneInfo{Op: limit.Success})
		}
	}
}

func BenchmarkBBRShouldDropUnderLowLoad(b *testing.B) {
	bbr := newLimiter(confForTest()).(*BBR)
	bbr.cpu = func() int64 {
		return 500
	}
	warmup(bbr, 10000)
	bbr.inFlight = 1000
	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		bbr.shouldDrop()
	}
}

func BenchmarkBBRShouldDropUnderHighLoad(b *testing.B) {
	bbr := newLimiter(confForTest()).(*BBR)
	bbr.cpu = func() int64 {
		return 900
	}
	warmup(bbr, 10000)
	bbr.prevDrop.Store(time.Now())
	bbr.inFlight = 1000

	b.ResetTimer()
	for i := 0; i <= b.N; i++ {
		bbr.shouldDrop()
	}
}
