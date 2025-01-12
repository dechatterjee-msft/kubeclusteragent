package _go

import (
	"context"
	"github.com/magiconair/properties"
	"kubeclusteragent/pkg/util/log/log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func GetPropfile(propFileLocation string) (map[string]string, error) {
	file, err := properties.LoadFile(propFileLocation, properties.UTF8)
	if err != nil {
		return nil, err
	}
	return file.Map(), nil
}

// WaitForChannelsToClose waits for channels to close. If all the channels do not
// close before the context is done, false will be returned.
func WaitForChannelsToClose(ctx context.Context, chans ...<-chan struct{}) bool {
	done := make(chan struct{}, 1)
	go func() {
		for _, v := range chans {
			<-v
		}
		close(done)
	}()

	select {
	case <-ctx.Done():
		return false
	case <-done:
		return true
	}
}

// HandleGracefulClose gracefully handles shutting down the process.
func HandleGracefulClose(ctx context.Context, cancel context.CancelFunc, chans ...<-chan struct{}) {
	logger := log.From(ctx).WithName("graceful")

	signalChan := make(chan os.Signal, 1)

	signal.Notify(
		signalChan,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)

	<-signalChan
	logger.Info("Shutting down gracefully")

	closeCtx, closeCancel := context.WithTimeout(ctx, 5*time.Second)
	defer closeCancel()

	go func() {
		<-signalChan
		logger.Info("Terminating")
		closeCancel()
	}()

	cancel()

	logger.Info("Waiting for servers to stop")

	if !WaitForChannelsToClose(closeCtx, chans...) {
		logger.Info("All channels were not closed")
	}
	logger.Info("Exiting normally")
}
