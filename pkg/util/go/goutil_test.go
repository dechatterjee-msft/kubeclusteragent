package _go

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestWaitForChannelsToClose(t *testing.T) {
	tests := []struct {
		name string
		fn   func() chan struct{}
		want bool
	}{
		{
			name: "all close",
			fn: func() chan struct{} {
				ch := make(chan struct{})
				go func() {
					time.AfterFunc(50*time.Millisecond, func() {
						close(ch)
					})
				}()

				return ch
			},
			want: true,
		},
		{
			name: "all do not close",
			fn: func() chan struct{} {
				ch := make(chan struct{})
				go func() {
					time.AfterFunc(150*time.Millisecond, func() {
						close(ch)
					})
				}()

				return ch
			},
			want: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()

			ch := tc.fn()
			require.Equal(t, tc.want, WaitForChannelsToClose(ctx, ch))
		})
	}
}
