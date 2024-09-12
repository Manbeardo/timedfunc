package timeout

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFunc(t *testing.T) {
	t.Run("does not cancel the context if fn returns in time", func(t *testing.T) {
		var ctxFromFunc context.Context
		v, err := Func(
			context.Background(),
			10*time.Millisecond,
			func(ctx context.Context) (any, error) {
				ctxFromFunc = ctx
				return "foo", nil
			})
		assert.NoError(t, err)
		assert.Equal(t, "foo", v)

		time.Sleep(15 * time.Millisecond)

		assert.NoError(t, ctxFromFunc.Err())
	})

	t.Run("does not cancel the context if fn panics", func(t *testing.T) {
		var ctxFromFunc context.Context
		func() {
			defer func() {
				_ = recover()
			}()
			_, _ = Func(
				context.Background(),
				10*time.Millisecond,
				func(ctx context.Context) (any, error) {
					ctxFromFunc = ctx
					panic("foo")
				})
		}()
		time.Sleep(15 * time.Millisecond)

		assert.NoError(t, ctxFromFunc.Err())
	})

	t.Run("cancels the context with context.DeadlineExceeded if fn does not return in time", func(t *testing.T) {
		_, err := Func(
			context.Background(),
			10*time.Millisecond,
			func(ctx context.Context) (any, error) {
				select {
				case <-ctx.Done():
					return nil, context.Cause(ctx)
				case <-time.After(15 * time.Millisecond):
					return nil, nil
				}
			},
		)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})
}

func TestChan(t *testing.T) {
	t.Run("returns the context error cause if ctx is cancelled before a message is received", func(t *testing.T) {
		c := make(chan any)
		go func() {
			time.Sleep(15 * time.Millisecond)
			c <- nil
		}()
		cause := errors.New("cause")
		ctx, cancel := context.WithCancelCause(context.Background())
		cancel(cause)
		_, err := Chan(ctx, 10*time.Millisecond, c)
		assert.ErrorIs(t, err, cause)
	})

	t.Run("returns context.DeadlineExceeded if a message is not received within d", func(t *testing.T) {
		c := make(chan any)
		go func() {
			time.Sleep(15 * time.Millisecond)
			c <- nil
		}()
		_, err := Chan(context.Background(), 10*time.Millisecond, c)
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	})

	t.Run("returns ErrChannelClosed if c is closed before a message is received", func(t *testing.T) {
		c := make(chan any)
		close(c)
		_, err := Chan(context.Background(), 10*time.Millisecond, c)
		assert.ErrorIs(t, err, ErrChannelClosed)
	})

	t.Run("returns the message if it is received", func(t *testing.T) {
		c := make(chan string)
		go func() {
			c <- "foo"
		}()
		v, err := Chan(context.Background(), 10*time.Millisecond, c)
		assert.NoError(t, err)
		assert.Equal(t, "foo", v)
	})
}
