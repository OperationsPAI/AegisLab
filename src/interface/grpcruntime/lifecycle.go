package grpcruntimeinterface

import (
	"context"
	"fmt"
	"net"

	"aegis/config"
	"aegis/httpx"
	runtimev1 "aegis/proto/runtime/v1"

	"github.com/sirupsen/logrus"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	grpc_health_v1 "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

const defaultRuntimeGRPCAddr = ":9094"

type Lifecycle struct {
	server    *grpc.Server
	addr      string
	listener  net.Listener
	StartFunc func(context.Context) error
	StopFunc  func()
}

func newLifecycle(runtimeServer *runtimeServer) (*Lifecycle, error) {
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(httpx.UnaryServerRequestIDInterceptor()))
	runtimev1.RegisterRuntimeServiceServer(grpcServer, runtimeServer)

	healthServer := health.NewServer()
	healthServer.SetServingStatus(runtimev1.RuntimeService_ServiceDesc.ServiceName, grpc_health_v1.HealthCheckResponse_SERVING)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)

	if config.GetBool("runtime_worker.grpc.reflection") {
		reflection.Register(grpcServer)
	}

	addr := config.GetString("runtime_worker.grpc.addr")
	if addr == "" {
		addr = defaultRuntimeGRPCAddr
	}

	return &Lifecycle{
		server: grpcServer,
		addr:   addr,
	}, nil
}

func (r *Lifecycle) start(ctx context.Context) error {
	if r.StartFunc != nil {
		return r.StartFunc(ctx)
	}

	listener, err := net.Listen("tcp", r.addr)
	if err != nil {
		return fmt.Errorf("listen runtime grpc on %s: %w", r.addr, err)
	}
	r.listener = listener

	go func() {
		logrus.Infof("Starting runtime gRPC server on %s", r.addr)
		if err := r.server.Serve(listener); err != nil {
			logrus.Errorf("runtime gRPC server error: %v", err)
		}
	}()
	return nil
}

func (r *Lifecycle) stop() {
	if r.StopFunc != nil {
		r.StopFunc()
		return
	}
	if r.server != nil {
		r.server.GracefulStop()
	}
}

func registerLifecycle(lc fx.Lifecycle, runner *Lifecycle) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			return runner.start(ctx)
		},
		OnStop: func(ctx context.Context) error {
			runner.stop()
			return nil
		},
	})
}
