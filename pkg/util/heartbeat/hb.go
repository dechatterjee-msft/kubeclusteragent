package heartbeat

import (
	"context"
	"kubeclusteragent/pkg/util/log/log"
	"time"
)

func HeartBeat(ctx context.Context, f func() error, interval time.Duration, stopped chan struct{}, forceQuit chan bool) {
	logger := log.From(ctx)
	defer close(stopped)

	tick := time.NewTicker(interval)
	defer tick.Stop()

	for {
		select {
		case <-tick.C:
			err := f()
			if err != nil {
				logger.Info("Cluster snapshot for this heartbeat failed with error", err)
			}
		case <-ctx.Done():
			return
		case <-forceQuit:
			return
		}
	}
}

func HeartBeatWithCtx(ctx context.Context, f func(ctx2 context.Context) error, interval time.Duration, stopped chan struct{}, forceQuit chan bool) {
	logger := log.From(ctx)
	defer close(stopped)

	tick := time.NewTicker(interval)
	defer tick.Stop()

	var after <-chan time.Time

	for {
		select {
		case <-tick.C:
			err := f(ctx)
			if err != nil {
				logger.Info("Cluster snapshot for this heartbeat failed with error", err)
			}
		case <-ctx.Done():
		case <-forceQuit:
		case <-after:
			logger.Info("Stopping reconciliation")
			return
		}
	}
}
