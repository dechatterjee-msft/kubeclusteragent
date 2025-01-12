package log

import (
	"context"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFrom(t *testing.T) {
	timeformat := "02-01-2006 15:04:05.000 UTC"
	ctxWithLogger := WithLogger(context.Background(), &timeformat)
	ogLogger := From(ctxWithLogger)
	tests := []struct {
		name string
		ctx  context.Context
	}{
		{name: "withLogger", ctx: ctxWithLogger},
		{name: "withoutLogger", ctx: context.Background()},
		{name: "nil context", ctx: nil},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			logger := From(tc.ctx)
			if !reflect.DeepEqual(ogLogger, logger) {
				t.Error()
			}
		})
	}

}

func TestWithLogger(t *testing.T) {
	timeformat := "02-01-2006 15:04:05.000 UTC"
	ctx := context.Background()
	newCtx := WithLogger(ctx, &timeformat)

	require.NotNil(t, newCtx.Value(logKey))
}

func TestSingleton(t *testing.T) {
	timeformat := "02-01-2006 15:04:05.000 UTC"
	logger1 := newLogger(&timeformat)
	logger2 := newLogger(&timeformat)
	if logger1.IsZero() != logger2.IsZero() {
		t.Errorf("Expected instance1 and instance2 to be the same instance, but they were not")
	}

}
