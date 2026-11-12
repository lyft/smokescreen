package smokescreen

import (
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTunnelLimiter_NilLimiter(t *testing.T) {
	var tl *TunnelLimiter

	// Nil limiter should always allow
	assert.True(t, tl.Acquire())
	tl.Release() // Should not panic
	assert.Equal(t, int64(0), tl.ActiveCount())
	assert.Equal(t, int64(0), tl.MaxTunnels())
}

func TestTunnelLimiter_ZeroLimit(t *testing.T) {
	tl := NewTunnelLimiter(0, nil)

	// Zero limit means disabled - should always allow
	assert.True(t, tl.Acquire())
	assert.True(t, tl.Acquire())
	assert.True(t, tl.Acquire())
	tl.Release()
	tl.Release()
	tl.Release()
}

func TestTunnelLimiter_BasicLimit(t *testing.T) {
	tl := NewTunnelLimiter(3, nil)

	// Should allow up to 3
	assert.True(t, tl.Acquire())
	assert.Equal(t, int64(1), tl.ActiveCount())

	assert.True(t, tl.Acquire())
	assert.Equal(t, int64(2), tl.ActiveCount())

	assert.True(t, tl.Acquire())
	assert.Equal(t, int64(3), tl.ActiveCount())

	// Fourth should fail
	assert.False(t, tl.Acquire())
	assert.Equal(t, int64(3), tl.ActiveCount())

	// Release one
	tl.Release()
	assert.Equal(t, int64(2), tl.ActiveCount())

	// Now should allow one more
	assert.True(t, tl.Acquire())
	assert.Equal(t, int64(3), tl.ActiveCount())

	// Cleanup
	tl.Release()
	tl.Release()
	tl.Release()
	assert.Equal(t, int64(0), tl.ActiveCount())
}

func TestTunnelLimiter_ConcurrentAccess(t *testing.T) {
	const limit = 10
	const goroutines = 100

	tl := NewTunnelLimiter(limit, nil)

	var acquired int64
	var wg sync.WaitGroup

	// Try to acquire from many goroutines simultaneously
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if tl.Acquire() {
				atomic.AddInt64(&acquired, 1)
			}
		}()
	}

	wg.Wait()

	// Should have acquired exactly 'limit' slots
	assert.Equal(t, int64(limit), acquired)
	assert.Equal(t, int64(limit), tl.ActiveCount())

	// Release all
	for i := 0; i < limit; i++ {
		tl.Release()
	}
	assert.Equal(t, int64(0), tl.ActiveCount())
}

func TestTunnelLimiter_MaxTunnels(t *testing.T) {
	tl := NewTunnelLimiter(42, nil)
	assert.Equal(t, int64(42), tl.MaxTunnels())
}
