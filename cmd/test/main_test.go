package main

import (
	"sync"
	"testing"
)

var (
	lock = new(sync.Mutex)
)

func lockTest() {
	lock.Lock()
	lock.Unlock()
}
func lockDeferTest() {
	lock.Lock()
	defer lock.Unlock()
}

// BenchmarkTest ...
func BenchmarkTest(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lockTest()
	}
}

// BenchmarkTestDefer ...
func BenchmarkTestDefer(b *testing.B) {
	for i := 0; i < b.N; i++ {
		lockDeferTest()
	}
}
