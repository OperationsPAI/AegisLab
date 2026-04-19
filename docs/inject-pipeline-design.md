# Fault Injection Pipeline — Design (2026-04-18)

**Supersedes** the retired `chaoscli.Spec` end-to-end design (same file,
same date). The `chaoscli` package was deleted; `pkg/guidedcli` is now the
canonical data model.

## Wire format

`POST /api/v2/projects/{id}/injections/inject` body is `dto.SubmitInjectionReq`:

- `pedestal`, `benchmark`, `interval`, `pre_duration`, `algorithms`
- `Specs [][]json.RawMessage` — outer = batches, inner = homogeneous specs

Each raw-message element is one of three shapes:

| Shape | Identified by | Status |
|---|---|---|
| **Guided** | `chaos_type` field present | preferred (new) |
| **Node DSL** | chaos-experiment Node tree (int-indexed children) | legacy |
| **FriendlyFaultSpec** | `type` + `target` + `params` strings | legacy |

See `src/dto/injection.go` (`SubmitInjectionReq`, `ResolveSpecs`).

## Flow

```
HTTP POST /injections/inject
    │
    ▼
SubmitInjectionReq.ResolveSpecs            (three-way dispatch by chaos_type probe)
    │                         │
    │ guided                  │ legacy (Node | Friendly)
    ▼                         ▼
ResolvedGuidedConfigs         resolved Node trees
    │                         │
    └──────────┬──────────────┘
               ▼
       RestartPedestal task (Redis task:delayed)
               │
               ▼
       FaultInjection subtask
               │
    ┌──────────┴─────────────┐
    │ guided_configs in      │ node payload
    │ payload                │
    ▼                        ▼
guidedcli.BuildInjection    chaos.NodeToStruct[handler.InjectionConf]
    │                        │
    └──────────┬─────────────┘
               ▼
       handler.BatchCreate
               │
               ▼
       Chaos Mesh CRD (PodChaos / NetworkChaos / …)
```

## File map

- `pkg/guidedcli/` (chaos-experiment, public) — `GuidedConfig`, `Resolve`,
  `BuildInjection(ctx, cfg) (handler.InjectionConf, handler.SystemType, error)`
- `src/dto/injection.go` — `SubmitInjectionReq`, `ResolveSpecs` (three-way),
  `ResolvedGuidedConfigs [][]guidedcli.GuidedConfig`
- `src/consts/` — `InjectGuidedConfigs = "guided_configs"` payload key
- `src/service/consumer/fault_injection.go` — `parseInjectionPayload`
  branches guided vs. legacy; guided path calls
  `guidedcli.BuildInjection`, legacy calls `chaos.NodeToStruct`
- `src/cmd/aegisctl/cmd/inject.go` — `aegisctl inject guided` subcommand
  (mirrors `cmd/chaos-exp/main.go` loop; wraps `GuidedConfig` in
  `SubmitInjectionReq` on `--apply`)
- `src/handlers/v2/injections.go` — `POST /translate` and
  `GET /injections/metadata` return **410 Gone** (routes kept)

## Landmines

1. **`chaos_type` probe is the dispatch key.** If a new spec shape adds a
   field named `chaos_type` unintentionally, it will be routed to the
   guided branch. Keep that field name guided-only.
2. **Batches must be homogeneous.** `ResolveSpecs` does not support mixing
   guided + legacy within one inner `[]json.RawMessage`.
3. **Legacy still compiles and runs.** `dto.FaultSpec`,
   `dto.FriendlyFaultSpec`, `producer.FriendlySpecToNode`,
   `parseBatchInjectionSpecs` are kept alive for frontend back-compat.
   Do not extend them — new work goes on the guided path.
4. **No Node round-trip on the guided path.** There is no
   `InjectionConfToNode`; `BuildInjection` builds `handler.InjectionConf`
   directly from live cluster state. idx drift between submit and
   consume time is therefore a non-issue — the consumer always resolves
   against its current view.
5. **`/translate` + `GET /metadata` return 410.** Routes exist but refuse
   to serve; frontend migration is deferred but the endpoints signal
   deprecation loudly.

## Migration status (2026-04-18)

- [x] PR 1 (chaos-experiment): `pkg/guidedcli` exported, `pkg/chaoscli`
      deleted, `cmd/chaos-exp/main.go` restored, `BuildInjection` added
- [x] PR 2 (AegisLab): `aegisctl inject guided`, three-way
      `ResolveSpecs`, consumer branch, 410 on deprecated endpoints
- [ ] Frontend migration off `/translate` and `GET /metadata`
- [ ] Delete `FriendlyFaultSpec` / `FriendlySpecToNode` /
      `parseBatchInjectionSpecs` (after frontend cuts over)

## Related

- Troubleshooting runbook: `aegis/docs/troubleshooting/e2e-cluster-bootstrap.md`
- OperationsPAI/aegis#23 (original refactor), #36–#40 (this migration)
