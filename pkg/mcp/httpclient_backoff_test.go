package mcp

import (
	"testing"
	"time"
)

// TestReconnectBackoff verifies the exponential-with-cap schedule used to
// space out reconnects to a flapping SSE stream. Without this backoff, a
// downstream server that accepts and immediately drops (or rejects, e.g. with
// 429) the event stream drives a tight reconnect loop that leaks resources on
// every cycle.
func TestReconnectBackoff(t *testing.T) {
	tests := []struct {
		name     string
		failures int
		want     time.Duration
	}{
		{"zero yields no delay", 0, 0},
		{"negative yields no delay", -5, 0},
		{"first failure is base", 1, baseReconnectBackoff},         // 500ms
		{"second failure doubles", 2, 2 * baseReconnectBackoff},    // 1s
		{"third failure quadruples", 3, 4 * baseReconnectBackoff},  // 2s
		{"grows exponentially", 5, 16 * baseReconnectBackoff},      // 8s
		{"last value under the cap", 6, 32 * baseReconnectBackoff}, // 16s
		{"first value hitting the cap", 7, maxReconnectBackoff},    // 32s -> capped to 30s
		{"is capped at max", 30, maxReconnectBackoff},
		{"huge count stays capped (no overflow)", 1000, maxReconnectBackoff},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := reconnectBackoff(tt.failures); got != tt.want {
				t.Fatalf("reconnectBackoff(%d) = %v, want %v", tt.failures, got, tt.want)
			}
		})
	}
}

// TestReconnectBackoffMonotonicAndBounded asserts the schedule never decreases
// and never exceeds the cap for any positive count.
func TestReconnectBackoffMonotonicAndBounded(t *testing.T) {
	var prev time.Duration
	for n := 1; n <= 64; n++ {
		got := reconnectBackoff(n)
		if got < prev {
			t.Fatalf("reconnectBackoff(%d)=%v decreased from previous %v", n, got, prev)
		}
		if got > maxReconnectBackoff {
			t.Fatalf("reconnectBackoff(%d)=%v exceeds cap %v", n, got, maxReconnectBackoff)
		}
		if got <= 0 {
			t.Fatalf("reconnectBackoff(%d)=%v is non-positive for a positive count", n, got)
		}
		prev = got
	}
}

// TestStreamFlapped verifies the decision that separates a healthy stream
// (delivered a message, or stayed up long enough) from a flapping one that
// should trigger reconnect backoff.
func TestStreamFlapped(t *testing.T) {
	tests := []struct {
		name             string
		messagesReceived int
		streamDuration   time.Duration
		want             bool
	}{
		{"short and silent flaps", 0, 100 * time.Millisecond, true},
		{"just under threshold, silent, flaps", 0, minStableStreamDuration - time.Millisecond, true},
		{"at threshold does not flap", 0, minStableStreamDuration, false},
		{"long and silent does not flap", 0, minStableStreamDuration + time.Second, false},
		{"short but delivered a message does not flap", 1, 10 * time.Millisecond, false},
		{"delivered messages never flaps", 5, time.Nanosecond, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := streamFlapped(tt.messagesReceived, tt.streamDuration); got != tt.want {
				t.Fatalf("streamFlapped(%d, %v) = %v, want %v",
					tt.messagesReceived, tt.streamDuration, got, tt.want)
			}
		})
	}
}
