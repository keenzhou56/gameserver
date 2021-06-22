package metric

import (
	"sync"
	"time"
)

// RollingPolicy is a policy for ring window based on time duration.
// RollingPolicy moves bucket offset with time duration.
// e.g. If the last point is appended one bucket duration ago,
// RollingPolicy will increment current offset.
type RollingPolicy struct {
	mu     sync.RWMutex
	size   int
	window *Window
	offset int

	bucketDuration time.Duration
	lastAppendTime time.Time
}

// RollingPolicyOpts contains the arguments for creating RollingPolicy.
type RollingPolicyOpts struct {
	BucketDuration time.Duration
}

// NewRollingPolicy creates a new RollingPolicy based on the given window and RollingPolicyOpts.
func NewRollingPolicy(window *Window, opts RollingPolicyOpts) *RollingPolicy {
	return &RollingPolicy{
		window: window,
		size:   window.Size(),
		offset: 0,

		bucketDuration: opts.BucketDuration,
		lastAppendTime: time.Now(),
	}
}

func (r *RollingPolicy) timespan() int {
	return int(time.Since(r.lastAppendTime) / r.bucketDuration)
}

func (r *RollingPolicy) add(f func(offset int, val float64), val float64) {
	r.mu.Lock()
	timespan := r.timespan()
	if timespan > 0 {
		var offset int
		// reset the expired buckets
		for i := 0; i < timespan && i < r.size; i++ {
			offset = (r.offset + 1 + i) % r.size
			r.window.ResetBucket(offset)
		}
		r.offset = offset
		r.lastAppendTime = r.lastAppendTime.Add(time.Duration(timespan * int(r.bucketDuration)))
	}
	f(r.offset, val)
	r.mu.Unlock()
}

// Append appends the given points to the window.
func (r *RollingPolicy) Append(val float64) {
	r.add(r.window.Append, val)
}

// Add adds the given value to the latest point within bucket.
func (r *RollingPolicy) Add(val float64) {
	r.add(r.window.Add, val)
}

// Reduce applies the reduction function to all buckets within the window.
func (r *RollingPolicy) Reduce(f func(Iterator) float64) (val float64) {
	r.mu.RLock()
	timespan := r.timespan()
	if count := r.size - timespan; count > 0 {
		offset := r.offset + timespan + 1
		if offset >= r.size {
			offset = offset - r.size
		}
		val = f(r.window.Iterator(offset, count))
	}
	r.mu.RUnlock()
	return val
}
