package cluster

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeEnv struct {
	cfg Config
	k8s *fakeK8s
	net *fakeNet
	ch  *fakeCH
	sql *fakeMySQL
	rds *fakeRedis
}

func (f *fakeEnv) Config() Config              { return f.cfg }
func (f *fakeEnv) K8s() K8sProbe               { return f.k8s }
func (f *fakeEnv) Net() NetProbe               { return f.net }
func (f *fakeEnv) ClickHouse() ClickHouseProbe { return f.ch }
func (f *fakeEnv) MySQL() MySQLProbe           { return f.sql }
func (f *fakeEnv) Redis() RedisProbe           { return f.rds }

type fakeK8s struct {
	namespaces      map[string]bool
	serviceAccounts map[string]bool
	pvcs            map[string]struct {
		exists bool
		bound  bool
	}
	crdGroups map[string]bool
	saCreated []string
	err       error
}

func (f *fakeK8s) NamespaceExists(_ context.Context, n string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.namespaces[n], nil
}
func (f *fakeK8s) ServiceAccountExists(_ context.Context, ns, n string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.serviceAccounts[ns+"/"+n], nil
}
func (f *fakeK8s) PVCBound(_ context.Context, ns, n string) (bool, bool, error) {
	if f.err != nil {
		return false, false, f.err
	}
	s := f.pvcs[ns+"/"+n]
	return s.exists, s.bound, nil
}
func (f *fakeK8s) HasCRDGroup(_ context.Context, g string) (bool, error) {
	if f.err != nil {
		return false, f.err
	}
	return f.crdGroups[g], nil
}
func (f *fakeK8s) CreateServiceAccount(_ context.Context, ns, n string) error {
	if f.serviceAccounts == nil {
		f.serviceAccounts = map[string]bool{}
	}
	f.serviceAccounts[ns+"/"+n] = true
	f.saCreated = append(f.saCreated, ns+"/"+n)
	return nil
}

type fakeNet struct {
	ok     map[string]bool
	called []string
}

func (f *fakeNet) DialTimeout(_ context.Context, addr string) error {
	f.called = append(f.called, addr)
	if f.ok[addr] {
		return nil
	}
	return errors.New("connection refused")
}

type fakeCH struct {
	tables map[string][]string
	err    error
}

func (f *fakeCH) TablesIn(_ context.Context, db string) ([]string, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.tables[db], nil
}

type fakeMySQL struct {
	state  map[string]int
	exists map[string]bool
}

func (f *fakeMySQL) TaskState(_ context.Context, id string) (int, bool, error) {
	if !f.exists[id] {
		return 0, false, nil
	}
	return f.state[id], true, nil
}

type fakeRedis struct {
	sets    map[string][]string
	removed map[string][]string
}

func (f *fakeRedis) SMembers(_ context.Context, key string) ([]string, error) {
	return append([]string(nil), f.sets[key]...), nil
}
func (f *fakeRedis) SRem(_ context.Context, key string, members ...string) (int64, error) {
	if f.removed == nil {
		f.removed = map[string][]string{}
	}
	f.removed[key] = append(f.removed[key], members...)
	toRemove := map[string]bool{}
	for _, m := range members {
		toRemove[m] = true
	}
	keep := f.sets[key][:0]
	for _, m := range f.sets[key] {
		if !toRemove[m] {
			keep = append(keep, m)
		}
	}
	f.sets[key] = keep
	return int64(len(members)), nil
}

func TestCheckRcabenchSA_Missing(t *testing.T) {
	env := &fakeEnv{
		cfg: Config{K8sNamespace: "exp", ServiceAccount: "rcabench-sa"},
		k8s: &fakeK8s{serviceAccounts: map[string]bool{}},
	}
	res := checkRcabenchSA(context.Background(), env)
	if res.Status != StatusFail {
		t.Fatalf("expected FAIL, got %s (detail=%q)", res.Status, res.Detail)
	}
	if res.Fix == "" {
		t.Fatalf("expected fix suggestion on FAIL")
	}
	if err := fixRcabenchSA(context.Background(), env); err != nil {
		t.Fatalf("fix: %v", err)
	}
	res = checkRcabenchSA(context.Background(), env)
	if res.Status != StatusOK {
		t.Fatalf("after fix expected OK, got %s", res.Status)
	}
}

func TestCheckOtelTables_MissingSome(t *testing.T) {
	env := &fakeEnv{
		cfg: Config{ClickHouseHost: "localhost", ClickHouseDB: "otel"},
		ch: &fakeCH{tables: map[string][]string{
			"otel": {"otel_traces", "otel_metrics_gauge"},
		}},
	}
	res := checkOtelTables(context.Background(), env)
	if res.Status != StatusFail {
		t.Fatalf("expected FAIL, got %s", res.Status)
	}
	for _, need := range []string{"otel_metrics_sum", "otel_metrics_histogram", "otel_logs"} {
		if !strings.Contains(res.Detail, need) {
			t.Errorf("detail %q should flag missing table %q", res.Detail, need)
		}
	}
}

func TestCheckOtelTables_AllPresent(t *testing.T) {
	env := &fakeEnv{
		cfg: Config{ClickHouseHost: "localhost", ClickHouseDB: "otel"},
		ch:  &fakeCH{tables: map[string][]string{"otel": ExpectedOtelTables}},
	}
	res := checkOtelTables(context.Background(), env)
	if res.Status != StatusOK {
		t.Fatalf("expected OK, got %s (detail=%q)", res.Status, res.Detail)
	}
}

func TestCheckTokenBucketLeaks_DetectsTerminalTasks(t *testing.T) {
	env := &fakeEnv{
		rds: &fakeRedis{sets: map[string][]string{
			RestartTokenBucketKey: {"live-1", "done-1", "err-1", "ghost-1"},
		}},
		sql: &fakeMySQL{
			exists: map[string]bool{"live-1": true, "done-1": true, "err-1": true},
			state:  map[string]int{"live-1": 2, "done-1": 3, "err-1": -1},
		},
	}
	res := checkTokenBucketLeaks(context.Background(), env)
	if res.Status != StatusFail {
		t.Fatalf("expected FAIL, got %s (detail=%q)", res.Status, res.Detail)
	}
	for _, want := range []string{"done-1", "err-1", "ghost-1"} {
		if !strings.Contains(res.Detail, want) {
			t.Errorf("expected %q in leaked list; got %q", want, res.Detail)
		}
	}
	if strings.Contains(res.Detail, "live-1") {
		t.Errorf("live task should not be flagged as leaked: %q", res.Detail)
	}
	if err := fixTokenBucketLeaks(context.Background(), env); err != nil {
		t.Fatalf("fix: %v", err)
	}
	res = checkTokenBucketLeaks(context.Background(), env)
	if res.Status != StatusOK {
		t.Fatalf("after fix expected OK, got %s (detail=%q)", res.Status, res.Detail)
	}
	members := env.rds.sets[RestartTokenBucketKey]
	if len(members) != 1 || members[0] != "live-1" {
		t.Errorf("expected bucket to contain only live-1, got %v", members)
	}
}

func TestDefaultChecks_RegistryCatalog(t *testing.T) {
	reg := NewRegistry(DefaultChecks())
	expected := []string{
		"k8s.exp-namespace", "k8s.rcabench-sa", "k8s.dataset-pvc", "k8s.chaosmesh-crds",
		"db.mysql", "db.clickhouse", "db.redis", "db.etcd",
		"clickhouse.otel-tables", "redis.token-bucket-leaks",
	}
	for _, id := range expected {
		if _, ok := reg.Get(id); !ok {
			t.Errorf("missing registered check: %s", id)
		}
	}
}
