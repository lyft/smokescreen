package smokescreen

import (
	"errors"
	"sync/atomic"
)

// ErrTunnelLimitExceeded is returned when the maximum number of concurrent tunnels is reached.
var ErrTunnelLimitExceeded = errors.New("maximum concurrent connect tunnels exceeded")

// TunnelLimiter limits the number of concurrent CONNECT tunnels.
// Unlike the rate limiter which counts request processing time,
// this tracks actual long-lived tunnel connections.
type TunnelLimiter struct {
	maxTunnels    int64
	activeTunnels int64
	config        *Config
}

// NewTunnelLimiter creates a new tunnel limiter with the specified maximum.
// If max is 0 or negative, limiting is disabled.
func NewTunnelLimiter(max int, config *Config) *TunnelLimiter {
	return &TunnelLimiter{
		maxTunnels: int64(max),
		config:     config,
	}
}

// Acquire attempts to acquire a tunnel slot.
// Returns true if successful, false if at capacity.
func (tl *TunnelLimiter) Acquire() bool {
	if tl == nil || tl.maxTunnels <= 0 {
		return true // Limiting disabled
	}

	for {
		current := atomic.LoadInt64(&tl.activeTunnels)
		if current >= tl.maxTunnels {
			if tl.config != nil && tl.config.MetricsClient != nil {
				tl.config.MetricsClient.Incr("tunnels.concurrency_limited", 1)
			}
			return false
		}
		if atomic.CompareAndSwapInt64(&tl.activeTunnels, current, current+1) {
			return true
		}
	}
}

// Release releases a tunnel slot.
func (tl *TunnelLimiter) Release() {
	if tl == nil || tl.maxTunnels <= 0 {
		return
	}
	atomic.AddInt64(&tl.activeTunnels, -1)
}

// ActiveCount returns the current number of active tunnels.
func (tl *TunnelLimiter) ActiveCount() int64 {
	if tl == nil {
		return 0
	}
	return atomic.LoadInt64(&tl.activeTunnels)
}

// MaxTunnels returns the maximum number of allowed tunnels.
func (tl *TunnelLimiter) MaxTunnels() int64 {
	if tl == nil {
		return 0
	}
	return tl.maxTunnels
}
