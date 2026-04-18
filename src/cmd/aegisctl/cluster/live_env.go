package cluster

import (
	"context"
	"database/sql"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"time"

	chdriver "github.com/ClickHouse/clickhouse-go/v2"
	_ "github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"github.com/spf13/viper"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// LiveEnv is the production CheckEnv implementation. It reads config from a
// viper-backed TOML file and lazily builds clients so that individual checks
// can run without every dependency being present (e.g. running only
// --check db.mysql should not require a kubeconfig).
type LiveEnv struct {
	cfg         Config
	k8s         K8sProbe
	net         NetProbe
	clickhouse  ClickHouseProbe
	mysql       MySQLProbe
	redis       RedisProbe
	mysqlCreds  mysqlCreds
	dialTimeout time.Duration
}

type mysqlCreds struct{ user, pass, db string }

var loadedMySQLCreds mysqlCreds

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// LoadConfig reads config.dev.toml (or the env-specific equivalent) and
// returns a populated Config. A missing file is tolerated — individual
// checks flag missing values themselves.
func LoadConfig(explicitPath string) (Config, error) {
	v := viper.New()
	v.SetConfigType("toml")
	if explicitPath != "" {
		v.SetConfigFile(explicitPath)
	} else {
		env := os.Getenv("ENV_MODE")
		if env == "" {
			env = "dev"
		}
		v.SetConfigName("config." + env)
		cwd, _ := os.Getwd()
		v.AddConfigPath(cwd)
		v.AddConfigPath(filepath.Join(cwd, "src"))
		v.AddConfigPath(".")
	}
	_ = v.ReadInConfig()

	cfg := Config{
		K8sNamespace:   v.GetString("k8s.namespace"),
		MySQLHost:      v.GetString("database.mysql.host"),
		MySQLPort:      v.GetString("database.mysql.port"),
		ClickHouseHost: v.GetString("database.clickhouse.host"),
		ClickHousePort: v.GetString("database.clickhouse.port"),
		ClickHouseDB:   v.GetString("database.clickhouse.database"),
		ClickHouseUser: v.GetString("database.clickhouse.user"),
		ClickHousePass: v.GetString("database.clickhouse.password"),
		RedisAddr:      v.GetString("redis.host"),
		EtcdEndpoints:  v.GetStringSlice("etcd.endpoints"),
		ServiceAccount: v.GetString("k8s.job.service_account.name"),
		DatasetPVC:     v.GetString("k8s.job.volume_mount.dataset.claim_name"),
	}
	loadedMySQLCreds = mysqlCreds{
		user: firstNonEmpty(v.GetString("database.mysql.user"), "root"),
		pass: v.GetString("database.mysql.password"),
		db:   firstNonEmpty(v.GetString("database.mysql.db"), "rcabench"),
	}
	return cfg, nil
}

func NewLiveEnv(cfg Config) *LiveEnv {
	return &LiveEnv{cfg: cfg, dialTimeout: 3 * time.Second, mysqlCreds: loadedMySQLCreds}
}

func (e *LiveEnv) Config() Config { return e.cfg }

func (e *LiveEnv) Net() NetProbe {
	if e.net == nil {
		e.net = &tcpProbe{timeout: e.dialTimeout}
	}
	return e.net
}

func (e *LiveEnv) K8s() K8sProbe {
	if e.k8s == nil {
		e.k8s = newLiveK8s()
	}
	return e.k8s
}

func (e *LiveEnv) ClickHouse() ClickHouseProbe {
	if e.clickhouse == nil {
		e.clickhouse = &liveClickHouse{cfg: e.cfg}
	}
	return e.clickhouse
}

func (e *LiveEnv) MySQL() MySQLProbe {
	if e.mysql == nil {
		e.mysql = &liveMySQL{cfg: e.cfg, creds: e.mysqlCreds}
	}
	return e.mysql
}

func (e *LiveEnv) Redis() RedisProbe {
	if e.redis == nil {
		e.redis = &liveRedis{cfg: e.cfg}
	}
	return e.redis
}

type tcpProbe struct{ timeout time.Duration }

func (t *tcpProbe) DialTimeout(ctx context.Context, address string) error {
	d := net.Dialer{Timeout: t.timeout}
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}
	_ = conn.Close()
	return nil
}

type liveK8s struct {
	cs      *kubernetes.Clientset
	dyn     dynamic.Interface
	loadErr error
}

func newLiveK8s() *liveK8s {
	k := &liveK8s{}
	cfg, err := inClusterOrKubeconfig()
	if err != nil {
		k.loadErr = err
		return k
	}
	cs, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		k.loadErr = fmt.Errorf("build clientset: %w", err)
		return k
	}
	dyn, err := dynamic.NewForConfig(cfg)
	if err != nil {
		k.loadErr = fmt.Errorf("build dynamic client: %w", err)
		return k
	}
	k.cs = cs
	k.dyn = dyn
	return k
}

func inClusterOrKubeconfig() (*rest.Config, error) {
	if c, err := rest.InClusterConfig(); err == nil {
		return c, nil
	}
	home, _ := os.UserHomeDir()
	path := filepath.Join(home, ".kube", "config")
	return clientcmd.BuildConfigFromFlags("", path)
}

func (k *liveK8s) NamespaceExists(ctx context.Context, name string) (bool, error) {
	if k.loadErr != nil {
		return false, k.loadErr
	}
	_, err := k.cs.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (k *liveK8s) ServiceAccountExists(ctx context.Context, namespace, name string) (bool, error) {
	if k.loadErr != nil {
		return false, k.loadErr
	}
	_, err := k.cs.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

func (k *liveK8s) PVCBound(ctx context.Context, namespace, name string) (bool, bool, error) {
	if k.loadErr != nil {
		return false, false, k.loadErr
	}
	pvc, err := k.cs.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if apierrors.IsNotFound(err) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	return true, pvc.Status.Phase == corev1.ClaimBound, nil
}

func (k *liveK8s) HasCRDGroup(ctx context.Context, group string) (bool, error) {
	if k.loadErr != nil {
		return false, k.loadErr
	}
	gvr := schema.GroupVersionResource{
		Group:    "apiextensions.k8s.io",
		Version:  "v1",
		Resource: "customresourcedefinitions",
	}
	list, err := k.dyn.Resource(gvr).List(ctx, metav1.ListOptions{})
	if err != nil {
		return false, err
	}
	for _, item := range list.Items {
		if unstructuredString(item.Object, "spec", "group") == group {
			return true, nil
		}
	}
	return false, nil
}

func (k *liveK8s) CreateServiceAccount(ctx context.Context, namespace, name string) error {
	if k.loadErr != nil {
		return k.loadErr
	}
	sa := &corev1.ServiceAccount{ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace}}
	_, err := k.cs.CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
	if apierrors.IsAlreadyExists(err) {
		return nil
	}
	return err
}

func unstructuredString(obj map[string]any, path ...string) string {
	var cur any = obj
	for _, p := range path {
		m, ok := cur.(map[string]any)
		if !ok {
			return ""
		}
		cur = m[p]
	}
	s, _ := cur.(string)
	return s
}

type liveClickHouse struct{ cfg Config }

func (c *liveClickHouse) TablesIn(ctx context.Context, database string) ([]string, error) {
	host := c.cfg.ClickHouseHost
	port := c.cfg.ClickHousePort
	if port == "" {
		port = "9000"
	}
	addr := net.JoinHostPort(host, port)
	opts := &chdriver.Options{
		Addr: []string{addr},
		Auth: chdriver.Auth{
			Database: database,
			Username: c.cfg.ClickHouseUser,
			Password: c.cfg.ClickHousePass,
		},
		DialTimeout: 3 * time.Second,
	}
	conn, err := chdriver.Open(opts)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	rows, err := conn.Query(ctx, "SELECT name FROM system.tables WHERE database = ?", database)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

type liveMySQL struct {
	cfg   Config
	creds mysqlCreds
}

func (m *liveMySQL) TaskState(ctx context.Context, taskID string) (int, bool, error) {
	host := m.cfg.MySQLHost
	port := m.cfg.MySQLPort
	if port == "" {
		port = "3306"
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&timeout=3s&readTimeout=3s",
		m.creds.user, m.creds.pass, host, port, m.creds.db)
	conn, err := sql.Open("mysql", dsn)
	if err != nil {
		return 0, false, err
	}
	defer conn.Close()

	row := conn.QueryRowContext(ctx, "SELECT state FROM tasks WHERE id = ? LIMIT 1", taskID)
	var state int
	if err := row.Scan(&state); err != nil {
		if err == sql.ErrNoRows {
			return 0, false, nil
		}
		return 0, false, err
	}
	return state, true, nil
}

type liveRedis struct {
	cfg    Config
	client *redis.Client
}

func (r *liveRedis) ensure() *redis.Client {
	if r.client != nil {
		return r.client
	}
	addr := r.cfg.RedisAddr
	if addr != "" && !strings.Contains(addr, ":") {
		addr += ":6379"
	}
	r.client = redis.NewClient(&redis.Options{
		Addr:        addr,
		DialTimeout: 3 * time.Second,
		ReadTimeout: 3 * time.Second,
	})
	return r.client
}

func (r *liveRedis) SMembers(ctx context.Context, key string) ([]string, error) {
	return r.ensure().SMembers(ctx, key).Result()
}

func (r *liveRedis) SRem(ctx context.Context, key string, members ...string) (int64, error) {
	args := make([]any, len(members))
	for i, m := range members {
		args[i] = m
	}
	return r.ensure().SRem(ctx, key, args...).Result()
}
