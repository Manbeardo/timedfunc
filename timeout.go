package timeout

import (
	"context"
	"errors"
	"time"
)

// Func runs fn with a context that is canceled with context.DeadlineExceeded if fn does not return within the duration specified by d.
// This function is not guaranteed to return within the specified duration because it allows fn to handle cancellation gracefully.
func Func[T any](
	ctx context.Context,
	d time.Duration,
	fn func(context.Context) (T, error),
) (T, error) {
	ctx, timer := Context(ctx, d)
	defer timer.Stop()
	return fn(ctx)
}

var ErrChannelClosed = errors.New("channel is closed")

// Chan listens for a message from c.
// If d expires before receiving a message from c, context.DeadlineExceeded is returned.
// If ctx is cancelled before receiving a message from c, the cancellation cause is returned.
// If c is closed before receiving a message from c, ErrChannelClosed is returned.
func Chan[T any](
	ctx context.Context,
	d time.Duration,
	c <-chan T,
) (T, error) {
	ctx, timer := Context(ctx, d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return *new(T), context.Cause(ctx)
	case t, ok := <-c:
		if !ok {
			return t, ErrChannelClosed
		}
		return t, nil
	}
}

// Context creates a child context of ctx that will be cancelled after d. It also returns the timer that
// will issue the cancellation so that the cancellation of the returned context can be prevented.
func Context(
	ctx context.Context,
	d time.Duration,
) (context.Context, *time.Timer) {
	ctx, cancel := context.WithCancelCause(ctx)
	timer := time.AfterFunc(d, func() {
		cancel(context.DeadlineExceeded)
	})
	return ctx, timer
}
