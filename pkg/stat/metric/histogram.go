package metric

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

// HistogramVecOpts is an alias of vecopts.
type HistogramVecOpts struct {
	Namespace string
	Subsystem string
	Name      string
	Help      string
	Labels    []string
	Buckets   []float64
}

// HistogramVec gauge vec.
type HistogramVec interface {
	// Observe adds a single observation to the histogram.
	Observe(v int64, labels ...string)
}

// Histogram prom histogram collection.
type promHistogramVec struct {
	histogram *prom.HistogramVec
}

// NewHistogramVec new a historgram vec.
func NewHistogramVec(cfg *HistogramVecOpts) HistogramVec {
	if cfg == nil {
		return nil
	}
	vec := prom.NewHistogramVec(
		prom.HistogramOpts{
			Namespace: cfg.Namespace,
			Subsystem: cfg.Subsystem,
			Name:      cfg.Name,
			Help:      cfg.Help,
			Buckets:   cfg.Buckets,
		}, cfg.Labels)
	prom.MustRegister(vec)
	return &promHistogramVec{
		histogram: vec,
	}
}

// Timing adds a single observation to the histogram.
func (histogram *promHistogramVec) Observe(v int64, labels ...string) {
	histogram.histogram.WithLabelValues(labels...).Observe(float64(v))
}
