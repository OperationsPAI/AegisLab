package app_test

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"testing"
	"time"

	"aegis/app"
	gateway "aegis/app/gateway"
	iam "aegis/app/iam"
	orchestrator "aegis/app/orchestrator"
	resource "aegis/app/resource"
	runtimeapp "aegis/app/runtime"
	system "aegis/app/system"
	buildkit "aegis/infra/buildkit"
	etcd "aegis/infra/etcd"
	harbor "aegis/infra/harbor"
	helm "aegis/infra/helm"
	k8s "aegis/infra/k8s"
	loki "aegis/infra/loki"
	redisinfra "aegis/infra/redis"
	controllerapi "aegis/interface/controller"
	httpapi "aegis/interface/http"
	receiverapi "aegis/interface/receiver"
	workerapi "aegis/interface/worker"
	resourcev1 "aegis/proto/resource/v1"
	runtimev1 "aegis/proto/runtime/v1"
	systemv1 "aegis/proto/system/v1"

	"github.com/DATA-DOG/go-sqlmock"
	goredis "github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/client-go/rest"
)

func newSmokeDB(t *testing.T) (*gorm.DB, func()) {
	t.Helper()

	sqlDB, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("create sqlmock: %v", err)
	}

	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	if err != nil {
		_ = sqlDB.Close()
		t.Fatalf("open gorm db: %v", err)
	}

	return db, func() {
		_ = sqlDB.Close()
	}
}

func newDedicatedServiceReplacements(t *testing.T) (fx.Option, func()) {
	t.Helper()

	db, cleanupDB := newSmokeDB(t)
	redisClient := goredis.NewClient(&goredis.Options{Addr: "127.0.0.1:0"})
	redisGateway := redisinfra.NewGateway(redisClient)
	etcdClient := &clientv3.Client{}
	etcdGateway := etcd.NewGateway(etcdClient)
	traceProvider := trace.NewTracerProvider()
	controller := &k8s.Controller{}
	k8sGateway := k8s.NewGateway(controller)

	return fx.Replace(
			db,
			redisGateway,
			redisClient,
			etcdGateway,
			etcdClient,
			&loki.Client{},
			traceProvider,
			&rest.Config{},
			controller,
			k8sGateway,
			harbor.NewGateway(),
			helm.NewGateway(),
			buildkit.NewGateway(),
			&app.ProducerInitializer{StartFunc: func(context.Context) error { return nil }},
			&workerapi.Lifecycle{StartFunc: func(context.Context) error { return nil }},
			&controllerapi.Lifecycle{RunFunc: func(context.Context, context.CancelFunc) error { return nil }},
			&receiverapi.Lifecycle{StartFunc: func(context.Context) error { return nil }},
		), func() {
			_ = redisClient.Close()
			_ = traceProvider.Shutdown(context.Background())
			cleanupDB()
		}
}

func reserveLoopbackAddr(t *testing.T) string {
	t.Helper()

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen on loopback: %v", err)
	}
	addr := listener.Addr().String()
	if err := listener.Close(); err != nil {
		t.Fatalf("close reserved listener: %v", err)
	}
	return addr
}

func setConfigValue(t *testing.T, key string, value any) {
	t.Helper()

	original := viper.Get(key)
	viper.Set(key, value)
	t.Cleanup(func() {
		viper.Set(key, original)
	})
}

func waitForHTTPStatus(t *testing.T, client *http.Client, method, url string, want int) {
	t.Helper()

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		req, err := http.NewRequest(method, url, nil)
		if err != nil {
			t.Fatalf("create request %s %s: %v", method, url, err)
		}

		resp, err := client.Do(req)
		if err == nil {
			_ = resp.Body.Close()
			if resp.StatusCode == want {
				return
			}
		}
		time.Sleep(50 * time.Millisecond)
	}

	req, _ := http.NewRequest(method, url, nil)
	resp, err := client.Do(req)
	if err != nil {
		t.Fatalf("request %s %s failed: %v", method, url, err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	t.Fatalf("expected %d from %s %s, got %d", want, method, url, resp.StatusCode)
}

func waitForRuntimePing(t *testing.T, addr string) {
	t.Helper()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("create runtime grpc client: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	client := runtimev1.NewRuntimeServiceClient(conn)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Ping(context.Background(), &runtimev1.PingRequest{})
		if err == nil && resp.GetService() != "" {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	resp, err := client.GetRuntimeStatus(context.Background(), &runtimev1.RuntimeStatusRequest{})
	if err != nil {
		t.Fatalf("runtime grpc request failed: %v", err)
	}
	if resp.GetService() == "" {
		t.Fatalf("runtime status missing service name: %+v", resp)
	}
}

func waitForResourcePing(t *testing.T, addr string) {
	t.Helper()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("create resource grpc client: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	client := resourcev1.NewResourceServiceClient(conn)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Ping(context.Background(), &resourcev1.PingRequest{})
		if err == nil && resp.GetService() != "" {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	resp, err := client.Ping(context.Background(), &resourcev1.PingRequest{})
	if err != nil {
		t.Fatalf("resource grpc request failed: %v", err)
	}
	if resp.GetService() == "" {
		t.Fatalf("resource ping missing service name: %+v", resp)
	}
}

func waitForSystemPing(t *testing.T, addr string) {
	t.Helper()

	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		t.Fatalf("create system grpc client: %v", err)
	}
	defer func() {
		_ = conn.Close()
	}()

	client := systemv1.NewSystemServiceClient(conn)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		resp, err := client.Ping(context.Background(), &systemv1.PingRequest{})
		if err == nil && resp.GetService() != "" {
			return
		}
		time.Sleep(50 * time.Millisecond)
	}

	resp, err := client.Ping(context.Background(), &systemv1.PingRequest{})
	if err != nil {
		t.Fatalf("system grpc request failed: %v", err)
	}
	if resp.GetService() == "" {
		t.Fatalf("system ping missing service name: %+v", resp)
	}
}

func TestDedicatedServiceOptionsValidate(t *testing.T) {
	for _, tc := range []struct {
		name   string
		option fx.Option
	}{
		{name: "gateway", option: gateway.Options("..", "0")},
		{name: "runtime", option: runtimeapp.Options("..")},
		{name: "resource", option: resource.Options("..")},
		{name: "system", option: system.Options("..")},
		{name: "iam", option: iam.Options("..")},
		{name: "orchestrator", option: orchestrator.Options("..")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if err := fx.ValidateApp(tc.option); err != nil {
				t.Fatalf("validate %s app: %v", tc.name, err)
			}
		})
	}
}

func TestAPIGatewayStandaloneHTTPIntegrationSmoke(t *testing.T) {
	replacements, cleanup := newDedicatedServiceReplacements(t)
	defer cleanup()

	setConfigValue(t, "clients.iam.target", reserveLoopbackAddr(t))
	setConfigValue(t, "clients.orchestrator.target", reserveLoopbackAddr(t))
	setConfigValue(t, "clients.resource.target", reserveLoopbackAddr(t))
	setConfigValue(t, "clients.system.target", reserveLoopbackAddr(t))

	addr := reserveLoopbackAddr(t)
	appInstance := fx.New(
		gateway.Options("..", "0"),
		replacements,
		fx.Replace(httpapi.ServerConfig{Addr: addr}),
	)

	startCtx, startCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer startCancel()
	if err := appInstance.Start(startCtx); err != nil {
		t.Fatalf("gateway app start failed: %v", err)
	}
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		if err := appInstance.Stop(stopCtx); err != nil {
			t.Fatalf("gateway app stop failed: %v", err)
		}
	}()

	client := &http.Client{Timeout: time.Second}
	baseURL := fmt.Sprintf("http://%s", addr)
	waitForHTTPStatus(t, client, http.MethodGet, baseURL+"/docs/doc.json", http.StatusOK)
	waitForHTTPStatus(t, client, http.MethodGet, baseURL+"/api/v2/system/configs/abc", http.StatusUnauthorized)
}

func TestRuntimeWorkerStandaloneGRPCIntegrationSmoke(t *testing.T) {
	replacements, cleanup := newDedicatedServiceReplacements(t)
	defer cleanup()

	setConfigValue(t, "clients.orchestrator.target", reserveLoopbackAddr(t))
	addr := reserveLoopbackAddr(t)
	setConfigValue(t, "runtime_worker.grpc.addr", addr)

	appInstance := fx.New(
		runtimeapp.Options(".."),
		replacements,
	)

	startCtx, startCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer startCancel()
	if err := appInstance.Start(startCtx); err != nil {
		t.Fatalf("runtime app start failed: %v", err)
	}
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		if err := appInstance.Stop(stopCtx); err != nil {
			t.Fatalf("runtime app stop failed: %v", err)
		}
	}()

	waitForRuntimePing(t, addr)
}

func TestResourceServiceStandaloneGRPCIntegrationSmoke(t *testing.T) {
	replacements, cleanup := newDedicatedServiceReplacements(t)
	defer cleanup()

	setConfigValue(t, "clients.orchestrator.target", reserveLoopbackAddr(t))
	addr := reserveLoopbackAddr(t)
	setConfigValue(t, "resource.grpc.addr", addr)

	appInstance := fx.New(
		resource.Options(".."),
		replacements,
	)

	startCtx, startCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer startCancel()
	if err := appInstance.Start(startCtx); err != nil {
		t.Fatalf("resource app start failed: %v", err)
	}
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		if err := appInstance.Stop(stopCtx); err != nil {
			t.Fatalf("resource app stop failed: %v", err)
		}
	}()

	waitForResourcePing(t, addr)
}

func TestSystemServiceStandaloneGRPCIntegrationSmoke(t *testing.T) {
	replacements, cleanup := newDedicatedServiceReplacements(t)
	defer cleanup()

	setConfigValue(t, "clients.runtime.target", reserveLoopbackAddr(t))
	addr := reserveLoopbackAddr(t)
	setConfigValue(t, "system.grpc.addr", addr)

	appInstance := fx.New(
		system.Options(".."),
		replacements,
	)

	startCtx, startCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer startCancel()
	if err := appInstance.Start(startCtx); err != nil {
		t.Fatalf("system app start failed: %v", err)
	}
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		if err := appInstance.Stop(stopCtx); err != nil {
			t.Fatalf("system app stop failed: %v", err)
		}
	}()

	waitForSystemPing(t, addr)
}
