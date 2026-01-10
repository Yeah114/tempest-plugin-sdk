package protocol

import (
	"context"
	"time"
)

func timeoutMsFromCtx(ctx context.Context) int64 {
	if ctx == nil {
		return 0
	}
	deadline, ok := ctx.Deadline()
	if !ok {
		return 0
	}
	d := time.Until(deadline)
	if d <= 0 {
		return 1
	}
	return int64(d / time.Millisecond)
}
