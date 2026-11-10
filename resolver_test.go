package tcpportscan

import (
	"context"
	"errors"
	"net"
	"testing"
)

type mockResolver struct {
	ips []net.IP
	err error
}

func (m *mockResolver) LookupIP(
	ctx context.Context,
	host string,
) ([]net.IP, error) {
	return m.ips, m.err
}

func TestResolveHostIP(t *testing.T) {
	resolver := &mockResolver{}

	ips, err := resolveHost(
		context.Background(),
		resolver,
		"127.0.0.1",
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(ips) != 1 || !ips[0].Equal(net.ParseIP("127.0.0.1")) {
		t.Fatalf("unexpected IPs: %v", ips)
	}
}

func TestResolveHostDNS(t *testing.T) {
	expected := []net.IP{
		net.ParseIP("127.0.0.1"),
		net.ParseIP("::1"),
	}

	resolver := &mockResolver{ips: expected}

	ips, err := resolveHost(
		context.Background(),
		resolver,
		"localhost",
	)
	if err != nil {
		t.Fatal(err)
	}

	if len(ips) != 2 {
		t.Fatalf("expected 2 IPs, got %d", len(ips))
	}
}

func TestResolveHostError(t *testing.T) {
	resolver := &mockResolver{
		err: errors.New("DNS failure"),
	}

	_, err := resolveHost(
		context.Background(),
		resolver,
		"example.invalid",
	)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestResolveHostEmpty(t *testing.T) {
	resolver := &mockResolver{}

	_, err := resolveHost(
		context.Background(),
		resolver,
		"",
	)
	if err == nil {
		t.Fatal("expected error")
	}
}
