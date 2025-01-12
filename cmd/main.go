package main

import (
	"context"
	"flag"
	"kubeclusteragent/pkg/agent"
	flagutil "kubeclusteragent/pkg/util/flag"
	"kubeclusteragent/pkg/util/log/log"
	"os"
)

func main() {
	var config agent.Config
	var timeformat *string
	flagutil.EnvStringVar(&config.GRPCAddr, "AGENT_GRPC_ADDR", "grpc-addr", "0.0.0.0:50055", "gRPC server address")
	flagutil.EnvStringVar(&config.ServerAddr, "AGENT_SERVER_ADDR", "server-addr", "0.0.0.0:8080", "HTTP server address")
	flagutil.EnvBoolVar(&config.DryRun, "AGENT_DRY_RUN", "dry-run", false, "Run in dry run mode")
	flagutil.EnvStringVar(&config.TokenSharedKey, "TOKEN_SHARED_KEY", "secret-key", "", "Secret key for token verification")
	flagutil.EnvStringVar(&config.ServerCertFilePath, "SERVER_CERT", "server-cert", "", "Server cert for tls")
	flagutil.EnvStringVar(&config.ServerKeyFilePath, "SERVER_KEY", "server-key", "", "Server key for tls")
	flagutil.EnvStringVar(&config.CACertFilePath, "CA_CERT", "ca-cert", "", "CA cert for tls")
	timeformat = flag.String("format", "02-01-2006 15:04:05.000 UTC", "time format")
	flag.Parse()
	ctx := log.WithLogger(context.Background(), timeformat)
	if err := run(ctx, config); err != nil {
		logger := log.From(ctx)
		logger.Error(err, "service failed")
		os.Exit(1)
	}
}
func run(ctx context.Context, config agent.Config) error {
	app := agent.New(config)
	return app.Start(ctx)
}
