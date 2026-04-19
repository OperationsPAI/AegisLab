package cluster

import (
	"context"
	"fmt"
	"net"
	"sort"
	"strings"
)

// ExpectedOtelTables lists the ClickHouse tables required by the otel pipeline.
var ExpectedOtelTables = []string{
	"otel_traces",
	"otel_metrics_gauge",
	"otel_metrics_sum",
	"otel_metrics_histogram",
	"otel_logs",
}

// ChaosMeshCRDGroup is the CRD API group exposed by Chaos Mesh.
const ChaosMeshCRDGroup = "chaos-mesh.org"

// RestartTokenBucketKey mirrors consts.RestartPedestalTokenBucket.
const RestartTokenBucketKey = "token_bucket:restart_service"

// terminalTaskStates matches the convention in src/consts/consts.go:
//   TaskCancelled = -2, TaskError = -1, TaskCompleted = 3.
// Any task in one of these states is no longer running and should not be
// holding a slot in the restart_service token bucket.
var terminalTaskStates = map[int]struct{}{
	-2: {}, -1: {}, 3: {},
}

// DefaultChecks returns the MVP preflight catalog.
func DefaultChecks() []Check {
	return []Check{
		{ID: "k8s.exp-namespace", Description: "namespace `exp` exists", Run: checkExpNamespace},
		{ID: "k8s.rcabench-sa", Description: "ServiceAccount `rcabench-sa` exists in `exp`", Run: checkRcabenchSA, Fix: fixRcabenchSA},
		{ID: "k8s.dataset-pvc", Description: "PVC `rcabench-juicefs-dataset` in `exp` is Bound", Run: checkDatasetPVC},
		{ID: "k8s.chaosmesh-crds", Description: "Chaos Mesh CRDs are installed", Run: checkChaosMeshCRDs},
		{ID: "db.mysql", Description: "MySQL is TCP-reachable", Run: checkMySQL},
		{ID: "db.clickhouse", Description: "ClickHouse is TCP-reachable", Run: checkClickHouseTCP},
		{ID: "db.redis", Description: "Redis is TCP-reachable", Run: checkRedis},
		{ID: "db.etcd", Description: "etcd is TCP-reachable", Run: checkEtcd},
		{ID: "clickhouse.otel-tables", Description: "ClickHouse has all otel pipeline tables", Run: checkOtelTables},
		{ID: "redis.token-bucket-leaks", Description: "restart_service token bucket has no terminal tasks", Run: checkTokenBucketLeaks, Fix: fixTokenBucketLeaks},
		// TODO(issue-17 follow-up): skipped in the MVP (too slow for a
		// synchronous preflight):
		//   - container_versions registries pullable
		//   - helm_configs.repo_url reachability
	}
}

func checkExpNamespace(ctx context.Context, env CheckEnv) Result {
	ns := env.Config().K8sNamespace
	if ns == "" {
		ns = "exp"
	}
	ok, err := env.K8s().NamespaceExists(ctx, ns)
	switch {
	case err != nil:
		return Result{Status: StatusFail, Detail: fmt.Sprintf("could not query k8s: %v", err),
			Fix: "ensure KUBECONFIG is set and the current context targets the AegisLab cluster"}
	case !ok:
		return Result{Status: StatusFail, Detail: fmt.Sprintf("namespace %q not found", ns),
			Fix: fmt.Sprintf("kubectl create namespace %s", ns)}
	}
	return Result{Status: StatusOK, Detail: fmt.Sprintf("namespace %q present", ns)}
}

func checkRcabenchSA(ctx context.Context, env CheckEnv) Result {
	cfg := env.Config()
	ns := cfg.K8sNamespace
	if ns == "" {
		ns = "exp"
	}
	sa := cfg.ServiceAccount
	if sa == "" {
		sa = "rcabench-sa"
	}
	ok, err := env.K8s().ServiceAccountExists(ctx, ns, sa)
	switch {
	case err != nil:
		return Result{Status: StatusFail, Detail: fmt.Sprintf("could not query k8s: %v", err),
			Fix: "ensure KUBECONFIG is valid"}
	case !ok:
		return Result{Status: StatusFail, Detail: fmt.Sprintf("ServiceAccount %s/%s missing", ns, sa),
			Fix: fmt.Sprintf("kubectl -n %s create serviceaccount %s (or rerun with --fix)", ns, sa)}
	}
	return Result{Status: StatusOK, Detail: fmt.Sprintf("%s/%s present", ns, sa)}
}

func fixRcabenchSA(ctx context.Context, env CheckEnv) error {
	cfg := env.Config()
	ns := cfg.K8sNamespace
	if ns == "" {
		ns = "exp"
	}
	sa := cfg.ServiceAccount
	if sa == "" {
		sa = "rcabench-sa"
	}
	return env.K8s().CreateServiceAccount(ctx, ns, sa)
}

func checkDatasetPVC(ctx context.Context, env CheckEnv) Result {
	cfg := env.Config()
	ns := cfg.K8sNamespace
	if ns == "" {
		ns = "exp"
	}
	pvc := cfg.DatasetPVC
	if pvc == "" {
		pvc = "rcabench-juicefs-dataset"
	}
	exists, bound, err := env.K8s().PVCBound(ctx, ns, pvc)
	switch {
	case err != nil:
		return Result{Status: StatusFail, Detail: fmt.Sprintf("could not query k8s: %v", err),
			Fix: "ensure KUBECONFIG is valid"}
	case !exists:
		return Result{Status: StatusFail, Detail: fmt.Sprintf("PVC %s/%s not found", ns, pvc),
			Fix: "apply the JuiceFS PVC manifest (manifests/juicefs-dataset-pvc.yaml)"}
	case !bound:
		return Result{Status: StatusFail, Detail: fmt.Sprintf("PVC %s/%s exists but is not Bound", ns, pvc),
			Fix: "verify the JuiceFS storage class and the corresponding PV have provisioned"}
	}
	return Result{Status: StatusOK, Detail: fmt.Sprintf("%s/%s Bound", ns, pvc)}
}

func checkChaosMeshCRDs(ctx context.Context, env CheckEnv) Result {
	ok, err := env.K8s().HasCRDGroup(ctx, ChaosMeshCRDGroup)
	switch {
	case err != nil:
		return Result{Status: StatusFail, Detail: fmt.Sprintf("could not list CRDs: %v", err),
			Fix: "ensure KUBECONFIG is valid and you have permission to list customresourcedefinitions"}
	case !ok:
		return Result{Status: StatusFail, Detail: "no CRDs found in group " + ChaosMeshCRDGroup,
			Fix: "helm install chaos-mesh chaos-mesh/chaos-mesh -n chaos-mesh --create-namespace"}
	}
	return Result{Status: StatusOK, Detail: "chaos-mesh.org CRDs present"}
}

func joinHostPort(host, port string) string {
	host = strings.TrimSpace(host)
	port = strings.TrimSpace(port)
	if host == "" {
		return ""
	}
	if strings.Contains(host, ":") {
		return host
	}
	if port == "" {
		return host
	}
	return net.JoinHostPort(host, port)
}

func checkMySQL(ctx context.Context, env CheckEnv) Result {
	cfg := env.Config()
	addr := joinHostPort(cfg.MySQLHost, cfg.MySQLPort)
	if addr == "" {
		return Result{Status: StatusFail, Detail: "database.mysql.host not configured",
			Fix: "set [database.mysql] host/port in config.dev.toml"}
	}
	if err := env.Net().DialTimeout(ctx, addr); err != nil {
		return Result{Status: StatusFail, Detail: fmt.Sprintf("cannot dial %s: %v", addr, err),
			Fix: "start MySQL (docker compose up mysql) and verify [database.mysql].host/port"}
	}
	return Result{Status: StatusOK, Detail: "dialed " + addr}
}

func checkClickHouseTCP(ctx context.Context, env CheckEnv) Result {
	cfg := env.Config()
	addr := joinHostPort(cfg.ClickHouseHost, cfg.ClickHousePort)
	if addr == "" {
		return Result{Status: StatusFail, Detail: "[database.clickhouse] host/port missing in config",
			Fix: "add a [database.clickhouse] section with host/port/database/user/password to config.dev.toml"}
	}
	if err := env.Net().DialTimeout(ctx, addr); err != nil {
		return Result{Status: StatusFail, Detail: fmt.Sprintf("cannot dial %s: %v", addr, err),
			Fix: "ensure ClickHouse is running and the [database.clickhouse] host/port match"}
	}
	return Result{Status: StatusOK, Detail: "dialed " + addr}
}

func checkRedis(ctx context.Context, env CheckEnv) Result {
	addr := strings.TrimSpace(env.Config().RedisAddr)
	if addr == "" {
		return Result{Status: StatusFail, Detail: "redis.host not configured",
			Fix: "set redis.host=host:port in config.dev.toml"}
	}
	if !strings.Contains(addr, ":") {
		addr = addr + ":6379"
	}
	if err := env.Net().DialTimeout(ctx, addr); err != nil {
		return Result{Status: StatusFail, Detail: fmt.Sprintf("cannot dial %s: %v", addr, err),
			Fix: "start Redis (docker compose up redis) and verify redis.host"}
	}
	return Result{Status: StatusOK, Detail: "dialed " + addr}
}

func checkEtcd(ctx context.Context, env CheckEnv) Result {
	eps := env.Config().EtcdEndpoints
	if len(eps) == 0 {
		return Result{Status: StatusFail, Detail: "etcd.endpoints not configured",
			Fix: "set etcd.endpoints in config.dev.toml"}
	}
	addr := strings.TrimSpace(eps[0])
	addr = strings.TrimPrefix(addr, "http://")
	addr = strings.TrimPrefix(addr, "https://")
	if err := env.Net().DialTimeout(ctx, addr); err != nil {
		return Result{Status: StatusFail, Detail: fmt.Sprintf("cannot dial %s: %v", addr, err),
			Fix: "start etcd and verify etcd.endpoints"}
	}
	return Result{Status: StatusOK, Detail: "dialed " + addr}
}

func checkOtelTables(ctx context.Context, env CheckEnv) Result {
	cfg := env.Config()
	db := cfg.ClickHouseDB
	if db == "" {
		db = "otel"
	}
	if cfg.ClickHouseHost == "" {
		return Result{Status: StatusFail, Detail: "[database.clickhouse] not configured",
			Fix: "add a [database.clickhouse] section to config.dev.toml"}
	}
	got, err := env.ClickHouse().TablesIn(ctx, db)
	if err != nil {
		return Result{Status: StatusFail, Detail: fmt.Sprintf("could not list tables in %s: %v", db, err),
			Fix: "verify ClickHouse is reachable and the configured user has SHOW TABLES permission"}
	}
	have := make(map[string]struct{}, len(got))
	for _, t := range got {
		have[t] = struct{}{}
	}
	var missing []string
	for _, need := range ExpectedOtelTables {
		if _, ok := have[need]; !ok {
			missing = append(missing, need)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return Result{Status: StatusFail,
			Detail: fmt.Sprintf("missing tables in %s: %s", db, strings.Join(missing, ", ")),
			Fix:    "re-run the otel-collector-contrib ClickHouse exporter schema bootstrap (see docs/e2e-deployment-pitfalls.md)"}
	}
	return Result{Status: StatusOK, Detail: fmt.Sprintf("all %d otel tables present in %s", len(ExpectedOtelTables), db)}
}

func checkTokenBucketLeaks(ctx context.Context, env CheckEnv) Result {
	members, err := env.Redis().SMembers(ctx, RestartTokenBucketKey)
	if err != nil {
		return Result{Status: StatusFail, Detail: fmt.Sprintf("could not SMEMBERS %s: %v", RestartTokenBucketKey, err),
			Fix: "verify redis.host is correct and the instance is reachable"}
	}
	if len(members) == 0 {
		return Result{Status: StatusOK, Detail: "token bucket empty"}
	}

	var leaked []string
	for _, id := range members {
		state, exists, err := env.MySQL().TaskState(ctx, id)
		if err != nil {
			return Result{Status: StatusFail,
				Detail: fmt.Sprintf("could not look up task %s: %v", id, err),
				Fix:    "verify [database.mysql] credentials in config.dev.toml"}
		}
		if !exists {
			leaked = append(leaked, id)
			continue
		}
		if _, terminal := terminalTaskStates[state]; terminal {
			leaked = append(leaked, id)
		}
	}
	if len(leaked) == 0 {
		return Result{Status: StatusOK, Detail: fmt.Sprintf("%d live holder(s)", len(members))}
	}
	sort.Strings(leaked)
	return Result{Status: StatusFail,
		Detail: fmt.Sprintf("%d leaked token(s): %s", len(leaked), strings.Join(leaked, ", ")),
		Fix:    "rerun with --fix to SREM leaked task_ids from " + RestartTokenBucketKey}
}

func fixTokenBucketLeaks(ctx context.Context, env CheckEnv) error {
	members, err := env.Redis().SMembers(ctx, RestartTokenBucketKey)
	if err != nil {
		return fmt.Errorf("smembers: %w", err)
	}
	var leaked []string
	for _, id := range members {
		state, exists, err := env.MySQL().TaskState(ctx, id)
		if err != nil {
			return fmt.Errorf("lookup task %s: %w", id, err)
		}
		if !exists {
			leaked = append(leaked, id)
			continue
		}
		if _, terminal := terminalTaskStates[state]; terminal {
			leaked = append(leaked, id)
		}
	}
	if len(leaked) == 0 {
		return nil
	}
	if _, err := env.Redis().SRem(ctx, RestartTokenBucketKey, leaked...); err != nil {
		return fmt.Errorf("srem: %w", err)
	}
	return nil
}
