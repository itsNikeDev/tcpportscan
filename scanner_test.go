package tcpportscan

import (
	"context"
	"errors"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCheckPortOpen(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	addr := listener.Addr().(*net.TCPAddr)

	result := checkPort(
		context.Background(),
		"localhost",
		net.ParseIP("127.0.0.1"),
		uint16(addr.Port),
		time.Second,
	)

	if result.State != StateOpen {
		t.Fatalf("expected open, got %s, err: %v", result.State, result.Err)
	}

	if result.Duration <= 0 {
		t.Fatal("expected positive duration")
	}

	if result.IP == nil {
		t.Fatal("expected IP")
	}
}

func TestScanOpenPort(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	port := uint16(listener.Addr().(*net.TCPAddr).Port)

	scanner, err := New(
		WithConcurrency(2),
		WithConnectTimeout(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	results, err := scanner.Scan(
		ctx,
		[]string{"127.0.0.1"},
		[]uint16{port},
	)
	if err != nil {
		t.Fatal(err)
	}

	count := 0

	for result := range results {
		count++

		if result.State != StateOpen {
			t.Errorf("expected open, got %s: %v", result.State, result.Err)
		}

		if result.Port != port {
			t.Errorf("expected port %d, got %d", port, result.Port)
		}
	}

	if count != 1 {
		t.Fatalf("expected 1 result, got %d", count)
	}
}

func TestScanCanceled(t *testing.T) {
	scanner, err := New(
		WithConcurrency(2),
		WithConnectTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	results, err := scanner.Scan(
		ctx,
		[]string{"192.0.2.1"},
		[]uint16{80, 81, 82, 83, 84},
	)
	if err != nil {
		t.Fatal(err)
	}

	cancel()

	for range results {
	}

}

func TestScanSlowConsumer(t *testing.T) {
	const listenerCount = 5

	var listeners []net.Listener
	var ports []uint16

	for i := 0; i < listenerCount; i++ {
		listener, err := net.Listen("tcp", "127.0.0.1:0")
		if err != nil {
			t.Fatal(err)
		}

		listeners = append(listeners, listener)
		ports = append(
			ports,
			uint16(listener.Addr().(*net.TCPAddr).Port),
		)
	}

	defer func() {
		for _, listener := range listeners {
			listener.Close()
		}
	}()

	scanner, err := New(
		WithConcurrency(2),
		WithConnectTimeout(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}

	results, err := scanner.Scan(
		context.Background(),
		[]string{"127.0.0.1"},
		ports,
	)
	if err != nil {
		t.Fatal(err)
	}

	// Имитируем медленного потребителя.
	time.Sleep(200 * time.Millisecond)

	count := 0

	for result := range results {
		if result.State != StateOpen {
			t.Errorf("expected open, got %s", result.State)
		}
		count++
	}

	if count != listenerCount {
		t.Fatalf("expected %d results, got %d", listenerCount, count)
	}
}

func TestScannerConcurrentUse(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer listener.Close()

	port := uint16(listener.Addr().(*net.TCPAddr).Port)

	scanner, err := New(
		WithConcurrency(2),
		WithConnectTimeout(time.Second),
	)
	if err != nil {
		t.Fatal(err)
	}

	var wg sync.WaitGroup
	errCh := make(chan error, 2)

	for i := 0; i < 2; i++ {
		wg.Add(1)

		go func() {
			defer wg.Done()

			results, err := scanner.Scan(
				context.Background(),
				[]string{"127.0.0.1"},
				[]uint16{port},
			)
			if err != nil {
				errCh <- err
				return
			}

			count := 0
			for result := range results {
				count++

				if result.State != StateOpen {
					errCh <- fmt.Errorf(
						"expected open, got %s",
						result.State,
					)
					return
				}
			}

			if count != 1 {
				errCh <- fmt.Errorf("expected 1 result, got %d", count)
			}
		}()
	}

	wg.Wait()
	close(errCh)

	for err := range errCh {
		t.Error(err)
	}
}

type staticResolver struct {
	ips []net.IP
}

func (r *staticResolver) LookupIP(
	ctx context.Context,
	host string,
) ([]net.IP, error) {
	return r.ips, nil
}

func TestScanConcurrencyLimit(t *testing.T) {
	const (
		concurrency = 2
		totalJobs   = 5
	)

	scanner, err := New(WithConcurrency(concurrency))
	if err != nil {
		t.Fatal(err)
	}

	jobsCh := make(chan job, totalJobs)
	resultsCh := make(chan Result, totalJobs)

	for i := 0; i < totalJobs; i++ {
		jobsCh <- job{
			host: "localhost",
			ip:   net.ParseIP("127.0.0.1"),
			port: uint16(8000 + i),
		}
	}
	close(jobsCh)

	var active atomic.Int32
	var maxActive atomic.Int32

	fakeCheck := func(
		ctx context.Context,
		host string,
		ip net.IP,
		port uint16,
		timeout time.Duration,
	) Result {
		current := active.Add(1)
		defer active.Add(-1)

		for {
			previous := maxActive.Load()
			if current <= previous ||
				maxActive.CompareAndSwap(previous, current) {
				break
			}
		}

		time.Sleep(100 * time.Millisecond)

		return Result{
			Host:  host,
			IP:    ip,
			Port:  port,
			State: StateOpen,
		}
	}

	wg := startWorkers(
		context.Background(),
		scanner.concurrency,
		jobsCh,
		resultsCh,
		scanner.timeout,
		fakeCheck,
	)

	wg.Wait()
	close(resultsCh)

	count := 0
	for range resultsCh {
		count++
	}

	if count != totalJobs {
		t.Fatalf("expected %d results, got %d", totalJobs, count)
	}

	if got := maxActive.Load(); got > int32(concurrency) {
		t.Errorf("max concurrency exceeded: got %d, limit %d", got, concurrency)
	}

	if got := maxActive.Load(); got != int32(concurrency) {
		t.Errorf("expected max concurrency %d, got %d", concurrency, got)
	}
}

func TestScanAlreadyCanceled(t *testing.T) {
	scanner, err := New()
	if err != nil {
		t.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	results, err := scanner.Scan(ctx, []string{"127.0.0.1"}, []uint16{80})

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want context.Canceled", err)
	}

	if results != nil {
		t.Fatal("expected nil results channel")
	}
}

func TestScanInvalidInputs(t *testing.T) {
	scanner, err := New()
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name  string
		ctx   context.Context
		hosts []string
		ports []uint16
	}{
		{
			name:  "nil context",
			ctx:   nil,
			hosts: []string{"127.0.0.1"},
			ports: []uint16{80},
		},
		{
			name:  "empty hosts",
			ctx:   context.Background(),
			hosts: nil,
			ports: []uint16{80},
		},
		{
			name:  "empty ports",
			ctx:   context.Background(),
			hosts: []string{"127.0.0.1"},
			ports: nil,
		},
		{
			name:  "zero port",
			ctx:   context.Background(),
			hosts: []string{"127.0.0.1"},
			ports: []uint16{0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := scanner.Scan(tt.ctx, tt.hosts, tt.ports)

			if err == nil {
				t.Fatal("expected error")
			}

			if results != nil {
				t.Fatal("expected nil results channel")
			}
		})
	}
}

func TestCheckPortClosed(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}

	port := uint16(listener.Addr().(*net.TCPAddr).Port)

	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}

	result := checkPort(
		context.Background(),
		"127.0.0.1",
		net.ParseIP("127.0.0.1"),
		port,
		time.Second,
	)

	if result.State != StateClosed {
		t.Fatalf(
			"got state %q, want %q; error: %v",
			result.State,
			StateClosed,
			result.Err,
		)
	}
}

func TestCheckPortCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	result := checkPort(
		ctx,
		"127.0.0.1",
		net.ParseIP("127.0.0.1"),
		80,
		time.Second,
	)

	if result.State != StateCanceled {
		t.Fatalf(
			"got state %q, want %q",
			result.State,
			StateCanceled,
		)
	}

	if !errors.Is(result.Err, context.Canceled) {
		t.Fatalf(
			"got error %v, want context.Canceled",
			result.Err,
		)
	}
}
