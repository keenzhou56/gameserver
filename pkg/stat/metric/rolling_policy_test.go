package metric

import (
	"testing"
	"time"

	"gopkg.in/go-playground/assert.v1"
)

func TestRollingPolicyTimespan(t *testing.T) {
	size := 3
	bucketDuration := time.Second
	opts := RollingGaugeOpts{
		Size:           size,
		BucketDuration: bucketDuration,
	}
	window := NewWindow(WindowOpts{Size: opts.Size})
	policy := NewRollingPolicy(window, RollingPolicyOpts{BucketDuration: opts.BucketDuration})

	// timespan = 0
	assert.Equal(t, policy.timespan(), 0)

	// timespan < r.size
	time.Sleep(bucketDuration * 2)
	assert.Equal(t, policy.timespan(), 2)

	// timespan > r.size
	time.Sleep(bucketDuration * 2)
	assert.Equal(t, policy.timespan(), 4)
}

func TestRollingPolicyAdd(t *testing.T) {
	size := 3
	bucketDuration := time.Second
	opts := RollingGaugeOpts{
		Size:           size,
		BucketDuration: bucketDuration,
	}
	window := NewWindow(WindowOpts{Size: opts.Size})
	policy := NewRollingPolicy(window, RollingPolicyOpts{BucketDuration: opts.BucketDuration})

	listBuckets := func() [][]float64 {
		buckets := make([][]float64, 0)
		for _, bucket := range policy.window.window {
			buckets = append(buckets, bucket.Points)
		}
		return buckets
	}

	time.Sleep(bucketDuration + bucketDuration/2)
	policy.Append(1)
	time.Sleep(bucketDuration/2 + time.Millisecond)
	policy.Append(2)
	assert.Equal(t, [][]float64{{}, {1}, {2}}, listBuckets())

	// cross window
	policy.offset = 1
	policy.lastAppendTime = time.Now().Add(-bucketDuration * 3)
	policy.Append(3)
	assert.Equal(t, [][]float64{{}, {3}, {}}, listBuckets())

	// cross multi window
	policy.offset = 1
	policy.lastAppendTime = time.Now().Add(-bucketDuration * 10)
	policy.Append(3)
	assert.Equal(t, [][]float64{{}, {3}, {}}, listBuckets())

	policy.offset = 1
	policy.lastAppendTime = time.Now().Add(-bucketDuration * 10)
	policy.Append(3)
	assert.Equal(t, [][]float64{{}, {3}, {}}, listBuckets())
}

func TestRollingPolicyReset(t *testing.T) {
	size := 5
	bucketDuration := time.Second
	opts := RollingGaugeOpts{
		Size:           size,
		BucketDuration: bucketDuration,
	}
	window := NewWindow(WindowOpts{Size: opts.Size})
	policy := NewRollingPolicy(window, RollingPolicyOpts{BucketDuration: opts.BucketDuration})

	listBuckets := func() [][]float64 {
		buckets := make([][]float64, 0)
		for _, bucket := range policy.window.window {
			buckets = append(buckets, bucket.Points)
		}
		return buckets
	}

	resetPolicy := func() {
		for i := 0; i < size; i++ {
			policy.window.ResetBucket(i)
			policy.window.Append(i, float64(i))
		}
	}

	resetPolicy()
	policy.offset = 3
	policy.lastAppendTime = time.Now()
	policy.Append(5)
	assert.Equal(t, [][]float64{{0}, {1}, {2}, {3, 5}, {4}}, listBuckets())

	resetPolicy()
	policy.offset = 2
	policy.lastAppendTime = time.Now().Add(-bucketDuration * 2)
	policy.Append(5)
	assert.Equal(t, [][]float64{{0}, {1}, {2}, {}, {5}}, listBuckets())

	resetPolicy()
	policy.offset = 2
	policy.lastAppendTime = time.Now().Add(-bucketDuration * 4)
	policy.Append(5)
	assert.Equal(t, [][]float64{{}, {5}, {2}, {}, {}}, listBuckets())

	resetPolicy()
	policy.offset = 2
	policy.lastAppendTime = time.Now().Add(-bucketDuration * 5)
	policy.Append(5)
	assert.Equal(t, [][]float64{{}, {}, {5}, {}, {}}, listBuckets())

	resetPolicy()
	policy.offset = 2
	policy.lastAppendTime = time.Now().Add(-bucketDuration * 6)
	policy.Append(5)
	assert.Equal(t, [][]float64{{}, {}, {5}, {}, {}}, listBuckets())
}
