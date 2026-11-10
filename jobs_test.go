package tcpportscan

import (
	"context"
	"errors"
	"net"
	"testing"
)

func TestBuildJobs(t *testing.T) {
	resolver := &mockResolver{
		ips: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
			net.ParseIP("127.0.0.1"),
		},
	}

	jobs, err := buildJobs(
		context.Background(),
		resolver,
		[]string{"example.com", "example.com"},
		[]uint16{80, 443},
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 4 {
		t.Fatalf("expected 4 jobs, got %d", len(jobs))
	}
}

func TestBuildJobsNoIPs(t *testing.T) {
	resolver := &mockResolver{
		ips: nil,
	}

	_, err := buildJobs(
		context.Background(),
		resolver,
		[]string{"example.com"},
		[]uint16{80},
	)

	if err == nil {
		t.Fatal("expected error when resolver returns no IPs")
	}
}

func TestBuildJobsEmptyHost(t *testing.T) {
	resolver := &mockResolver{}

	_, err := buildJobs(
		context.Background(),
		resolver,
		[]string{""},
		[]uint16{80},
	)

	if err == nil {
		t.Fatal("expected error for empty host")
	}
}

func TestBuildJobsDuplicateHosts(t *testing.T) {
	resolver := &mockResolver{
		ips: []net.IP{
			net.ParseIP("127.0.0.1"),
		},
	}

	jobs, err := buildJobs(
		context.Background(),
		resolver,
		[]string{"example.com", "example.com"},
		[]uint16{80, 443},
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 2 {
		t.Fatalf("got %d jobs, want 2", len(jobs))
	}
}

func TestBuildJobsDuplicateIPs(t *testing.T) {
	resolver := &mockResolver{
		ips: []net.IP{
			net.ParseIP("127.0.0.1"),
			net.ParseIP("127.0.0.1"),
			net.ParseIP("::1"),
			net.ParseIP("::1"),
		},
	}

	jobs, err := buildJobs(
		context.Background(),
		resolver,
		[]string{"example.com"},
		[]uint16{80, 443},
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 4 {
		t.Fatalf("got %d jobs, want 4", len(jobs))
	}
}

func TestBuildJobsCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := buildJobs(
		ctx,
		&mockResolver{},
		[]string{"127.0.0.1"},
		[]uint16{80},
	)

	if !errors.Is(err, context.Canceled) {
		t.Fatalf("got error %v, want context.Canceled", err)
	}
}

func TestBuildJobsResolverError(t *testing.T) {
	expectedErr := errors.New("dns lookup failed")

	resolver := &mockResolver{
		err: expectedErr,
	}

	_, err := buildJobs(
		context.Background(),
		resolver,
		[]string{"example.com"},
		[]uint16{80},
	)

	if !errors.Is(err, expectedErr) {
		t.Fatalf("got error %v, want %v", err, expectedErr)
	}
}

func TestBuildJobsHostsCaseInsensitive(t *testing.T) {
	resolver := &mockResolver{
		ips: []net.IP{net.ParseIP("127.0.0.1")},
	}

	jobs, err := buildJobs(
		context.Background(),
		resolver,
		[]string{"example.com", "EXAMPLE.COM"},
		[]uint16{80},
	)

	if err != nil {
		t.Fatal(err)
	}

	if len(jobs) != 1 {
		t.Fatalf("got %d jobs, want 1", len(jobs))
	}
}

func TestBuildJobsCopiesIPs(t *testing.T) {
	originalIP := net.ParseIP("127.0.0.1")

	resolver := &mockResolver{
		ips: []net.IP{originalIP},
	}

	jobs, err := buildJobs(
		context.Background(),
		resolver,
		[]string{"example.com"},
		[]uint16{80},
	)
	if err != nil {
		t.Fatal(err)
	}

	originalIP[0] = 192

	if jobs[0].ip.String() != "127.0.0.1" {
		t.Fatalf("IP was modified: %s", jobs[0].ip)
	}
}
