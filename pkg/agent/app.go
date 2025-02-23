package agent

import (
	"context"
	"fmt"
	prometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
	"kubeclusteragent/gen/go/agent/v1alpha1"
	"kubeclusteragent/pkg/cluster"
	"kubeclusteragent/pkg/constants"
	"kubeclusteragent/pkg/reconciler/certsreconciler"
	"kubeclusteragent/pkg/reconciler/statusreconciler"
	"kubeclusteragent/pkg/tools/patchtool"
	"kubeclusteragent/pkg/util/auth"
	"kubeclusteragent/pkg/util/go"
	grpcutil2 "kubeclusteragent/pkg/util/grpc"
	"kubeclusteragent/pkg/util/log/log"
	"kubeclusteragent/pkg/util/osutility/linux"
	"kubeclusteragent/pkg/util/reconcile"
	"net"
	"os"
)

type App struct {
	config Config
}

func New(config Config) *App {
	app := &App{
		config: config,
	}
	return app
}

func unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler) (resp interface{}, err error) {
	return handler(ctx, req)
}

func (a *App) Start(ctx context.Context) error {
	logger := log.From(ctx).WithName("App")
	if a.config.DryRun {
		logger = logger.WithValues("dryRun", true)
	}
	if err := os.MkdirAll(constants.ResourceDirectory, os.ModePerm); err != nil {
		logger.Error(err, "unable to create resource directory for cluster state store")
		return err
	}
	if err := os.MkdirAll(constants.CertsDirectory, os.ModePerm); err != nil {
		logger.Error(err, "unable to create resource directory for cluster certs")
		return err
	}
	logger.Info("Starting application")
	runCtx, runCancel := context.WithCancel(ctx)
	defer runCancel()
	handles, err := a.initGRPC(runCtx)
	if err != nil {
		return err
	}
	logger.Info("Application is started")
	_go.HandleGracefulClose(ctx, runCancel, handles...)
	logger.Info("Application has stopped")

	return nil
}

func (a *App) initGRPC(ctx context.Context) ([]<-chan struct{}, error) {
	grpcServerDone, err := a.startGRPC(ctx)
	if err != nil {
		return nil, fmt.Errorf("start gRPC server: %w", err)
	}
	grpcGatewayDone, err := a.startGateway(ctx)
	if err != nil {
		return nil, fmt.Errorf("start gRPC gateway server: %w", err)
	}
	handles := []<-chan struct{}{grpcServerDone, grpcGatewayDone}
	return handles, nil
}

func (a *App) startGRPC(ctx context.Context) (<-chan struct{}, error) {
	logger := log.From(ctx).WithName("App")
	var err error
	clusterStatus, err := cluster.NewLiveStatus(ctx, a.config.DryRun) // nolint
	if err != nil {
		return nil, fmt.Errorf("initialize cluster status: %w", err)
	}
	reconcilerRegistery, err := reconcile.NewReconcileManager()
	if err != nil {
		return nil, fmt.Errorf("start reconciler manager: %w", err)
	}

	jwtManager := *auth.CreateJwtManager(a.config.TokenSharedKey)
	// If current status of the Cluster is in any of the Intermediate states like Provisioning,Updating,Deleting automatically it will be marked as Failed on start-up
	currentClusterStatus := clusterStatus.GetStatus(ctx)
	if currentClusterStatus != nil && (currentClusterStatus.Phase == constants.ClusterPhaseProvisioning ||
		currentClusterStatus.Phase == constants.ClusterPhaseUpgrading ||
		currentClusterStatus.Phase == constants.ClusterPhaseDeleting ||
		currentClusterStatus.Phase == constants.ClusterPhaseKubeConfigResetting) {
		logger.Info("current cluster phase is an intermediate phase marking it as failed", "phase", currentClusterStatus.Phase)
		// TODO start the failed task once again
		currentClusterStatus.Phase = constants.ClusterPhaseFailed
		clusterStatus.SetStatus(ctx, currentClusterStatus)
	} else if currentClusterStatus != nil && currentClusterStatus.Phase == constants.ClusterPhaseProvisioned {
		// considering the scenario where agent got restarted or the system is rebooted
		// Sanity check on containerd and kubelet
		logger.Info("agent got restarted it will check the status of containerd and kubelet...")
		logger.Info("Agent started going next...")
		var ou = linux.New()
		err := linux.CheckAndStartSystemdProcess(ctx, "containerd", 3, ou)
		if err != nil {
			logger.Error(err, "containerd process failed to start")
			return nil, err
		}
		err = linux.CheckAndStartSystemdProcess(ctx, "kubelet", 3, ou)
		if err != nil {
			logger.Error(err, "kubelet process failed to start")
			return nil, err
		}
	}
	svc := NewLiveService(ctx, patchtool.NewClusterConfigInstallTool(clusterStatus, a.config.DryRun), jwtManager, reconcilerRegistery)
	// if kubeconfig is present, then register the reconcile to get the heart beat of the cluster
	if kc, err := svc.InstallTool.Config(ctx); err == nil {
		if svc.ReconcileRegistry.GetReconciler(statusreconciler.ClusterStatusReconcilerName) == nil && kc != nil {
			clusterStatusReconciler, err := statusreconciler.NewClusterStatusReconciler(ctx, string(kc))
			if err != nil {
				return nil, status.Error(codes.Unknown, err.Error())
			}
			certificateRotationReconciler, err := certsreconciler.NewCertificateReconciler(ctx)
			if err != nil {
				return nil, status.Error(codes.Unknown, err.Error())
			}
			svc.ReconcileRegistry.Register(clusterStatusReconciler)
			svc.ReconcileRegistry.Register(certificateRotationReconciler)
		}
	}
	registerFn := func(s *grpc.Server) error {
		v1alpha1.RegisterAgentAPIServer(s, NewServer(svc))
		return nil
	}
	lis, err := net.Listen("tcp", a.config.GRPCAddr)
	if err != nil {
		return nil, fmt.Errorf("create gRPC listener: %w", err)
	}
	config := grpcutil2.ServerConfig{
		Listener:     lis,
		RegisterFunc: registerFn,
	}
	server, err := grpcutil2.NewServer("GRPCServer", config)
	if err != nil {
		return nil, fmt.Errorf("create gRPC server: %w", err)
	}
	if err != nil {
		logger.Error(err, "unable to load tls credentials")
		return nil, fmt.Errorf("unable to load tls credentials: %w", err)
	}
	serverOptions := []grpc.ServerOption{
		grpc.ChainUnaryInterceptor(unaryInterceptor, prometheus.NewServerMetrics().UnaryServerInterceptor()),
	}
	ch, err := server.StartWithMetricsServer(ctx, a.config.ServerCertFilePath, a.config.ServerKeyFilePath, a.config.TokenSharedKey, serverOptions...)
	if err != nil {
		return nil, fmt.Errorf("start server: %w", err)
	}
	return ch, nil
}

func (a *App) startGateway(ctx context.Context) (<-chan struct{}, error) {
	config := grpcutil2.GatewayConfig{
		ServerAddr: a.config.GRPCAddr,
		HTTPAddr:   a.config.ServerAddr,
		Endpoints:  []grpcutil2.Endpoint{v1alpha1.RegisterAgentAPIHandlerFromEndpoint},
	}
	GRPCDialOptions := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	g := grpcutil2.NewGateway("GRPCGateway", config, GRPCDialOptions)

	options := []runtime.ServeMuxOption{
		runtime.WithMarshalerOption(runtime.MIMEWildcard, &runtime.JSONPb{
			MarshalOptions: protojson.MarshalOptions{
				Multiline: true,
				Indent:    "  ",
			},
			UnmarshalOptions: protojson.UnmarshalOptions{
				DiscardUnknown: true,
			},
		}),
	}

	ch, err := g.Start(ctx, a.config.ServerCertFilePath, a.config.ServerKeyFilePath, options...)
	if err != nil {
		return nil, fmt.Errorf("start gateway: %w", err)
	}

	return ch, nil
}
