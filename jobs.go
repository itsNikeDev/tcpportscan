package tcpportscan

import (
	"context"
	"fmt"
	"net"
	"strings"
)

func buildJobs(
	ctx context.Context,
	resolver Resolver,
	hosts []string,
	ports []uint16,
) ([]job, error) {
	var jobs []job

	seenHosts := make(map[string]struct{})

	for _, host := range hosts {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if host == "" {
			return nil, fmt.Errorf("host is empty")
		}

		hostKey := strings.ToLower(host)

		if _, exists := seenHosts[hostKey]; exists {
			continue
		}

		seenHosts[hostKey] = struct{}{}

		ips, err := resolveHost(ctx, resolver, host)
		if err != nil {
			return nil, err
		}

		if len(ips) == 0 {
			return nil, fmt.Errorf("host %q resolved to no IP addresses", host)
		}

		seenIPs := make(map[string]struct{})

		for _, ip := range ips {
			if ip == nil {
				continue
			}

			key := ip.String()
			if _, exists := seenIPs[key]; exists {
				continue
			}
			seenIPs[key] = struct{}{}

			for _, port := range ports {
				jobs = append(jobs, job{
					host: host,
					ip:   append(net.IP(nil), ip...),
					port: port,
				})
			}
		}
	}

	return jobs, nil
}
