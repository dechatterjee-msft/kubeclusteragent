package grpc

import (
	"context"
	"errors"
	"fmt"
	grpc_prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/multierr"
	"google.golang.org/grpc"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/metrcis"
	"net"
	"net/http"
	"time"
)

// RegisterFn is a function that takes a GRPC server as input.
type RegisterFn func(s *grpc.Server) error

// ServerConfig is configuration for GRPCServer.
type ServerConfig struct {
	// Listener is where the server will listen.
	Listener net.Listener

	// RegisterFunc is a function that allows you to register servers.
	RegisterFunc RegisterFn
}

// Validate validates the ServerConfig. If not valid, an error is returned.
func (config *ServerConfig) Validate() error {
	var err error

	if config.Listener == nil {
		err = multierr.Append(err, errors.New("listener is required"))
	}

	if config.RegisterFunc == nil {
		err = multierr.Append(err, errors.New("register function is required"))
	}

	return err
}

// Server provides a GRPC server
type Server struct {
	config ServerConfig
	name   string
}

// NewServer creates an instance of Server.
func NewServer(name string, config ServerConfig) (*Server, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	server := &Server{
		name:   name,
		config: config,
	}

	return server, nil
}

// StartWithMetricsServer Start starts the GRPC server with metrics server
func (server *Server) StartWithMetricsServer(ctx context.Context, serverCertFilePath, serverKeyFilePath string, tokenSharedKey string, options ...grpc.ServerOption) (<-chan struct{}, error) {
	logger := log.From(ctx).WithName(server.name)
	s := grpc.NewServer(options...)
	grpcMetrics := grpc_prometheus.NewServerMetrics()
	grpcMetrics.InitializeMetrics(s)
	grpcMetrics.EnableHandlingTimeHistogram()
	prometheusRegistry := metrcis.MetricsRegistry()
	prometheusRegistry.MustRegister(grpcMetrics)
	if err := server.config.RegisterFunc(s); err != nil {
		return nil, fmt.Errorf("gRPC register: %w", err)
	}
	lis := server.config.Listener
	logger.Info("Starting gRPC server", "addr", lis.Addr().String())
	go func() {
		logger.Info("Starting metrics server", "addr", fmt.Sprintf("localhost:%v/api/v1alpha1/metrics", constants.MetricsServerPort))
		prometheusHTTPHandler := promhttp.HandlerFor(
			prometheusRegistry, promhttp.HandlerOpts{
				Registry: prometheusRegistry,
			})
		// http.Handle("/metrics", authutils.AuthWrapperHandler(tokenSharedKey, prometheusHTTPHandler))
		// if err := http.ListenAndServeTLS(constants.MetricsServerPort, serverCertFilePath, serverKeyFilePath, nil); err != nil {
		//	logger.Error(err, "Unable to start metrics server")
		// }
		http.Handle("/api/v1alpha1/metrics", prometheusHTTPHandler)
		metricsServer := &http.Server{
			Addr:              "localhost:31800",
			Handler:           prometheusHTTPHandler,
			ReadHeaderTimeout: 1 * time.Second,
		}
		err := metricsServer.ListenAndServe()
		if err != nil {
			logger.Error(err, "Unable to start metrics server")
		}
	}()
	grpc_prometheus.Register(s)
	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error(err, "Unable to stop gRPC server cleanly")
		}
		logger.Info("gRPC server has stopped")
	}()
	ch := make(chan struct{}, 1)
	go func() {
		<-ctx.Done()
		var clusterStatus cluster.Status = &cluster.LiveStatus{}
		currentClusterStatus := clusterStatus.GetStatus(ctx)
		logger.Info("currently agent phase", "phase", currentClusterStatus.Phase)
		if currentClusterStatus.Phase == constants.ClusterPhaseDeleting ||
			currentClusterStatus.Phase == constants.ClusterPhaseProvisioning ||
			currentClusterStatus.Phase == constants.ClusterPhaseUpgrading ||
			currentClusterStatus.Phase == constants.ClusterPhaseKubeConfigResetting {
			currentClusterStatus.Phase = constants.ClusterPhaseFailed
			logger.Info("agent is shutting down , updating current on going operation to ", "current phase", currentClusterStatus.Phase)
			clusterStatus.SetStatus(ctx, currentClusterStatus)
		}
		logger.Info("Stopping gRPC server gracefully")
		s.GracefulStop()
		close(ch)
	}()

	return ch, nil
}

// Start starts the GRPC server
func (server *Server) Start(ctx context.Context, options ...grpc.ServerOption) (<-chan struct{}, error) {
	logger := log.From(ctx).WithName(server.name)
	s := grpc.NewServer(options...)
	if err := server.config.RegisterFunc(s); err != nil {
		return nil, fmt.Errorf("gRPC register: %w", err)
	}
	lis := server.config.Listener
	logger.Info("Starting gRPC server", "addr", lis.Addr().String())

	go func() {
		if err := s.Serve(lis); err != nil {
			logger.Error(err, "Unable to stop gRPC server cleanly")
		}
		logger.Info("gRPC server has stopped")
	}()
	ch := make(chan struct{}, 1)
	go func() {
		<-ctx.Done()
		logger.Info("Stopping gRPC server gracefully")
		s.GracefulStop()
		close(ch)
	}()
	return ch, nil
}
