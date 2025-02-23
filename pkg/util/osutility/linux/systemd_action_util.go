package linux

import (
	"context"
	"fmt"
	"go.uber.org/multierr"
	"time"
)

func CheckAndStartSystemdProcess(ctx context.Context, appName string,
	retryCount int, ou OSUtil) error {
	currentCount := 0
	appRunning, err := ou.Systemd().IsRunning(ctx, appName)
	if err != nil {
		err = multierr.Append(err, fmt.Errorf("%s status produced error", appName))
		return err
	}
	for !appRunning && currentCount < retryCount {
		currentCount += 1
		err := ou.Systemd().Start(ctx, appName)
		if err != nil {
			err = multierr.Append(err, fmt.Errorf("unable to start %s", appName))
			return err
		}
		time.Sleep(20 * time.Second)
		appRunning, err = ou.Systemd().IsRunning(ctx, appName)
		if err != nil {
			err = multierr.Append(err, fmt.Errorf("%s status produced error after start", appName))
			return err
		}
	}
	if !appRunning {
		err = fmt.Errorf("%s not started after 20 seconds and retry count %d", appName, retryCount)
		return err
	}
	return nil
}
