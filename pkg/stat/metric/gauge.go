package metric

import (
	"sync/atomic"

	prom "github.com/prometheus/client_golang/prometheus"
)

var _ Metric = &gauge{}

// Gauge stores a numerical value that can be add arbitrarily.
type Gauge interface {
	Metric
	// Sets sets the value to the given number.
	Set(int64)
}

// GaugeOpts is an alias of Opts.
type GaugeOpts Opts

type gauge struct {
	val int64
}

// NewGauge creates a new Gauge based on the GaugeOpts.
func NewGauge(opts GaugeOpts) Gauge {
	return &gauge{}
}

func (g *gauge) Add(val int64) {
	atomic.AddInt64(&g.val, val)
}

func (g *gauge) Set(val int64) {
	old := atomic.LoadInt64(&g.val)
	atomic.CompareAndSwapInt64(&g.val, old, val)
}

func (g *gauge) Value() int64 {
	return atomic.LoadInt64(&g.val)
}

// GaugeVecOpts is an alias of vecopts.
type GaugeVecOpts VectorOpts

// GaugeVec gauge vec.
type GaugeVec interface {
	// Set sets the Gauge to an arbitrary value.
	Set(v float64, labels ...string)
	// Inc increments the Gauge by 1. Use Add to increment it by arbitrary
	// values.
	Inc(labels ...string)
	// Add adds the given value to the Gauge. (The value can be negative,
	// resulting in a decrease of the Gauge.)
	Add(v float64, labels ...string)
}

// gaugeVec gauge vec.
type promGaugeVec struct {
	gauge *prom.GaugeVec
}

// NewGaugeVec .
func NewGaugeVec(cfg *GaugeVecOpts) GaugeVec {
	if cfg == nil {
		return nil
	}
	vec := prom.NewGaugeVec(
		prom.GaugeOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.Subsystem,
			Name:      cfg.Name,
			Help:      cfg.Help,
		}, cfg.Labels)
	prom.MustRegister(vec)
	return &promGaugeVec{
		gauge: vec,
	}
}

// Inc Inc increments the counter by 1. Use Add to increment it by arbitrary.
func (gauge *promGaugeVec) Inc(labels ...string) {
	gauge.gauge.WithLabelValues(labels...).Inc()
}

// Add Inc increments the counter by 1. Use Add to increment it by arbitrary.
func (gauge *promGaugeVec) Add(v float64, labels ...string) {
	gauge.gauge.WithLabelValues(labels...).Add(v)
}

// Set set the given value to the colltion.
func (gauge *promGaugeVec) Set(v float64, labels ...string) {
	gauge.gauge.WithLabelValues(labels...).Set(v)
}
