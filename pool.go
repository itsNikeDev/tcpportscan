package tcpportscan

import (
	"context"
	"sync"
	"time"
)

func startWorkers(
	ctx context.Context,
	count int,
	jobs <-chan job,
	results chan<- Result,
	timeout time.Duration,
	check portChecker,
) *sync.WaitGroup {
	var wg sync.WaitGroup

	wg.Add(count)

	for i := 0; i < count; i++ {
		go func() {
			defer wg.Done()

			worker(ctx, jobs, results, timeout, check)
		}()
	}

	return &wg
}
