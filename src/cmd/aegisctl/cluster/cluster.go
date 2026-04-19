// Package cluster implements the `aegisctl cluster preflight` dependency
// checker. Each check is expressed as a small CheckFunc operating on a
// CheckEnv abstraction so individual check functions can be exercised with
// mocks from unit tests without a live cluster.
package cluster

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"
)

// Status represents the outcome of a single preflight check.
type Status string

const (
	StatusOK   Status = "OK"
	StatusFail Status = "FAIL"
	StatusWarn Status = "WARN"
	StatusSkip Status = "SKIP"
)

// Result captures the outcome of one check.
type Result struct {
	ID     string
	Status Status
	Detail string
	Fix    string
}

// CheckFunc runs a single check.
type CheckFunc func(ctx context.Context, env CheckEnv) Result

// FixFunc performs idempotent remediation. Not every check has a fix.
type FixFunc func(ctx context.Context, env CheckEnv) error

// Check describes a single preflight check and its optional fixer.
type Check struct {
	ID          string
	Description string
	Run         CheckFunc
	Fix         FixFunc
}

// Registry is the ordered catalog of checks.
type Registry struct {
	checks []Check
	byID   map[string]int
}

func NewRegistry(checks []Check) *Registry {
	r := &Registry{checks: checks, byID: make(map[string]int, len(checks))}
	for i, c := range checks {
		r.byID[c.ID] = i
	}
	return r
}

func (r *Registry) Get(id string) (Check, bool) {
	idx, ok := r.byID[id]
	if !ok {
		return Check{}, false
	}
	return r.checks[idx], true
}

func (r *Registry) IDs() []string {
	out := make([]string, len(r.checks))
	for i, c := range r.checks {
		out[i] = c.ID
	}
	return out
}

func (r *Registry) All() []Check {
	out := make([]Check, len(r.checks))
	copy(out, r.checks)
	return out
}

// RunOptions controls a preflight invocation.
type RunOptions struct {
	OnlyID          string
	Fix             bool
	PerCheckTimeout time.Duration
}

// Runner runs checks against a CheckEnv and renders results.
type Runner struct {
	Registry *Registry
}

func (r *Runner) Run(ctx context.Context, env CheckEnv, opts RunOptions, w io.Writer) (bool, []Result) {
	var toRun []Check
	if opts.OnlyID != "" {
		c, ok := r.Registry.Get(opts.OnlyID)
		if !ok {
			fmt.Fprintf(w, "unknown check id: %q\n", opts.OnlyID)
			fmt.Fprintln(w, "available checks:")
			ids := r.Registry.IDs()
			sort.Strings(ids)
			for _, id := range ids {
				fmt.Fprintf(w, "  - %s\n", id)
			}
			return false, nil
		}
		toRun = []Check{c}
	} else {
		toRun = r.Registry.All()
	}

	results := make([]Result, 0, len(toRun))
	for _, c := range toRun {
		res := runOne(ctx, env, c, opts.PerCheckTimeout)
		if opts.Fix && res.Status == StatusFail && c.Fix != nil {
			if err := c.Fix(ctx, env); err != nil {
				res.Detail = fmt.Sprintf("%s; fix attempt failed: %v", res.Detail, err)
			} else {
				re := runOne(ctx, env, c, opts.PerCheckTimeout)
				re.Detail = "after --fix: " + re.Detail
				res = re
			}
		}
		results = append(results, res)
	}

	renderTable(w, results)
	allOK := true
	for _, res := range results {
		if res.Status != StatusOK {
			allOK = false
			break
		}
	}
	return allOK, results
}

func runOne(ctx context.Context, env CheckEnv, c Check, timeout time.Duration) Result {
	callCtx := ctx
	if timeout > 0 {
		var cancel context.CancelFunc
		callCtx, cancel = context.WithTimeout(ctx, timeout)
		defer cancel()
	}
	res := c.Run(callCtx, env)
	if res.ID == "" {
		res.ID = c.ID
	}
	return res
}

func renderTable(w io.Writer, results []Result) {
	const (
		idHdr     = "CHECK"
		statusHdr = "STATUS"
		detailHdr = "DETAIL"
	)
	idWidth := len(idHdr)
	for _, r := range results {
		if l := len(r.ID); l > idWidth {
			idWidth = l
		}
	}
	fmt.Fprintf(w, "%-*s  %-6s  %s\n", idWidth, idHdr, statusHdr, detailHdr)
	fmt.Fprintf(w, "%s  %s  %s\n", strings.Repeat("-", idWidth), strings.Repeat("-", 6), strings.Repeat("-", 20))
	for _, r := range results {
		tag := fmt.Sprintf("[%s]", r.Status)
		fmt.Fprintf(w, "%-*s  %-6s  %s\n", idWidth, r.ID, tag, r.Detail)
		if r.Fix != "" && (r.Status == StatusFail || r.Status == StatusWarn) {
			fmt.Fprintf(w, "%-*s  %-6s    fix: %s\n", idWidth, "", "", r.Fix)
		}
	}
}
