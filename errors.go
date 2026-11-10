package tcpportscan

import (
	"context"
	"errors"
	"net"
	"syscall"
)

func classifyError(ctx context.Context, err error) State {
	if ctx.Err() != nil {
		return StateCanceled
	}

	if errors.Is(err, syscall.ECONNREFUSED) {
		return StateClosed
	}

	if errors.Is(err, syscall.ENETUNREACH) ||
		errors.Is(err, syscall.EHOSTUNREACH) {
		return StateUnreachable
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return StateTimeout
	}

	return StateError
}
