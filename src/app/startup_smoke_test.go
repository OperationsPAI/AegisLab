package app

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"testing"
	"time"

	buildkitinfra "aegis/infra/buildkit"
	etcdinfra "aegis/infra/etcd"
	harborinfra "aegis/infra/harbor"
	helminfra "aegis/infra/helm"
	k8sinfra "aegis/infra/k8s"
	lokiinfra "aegis/infra/loki"
	redisinfra "aegis/infra/redis"
	controllerinterface "aegis/interface/controller"
	httpinterface "aegis/interface/http"
	receiverinterface "aegis/interface/receiver"
	workerinterface "aegis/interface/worker"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/redis/go-redis/v9"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.uber.org/fx"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"k8s.io/client-go/rest"
)

type smokeLifecycleSpies struct {
	producerStarts   int32
	workerStarts     int32
	workerStops      int32
	controllerStarts int32
	controllerStops  int32
	receiverStarts   int32
	receiverStops    int32
}

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

func newSmokeReplacements(t *testing.T, spies *smokeLifecycleSpies) (fx.Option, func()) {
	t.Helper()

	db, cleanupDB := newSmokeDB(t)
	redisClient := redis.NewClient(&redis.Options{Addr: "127.0.0.1:0"})
	redisGateway := redisinfra.NewGateway(redisClient)
	etcdClient := &clientv3.Client{}
	etcdGateway := etcdinfra.NewGateway(etcdClient)
	traceProvider := trace.NewTracerProvider()
	controller := &k8sinfra.Controller{}
	k8sGateway := k8sinfra.NewGateway(controller)

	producerInitializer := &ProducerInitializer{StartFunc: func(context.Context) error {
		if spies != nil {
			atomic.AddInt32(&spies.producerStarts, 1)
		}
		return nil
	}}
	workerLifecycle := &workerinterface.Lifecycle{
		StartFunc: func(context.Context) error {
			if spies != nil {
				atomic.AddInt32(&spies.workerStarts, 1)
			}
			return nil
		},
		StopFunc: func() {
			if spies != nil {
				atomic.AddInt32(&spies.workerStops, 1)
			}
		},
	}
	controllerLifecycle := &controllerinterface.Lifecycle{
		RunFunc: func(context.Context, context.CancelFunc) error {
			if spies != nil {
				atomic.AddInt32(&spies.controllerStarts, 1)
			}
			return nil
		},
		StopFunc: func() {
			if spies != nil {
				atomic.AddInt32(&spies.controllerStops, 1)
			}
		},
	}
	receiverLifecycle := &receiverinterface.Lifecycle{
		StartFunc: func(context.Context) error {
			if spies != nil {
				atomic.AddInt32(&spies.receiverStarts, 1)
			}
			return nil
		},
		StopFunc: func() {
			if spies != nil {
				atomic.AddInt32(&spies.receiverStops, 1)
			}
		},
	}

	return fx.Replace(
			db,
			redisGateway,
			redisClient,
			etcdGateway,
			etcdClient,
			&lokiinfra.Client{},
			traceProvider,
			&rest.Config{},
			controller,
			k8sGateway,
			harborinfra.NewGateway(),
			helminfra.NewGateway(),
			buildkitinfra.NewGateway(),
			producerInitializer,
			workerLifecycle,
			controllerLifecycle,
			receiverLifecycle,
		), func() {
			redisClient.Close()
			traceProvider.Shutdown(context.Background())
			cleanupDB()
		}
}

func startAndStopApp(t *testing.T, option fx.Option) {
	t.Helper()

	app := fx.New(option)
	startCtx, startCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer startCancel()
	if err := app.Start(startCtx); err != nil {
		t.Fatalf("app start failed: %v", err)
	}

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer stopCancel()
	if err := app.Stop(stopCtx); err != nil {
		t.Fatalf("app stop failed: %v", err)
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
	defer resp.Body.Close()
	t.Fatalf("expected %d from %s %s, got %d", want, method, url, resp.StatusCode)
}

func requireLifecycleCallCount(t *testing.T, name string, got *int32, want int32) {
	t.Helper()

	if actual := atomic.LoadInt32(got); actual != want {
		t.Fatalf("expected %s call count %d, got %d", name, want, actual)
	}
}

func TestProducerOptionsStartStopSmoke(t *testing.T) {
	replacements, cleanup := newSmokeReplacements(t, nil)
	defer cleanup()

	startAndStopApp(t, fx.Options(
		ProducerOptions("..", "0"),
		replacements,
	))
}

func TestConsumerOptionsStartStopSmoke(t *testing.T) {
	replacements, cleanup := newSmokeReplacements(t, nil)
	defer cleanup()

	startAndStopApp(t, fx.Options(
		ConsumerOptions(".."),
		replacements,
	))
}

func TestBothOptionsStartStopSmoke(t *testing.T) {
	replacements, cleanup := newSmokeReplacements(t, nil)
	defer cleanup()

	startAndStopApp(t, fx.Options(
		BothOptions("..", "0"),
		replacements,
	))
}

func TestProducerOptionsHTTPIntegrationSmoke(t *testing.T) {
	replacements, cleanup := newSmokeReplacements(t, nil)
	defer cleanup()

	addr := reserveLoopbackAddr(t)
	app := fx.New(
		ProducerOptions("..", "0"),
		replacements,
		fx.Replace(httpinterface.ServerConfig{Addr: addr}),
	)

	startCtx, startCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer startCancel()
	if err := app.Start(startCtx); err != nil {
		t.Fatalf("app start failed: %v", err)
	}
	defer func() {
		stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer stopCancel()
		if err := app.Stop(stopCtx); err != nil {
			t.Fatalf("app stop failed: %v", err)
		}
	}()

	client := &http.Client{Timeout: time.Second}
	baseURL := fmt.Sprintf("http://%s", addr)
	waitForHTTPStatus(t, client, http.MethodGet, baseURL+"/docs/doc.json", http.StatusOK)
	waitForHTTPStatus(t, client, http.MethodGet, baseURL+"/system/configs/abc", http.StatusUnauthorized)
}

func TestConsumerOptionsLifecycleIntegrationSmoke(t *testing.T) {
	spies := &smokeLifecycleSpies{}
	replacements, cleanup := newSmokeReplacements(t, spies)
	defer cleanup()

	startAndStopApp(t, fx.Options(
		ConsumerOptions(".."),
		replacements,
	))

	requireLifecycleCallCount(t, "worker start", &spies.workerStarts, 1)
	requireLifecycleCallCount(t, "worker stop", &spies.workerStops, 1)
	requireLifecycleCallCount(t, "controller start", &spies.controllerStarts, 1)
	requireLifecycleCallCount(t, "controller stop", &spies.controllerStops, 1)
	requireLifecycleCallCount(t, "receiver start", &spies.receiverStarts, 1)
	requireLifecycleCallCount(t, "receiver stop", &spies.receiverStops, 1)
}

func TestBothOptionsHTTPAndLifecycleIntegrationSmoke(t *testing.T) {
	spies := &smokeLifecycleSpies{}
	replacements, cleanup := newSmokeReplacements(t, spies)
	defer cleanup()

	addr := reserveLoopbackAddr(t)
	app := fx.New(
		BothOptions("..", "0"),
		replacements,
		fx.Replace(httpinterface.ServerConfig{Addr: addr}),
	)

	startCtx, startCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer startCancel()
	if err := app.Start(startCtx); err != nil {
		t.Fatalf("app start failed: %v", err)
	}

	client := &http.Client{Timeout: time.Second}
	baseURL := fmt.Sprintf("http://%s", addr)
	waitForHTTPStatus(t, client, http.MethodGet, baseURL+"/docs/doc.json", http.StatusOK)
	waitForHTTPStatus(t, client, http.MethodGet, baseURL+"/system/configs/abc", http.StatusUnauthorized)
	requireLifecycleCallCount(t, "producer start", &spies.producerStarts, 1)
	requireLifecycleCallCount(t, "worker start", &spies.workerStarts, 1)
	requireLifecycleCallCount(t, "controller start", &spies.controllerStarts, 1)
	requireLifecycleCallCount(t, "receiver start", &spies.receiverStarts, 1)

	stopCtx, stopCancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer stopCancel()
	if err := app.Stop(stopCtx); err != nil {
		t.Fatalf("app stop failed: %v", err)
	}

	requireLifecycleCallCount(t, "worker stop", &spies.workerStops, 1)
	requireLifecycleCallCount(t, "controller stop", &spies.controllerStops, 1)
	requireLifecycleCallCount(t, "receiver stop", &spies.receiverStops, 1)
}
