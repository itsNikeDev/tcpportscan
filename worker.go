package tcpportscan

import (
	"context"
	"net"
	"time"
)

type portChecker func(
	context.Context,
	string,
	net.IP,
	uint16,
	time.Duration,
) Result

func worker(
	ctx context.Context,
	jobs <-chan job,
	results chan<- Result,
	timeout time.Duration,
	check portChecker,
) {
	for {
		select {
		case <-ctx.Done():
			return

		case j, ok := <-jobs:
			if !ok {
				return
			}

			if ctx.Err() != nil {
				return
			}

			result := check(
				ctx,
				j.host,
				j.ip,
				j.port,
				timeout,
			)

			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}
}
