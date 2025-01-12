package grpc

import (
	"context"
	"errors"
	"fmt"
	"kubeclusteragent/pkg/util/log/log"
	"net"
	"net/http"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/rs/cors"
	"github.com/tmc/grpc-websocket-proxy/wsproxy"
	"google.golang.org/grpc"
)

var (
	// defaultGatewayShutdownTimeout is the default gateway shutdown timeout.
	defaultGatewayShutdownTimeout = time.Second * 3
)

// Endpoint is an endpoint function.
type Endpoint func(ctx context.Context, mux *runtime.ServeMux, endpoint string, opts []grpc.DialOption) (err error)

// GatewayConfig is configuration Gateway.
type GatewayConfig struct {
	// ServerAddr is the GRPC address.
	ServerAddr string
	// HTTPAddr is the address the HTTP server listens on.
	HTTPAddr string
	// Endpoints is a slice of endpoint functions.
	Endpoints []Endpoint
	// EnableWebsockets enables websocket support.
	EnableWebsockets bool
}

// Gateway is a GRPC HTTP gateway.
type Gateway struct {
	name            string
	config          GatewayConfig
	grpcDialOptions []grpc.DialOption
	shutdownTimeout time.Duration
}

// NewGateway creates an instance of Gateway.
func NewGateway(name string, config GatewayConfig, grpcDialOptions []grpc.DialOption) *Gateway {
	g := &Gateway{
		name:            name,
		config:          config,
		grpcDialOptions: grpcDialOptions,
		shutdownTimeout: defaultGatewayShutdownTimeout,
	}

	return g
}

// Start starts the gateway. It returns a channel that, when closed, stops the gateway.
func (g *Gateway) Start(ctx context.Context, serverCertFilePath, serverKeyFilePath string, options ...runtime.ServeMuxOption) (<-chan struct{}, error) {
	logger := log.From(ctx).WithName(g.name)
	ctx = log.WithExistingLogger(ctx, logger)
	handler, err := g.createHandler(ctx, options...)
	if err != nil {
		return nil, fmt.Errorf("create handler for gateway: %w", err)
	}

	// cer, err := tls.LoadX509KeyPair(serverCertFilePath, serverKeyFilePath)
	// if err != nil {
	//	return nil, fmt.Errorf("unable to load TLS certificates: %w", err)
	// }
	// config := &tls.Config{
	//	Certificates: []tls.Certificate{cer},
	//	MinVersion:   tls.VersionTLS12,
	// }

	httpServer := &http.Server{
		Addr:        g.config.HTTPAddr,
		Handler:     handler,
		BaseContext: func(_ net.Listener) context.Context { return ctx },
		//	TLSConfig:         config,
		ReadHeaderTimeout: 3 * time.Second,
	}
	ch := make(chan struct{}, 1)
	go func() {
		logger.Info("Starting gRPC gateway", "addr", g.config.HTTPAddr)
		err := httpServer.ListenAndServe()
		if !errors.Is(err, http.ErrServerClosed) {
			logger.Error(err, "Failed to stop HTTP server for GRPC gateway cleanly")
		}
		logger.Info("gRPC gateway has stopped")
	}()

	go func() {
		<-ctx.Done()
		logger.Info("Stopping gRPC gateway gracefully")

		gracefulCtx, cancelShutdown := context.WithTimeout(context.Background(), g.shutdownTimeout)
		defer cancelShutdown()

		if err := httpServer.Shutdown(gracefulCtx); err != nil {
			logger.Error(err, "Unable to shut GRPC gateway down cleanly")
		}

		close(ch)
	}()

	return ch, nil
}

func (g *Gateway) registerEndpoint(ctx context.Context, mux *runtime.ServeMux, fn Endpoint) error {
	return fn(ctx, mux, g.config.ServerAddr, g.grpcDialOptions)
}

func (g *Gateway) createHandler(ctx context.Context, options ...runtime.ServeMuxOption) (http.Handler, error) {
	logger := log.From(ctx)

	mux := runtime.NewServeMux(options...)

	for _, endpoint := range g.config.Endpoints {
		if err := g.registerEndpoint(ctx, mux, endpoint); err != nil {
			return nil, fmt.Errorf("register endpoint: %w", err)
		}
	}

	var h http.Handler = mux
	if g.config.EnableWebsockets {
		logger.Info("Enabling websockets")
		h = wsproxy.WebsocketProxy(h)
	}

	handler := cors.AllowAll().Handler(h)

	return handler, nil
}
