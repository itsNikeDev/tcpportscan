package tcpportscan

import (
	"context"
	"fmt"
	"net"
)

// Resolver резолвит имя хоста в IP-адреса. Его реализует *net.Resolver.
// Свою реализацию можно внедрить через WithResolver, например в тестах.
type Resolver interface {
	LookupIP(ctx context.Context, host string) ([]net.IP, error)
}

type netResolver struct {
	resolver *net.Resolver
}

func newNetResolver() *netResolver {
	return &netResolver{
		resolver: net.DefaultResolver,
	}
}

func (r *netResolver) LookupIP(
	ctx context.Context,
	host string,
) ([]net.IP, error) {
	ips, err := r.resolver.LookupIP(ctx, "ip", host)
	if err != nil {
		return nil, fmt.Errorf("resolve host %q: %w", host, err)
	}

	return ips, nil
}

func resolveHost(
	ctx context.Context,
	resolver Resolver,
	host string,
) ([]net.IP, error) {
	if host == "" {
		return nil, fmt.Errorf("host is empty")
	}

	if ip := net.ParseIP(host); ip != nil {
		return []net.IP{ip}, nil
	}

	return resolver.LookupIP(ctx, host)
}
