package tcpportscan

import (
	"testing"
	"time"
)

func TestNewDefaults(t *testing.T) {
	s, err := New()
	if err != nil {
		t.Fatal(err)
	}

	if s.concurrency != 100 {
		t.Fatalf("expected concurrency 100, got %d", s.concurrency)
	}

	if s.timeout != time.Second {
		t.Fatalf("expected timeout 1s, got %v", s.timeout)
	}
}

func TestNewOptions(t *testing.T) {
	s, err := New(
		WithConcurrency(50),
		WithConnectTimeout(2*time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}

	if s.concurrency != 50 {
		t.Fatalf("expected concurrency 50, got %d", s.concurrency)
	}

	if s.timeout != 2*time.Second {
		t.Fatalf("expected timeout 2s, got %v", s.timeout)
	}
}

func TestNewWithResolver(t *testing.T) {
	resolver := &mockResolver{}

	s, err := New(WithResolver(resolver))
	if err != nil {
		t.Fatal(err)
	}

	if s.resolver != resolver {
		t.Fatal("expected custom resolver")
	}
}

func TestNewInvalidOptions(t *testing.T) {
	tests := []struct {
		name string
		opt  Option
	}{
		{"zero concurrency", WithConcurrency(0)},
		{"negative concurrency", WithConcurrency(-1)},
		{"zero timeout", WithConnectTimeout(0)},
		{"negative timeout", WithConnectTimeout(-time.Second)},
		{"nil resolver", WithResolver(nil)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := New(tt.opt)
			if err == nil {
				t.Fatal("expected error")
			}
		})
	}
}
