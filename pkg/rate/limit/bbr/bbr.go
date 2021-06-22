package bbr

import (
	"context"
	"fmt"
	"gameserver/pkg/container/group"
	"gameserver/pkg/ecode"
	"gameserver/pkg/rate/limit"
	"gameserver/pkg/stat/metric"
	"math"
	"sync"
	"sync/atomic"
	"time"

	cpustat "gameserver/pkg/stat/sys/cpu"
)

var (
	cpu         int64
	decay       = 0.95
	defaultConf = &Config{
		Window:       time.Second * 10,
		WinBucket:    100,
		CPUThreshold: 800,
	}
	cpuProcOnce sync.Once
)

type cpuGetter func() int64

func startCPUProc() {
	cpustat.Init()
	go cpuproc()
}

func min(l, r uint64) uint64 {
	if l < r {
		return l
	}
	return r
}

// cpu = cpuᵗ⁻¹ * decay + cpuᵗ * (1 - decay)
func cpuproc() {
	defer func() {
		if err := recover(); err != nil {
			fmt.Errorf("rate.limit.cpuproc() err(%+v)", err)
			go cpuproc()
		}
	}()
	ticker := time.NewTicker(time.Millisecond * 250)
	// EMA algorithm: https://blog.csdn.net/m0_38106113/article/details/81542863
	for range ticker.C {
		stat := &cpustat.Stat{}
		cpustat.ReadStat(stat)
		stat.Usage = min(stat.Usage, 1000)
		prevCPU := atomic.LoadInt64(&cpu)
		curCPU := int64(float64(prevCPU)*decay + float64(stat.Usage)*(1.0-decay))
		atomic.StoreInt64(&cpu, curCPU)
	}
}

// Stat contains the metrics's snapshot of bbr.
type Stat struct {
	CPU         int64
	InFlight    int64
	MaxInFlight int64
	MinRt       int64
	MaxPass     int64
}

// BBR implements bbr-like limiter.
// It is inspired by sentinel.
// https://github.com/alibaba/Sentinel/wiki/%E7%B3%BB%E7%BB%9F%E8%87%AA%E9%80%82%E5%BA%94%E9%99%90%E6%B5%81
type BBR struct {
	cpu             cpuGetter
	passStat        metric.RollingCounter
	rtStat          metric.RollingCounter
	inFlight        int64
	winBucketPerSec int64
	conf            *Config
	prevDrop        atomic.Value
}

// Config contains configs of bbr limiter.
type Config struct {
	Enabled      bool
	Window       time.Duration
	WinBucket    int
	Rule         string
	Debug        bool
	CPUThreshold int64
}

func (l *BBR) maxPASS() int64 {
	val := int64(l.passStat.Reduce(func(iterator metric.Iterator) float64 {
		var result = 1.0
		for iterator.Next() {
			bucket := iterator.Bucket()
			count := 0.0
			for _, p := range bucket.Points {
				count += p
			}
			result = math.Max(result, count)
		}
		return result
	}))
	if val == 0 {
		return 1
	}
	return val
}

func (l *BBR) minRT() int64 {
	val := l.rtStat.Reduce(func(iterator metric.Iterator) float64 {
		var result = math.MaxFloat64
		for iterator.Next() {
			bucket := iterator.Bucket()
			if len(bucket.Points) == 0 {
				continue
			}
			total := 0.0
			for _, p := range bucket.Points {
				total += p
			}
			avg := total / float64(bucket.Count)
			result = math.Min(result, avg)
		}
		return result
	})
	return int64(math.Ceil(val))
}

func (l *BBR) maxFlight() int64 {
	return int64(math.Floor(float64(l.maxPASS()*l.minRT()*l.winBucketPerSec)/1000.0 + 0.5))
}

func (l *BBR) shouldDrop() bool {
	if l.cpu() < l.conf.CPUThreshold {
		prevDrop, ok := l.prevDrop.Load().(time.Time)
		if !ok {
			return false
		}
		if time.Since(prevDrop) <= time.Second {
			inFlight := atomic.LoadInt64(&l.inFlight)
			return inFlight > 1 && inFlight > l.maxFlight()
		}
		return false
	}
	inFlight := atomic.LoadInt64(&l.inFlight)
	drop := inFlight > 1 && inFlight > l.maxFlight()
	if drop {
		l.prevDrop.Store(time.Now())
	}
	return drop
}

// Stat tasks a snapshot of the bbr limiter.
func (l *BBR) Stat() Stat {
	return Stat{
		CPU:         l.cpu(),
		InFlight:    atomic.LoadInt64(&l.inFlight),
		MinRt:       l.minRT(),
		MaxPass:     l.maxPASS(),
		MaxInFlight: l.maxFlight(),
	}
}

// Allow checks all inbound traffic.
// Once overload is detected, it raises ecode.LimitExceed error.
func (l *BBR) Allow(ctx context.Context, opts ...limit.AllowOption) (func(info limit.DoneInfo), error) {
	allowOpts := limit.DefaultAllowOpts()
	for _, opt := range opts {
		opt.Apply(&allowOpts)
	}
	if l.shouldDrop() {
		return nil, ecode.LimitExceed
	}
	atomic.AddInt64(&l.inFlight, 1)
	stime := time.Now()
	return func(do limit.DoneInfo) {
		rt := int64(time.Since(stime) / time.Millisecond)
		l.rtStat.Add(rt)
		atomic.AddInt64(&l.inFlight, -1)
		switch do.Op {
		case limit.Success:
			l.passStat.Add(1)
			return
		default:
			return
		}
	}, nil
}

func newLimiter(conf *Config) limit.Limiter {
	if conf == nil {
		conf = defaultConf
	}
	size := conf.WinBucket
	bucketDuration := conf.Window / time.Duration(conf.WinBucket)
	passStat := metric.NewRollingCounter(metric.RollingCounterOpts{Size: size, BucketDuration: bucketDuration})
	rtStat := metric.NewRollingCounter(metric.RollingCounterOpts{Size: size, BucketDuration: bucketDuration})
	cpu := func() int64 {
		return atomic.LoadInt64(&cpu)
	}
	limiter := &BBR{
		cpu:             cpu,
		conf:            conf,
		passStat:        passStat,
		rtStat:          rtStat,
		winBucketPerSec: int64(time.Second) / (int64(conf.Window) / int64(conf.WinBucket)),
	}
	return limiter
}

// Group represents a class of BBRLimiter and forms a namespace in which
// units of BBRLimiter.
type Group struct {
	group *group.Group
}

// NewGroup new a limiter group container, if conf nil use default conf.
func NewGroup(conf *Config) *Group {
	cpuProcOnce.Do(startCPUProc)
	if conf == nil {
		conf = defaultConf
	}
	group := group.NewGroup(func() interface{} {
		return newLimiter(conf)
	})
	return &Group{
		group: group,
	}
}

// Get get a limiter by a specified key, if limiter not exists then make a new one.
func (g *Group) Get(key string) limit.Limiter {
	limiter := g.group.Get(key)
	return limiter.(limit.Limiter)
}
