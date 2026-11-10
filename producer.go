package tcpportscan

import "context"

func produceJobs(
	ctx context.Context,
	jobs []job,
	jobsCh chan<- job,
) {
	defer close(jobsCh)

	for _, j := range jobs {
		select {
		case <-ctx.Done():
			return
		case jobsCh <- j:
		}
	}
}
