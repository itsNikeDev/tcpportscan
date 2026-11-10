package tcpportscan

import (
	"context"
	"errors"
	"fmt"
	"net"
	"syscall"
	"testing"
)

func TestClassifyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want State
	}{
		{
			name: "connection refused",
			err:  syscall.ECONNREFUSED,
			want: StateClosed,
		},
		{
			name: "network unreachable",
			err:  syscall.ENETUNREACH,
			want: StateUnreachable,
		},
		{
			name: "host unreachable",
			err:  syscall.EHOSTUNREACH,
			want: StateUnreachable,
		},
		{
			name: "timeout",
			err:  &net.DNSError{IsTimeout: true},
			want: StateTimeout,
		},
		{
			name: "unknown error",
			err:  errors.New("something went wrong"),
			want: StateError,
		},
		{
			name: "wrapped connection refused",
			err:  fmt.Errorf("dial failed: %w", syscall.ECONNREFUSED),
			want: StateClosed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := classifyError(context.Background(), tt.err)

			if got != tt.want {
				t.Fatalf("expected %s, got %s", tt.want, got)
			}
		})
	}
}

func TestClassifyErrorCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	got := classifyError(ctx, errors.New("operation interrupted"))

	if got != StateCanceled {
		t.Fatalf("expected %s, got %s", StateCanceled, got)
	}
}
