## Context

`GET /api/dataplane/instances` is a shared APO backend endpoint that currently prefers ClickHouse app-info data and only falls back to Prometheus when ClickHouse returns an empty result and the existing APO data-plane ownership check allows APO metric fallback. In the dev environment, the ClickHouse query against `originx_app_info` has degraded into a full-table scan, and the same table is also under heavy mutation pressure from `apo-collector` updates to `heart_time` and `heart_flag`.

This endpoint is on the critical path of AgentSRE `findings-data`, which amplifies the latency by calling `/api/dataplane/instances` repeatedly and serially. The immediate need is a narrow owner-side stop-loss that reduces the tail for `/api/dataplane/instances` without mixing in the larger writer-side redesign of `originx_app_info`.

Constraints:

- The endpoint is shared by APO callers, so request/response shape must remain unchanged.
- Existing Prometheus fallback already exists and is materially faster in the affected environment, but it is not a perfect semantic replacement for app-info heartbeat state.
- Prometheus fallback is an APO-only fallback. If an active non-`apo` data plane declares `ServiceInstance`, the endpoint must not silently replace that data-plane ownership with Prometheus results.
- The root cause of the write amplification sits outside APO backend read code: `apo-collector` is issuing the high-frequency `ALTER TABLE originx_app_info UPDATE ...` mutations.

## Goals / Non-Goals

**Goals:**

- Cut the long tail on `GET /api/dataplane/instances` caused by slow ClickHouse lookups.
- Keep ClickHouse as the preferred source when it responds quickly enough.
- Fall back to existing Prometheus instance lookup when ClickHouse times out, errors, or returns no instances and the APO data-plane ownership check allows the fallback.
- Add a short TTL cache so repeated identical successful instance lookups do not repeatedly hit ClickHouse or Prometheus.
- Add regression tests that prove the control flow and cache behavior.

**Non-Goals:**

- Do not redesign `originx_app_info` schema, partitioning, primary key, or latest-state read model in this change.
- Do not modify `apo-collector` writer behavior in this change.
- Do not change the response schema of `GET /api/dataplane/instances`.
- Do not broaden this patch to other dataplane APIs such as `/api/dataplane/topology`, RED, logs, or traces.
- Do not declare Prometheus lookup to be a full semantic replacement for app-info heartbeat state in all scenarios.
- Do not bypass non-`apo` data planes that declare `ServiceInstance` by falling back to Prometheus as if it were the configured data-plane source.

## Decisions

### Decision 1: Keep the endpoint CH-first, but make ClickHouse bounded

`GET /api/dataplane/instances` will continue to prefer ClickHouse so that the existing app-info semantics remain the primary source when ClickHouse is healthy. However, the ClickHouse lookup must run under a small, explicit budget configured for this endpoint.

If the ClickHouse lookup exceeds the budget, the service SHALL stop waiting and move to the fallback decision path. The fallback decision path then checks whether APO metric fallback is allowed before invoking Prometheus.

This is preferred over a permanent Prometheus-only switch because Prometheus currently answers a slightly different question: instances that emitted trace activity within the time window, not necessarily every heartbeat-visible active instance.

Alternatives considered:

- Prometheus-only: faster, but too aggressive a semantic change for a shared API.
- Timeout-only with no fallback: still returns errors or empty results to callers, which does not solve the owner-side stop-loss problem.
- Leave the current logic unchanged and only rely on AgentSRE changes: would continue to expose every caller to the same APO long tail.

### Decision 2: Fallback candidates are timeout, error, or empty result, but Prometheus fallback remains APO-gated

The endpoint SHALL treat these ClickHouse outcomes as fallback candidates:

- ClickHouse lookup exceeds the configured budget
- ClickHouse lookup returns an error
- ClickHouse lookup completes successfully but returns zero instances

Before invoking `promRepo.GetActiveInstanceList()`, the service SHALL preserve the existing APO data-plane ownership gate, currently represented by `dbRepo.CheckGetApoMetric(ctx)`. When that gate allows APO metric fallback, the service SHALL call Prometheus for timeout, error, and empty-result candidates. When that gate does not allow APO metric fallback because a non-`apo` data plane declares `ServiceInstance`, the service SHALL NOT call Prometheus.

For non-APO-owned paths, this change still bounds the ClickHouse wait. However, it must preserve source ownership: empty ClickHouse results remain empty results, and timeout/error cases remain schema-compatible failures rather than silently returning Prometheus instances from the wrong source.

Empty-result fallback is kept for APO-owned paths because the current endpoint already treats Prometheus as a valid secondary source when ClickHouse has no useful result. This change widens the APO-owned fallback trigger set to include operational failure and budget exhaustion.

Alternatives considered:

- Only fallback on timeout: still leaves APO-owned error paths exposed to callers.
- Only fallback on timeout/error, but not empty: would regress existing APO-owned empty-result behavior.
- Always fallback to Prometheus regardless of data-plane ownership: faster, but violates the existing `ServiceInstance` ownership boundary for non-`apo` data planes.

### Decision 3: Cache the final instance list, not the source-specific intermediate result

The cache key will be `(service, cluster, startTime, endTime)`, and the cached value will be the successful final list returned by the endpoint after source selection and fallback. Empty successful lists may be cached for the short TTL, but timeout/error responses SHALL NOT be cached.

The cache SHALL be short-lived and configuration-driven so it acts as a stop-loss against repeated identical calls, not as a long-lived source of truth.

Alternatives considered:

- Cache only ClickHouse results: does not help repeated APO-owned fallback cases.
- Cache only Prometheus results: misses the healthy ClickHouse fast path.
- Cache errors: would turn a transient ClickHouse or Prometheus outage into repeated stale failures.
- No cache: leaves repeated request storms to hit both backends every time.

### Decision 4: Scope the implementation to dataplane service code and a minimal context helper if needed

The main implementation scope for Phase A is:

- `backend/config/config.go`
- `backend/config/apo.yml`
- `backend/pkg/services/dataplane/service.go`
- `backend/pkg/services/dataplane/service_get_service_instances.go`
- a new dataplane cache/helper file
- a new dataplane service test file

If enforcing a ClickHouse budget requires a small helper to derive a bounded context safely, that helper is allowed, but the change must stay local to the dataplane service / repo boundary. It must not expand into a broad refactor of `core.Context`.

Alternatives considered:

- Reusing `core.Context.Clone()` directly: not suitable, because it does not preserve the request-backed Gin context needed for proper cancellation semantics.
- Broad `core.Context` redesign: too large for a stop-loss patch.

### Decision 5: Regression proof must live at the service layer, with live verification after implementation

The automated regression gate for this change belongs at the APO backend service layer, because the behavior change is control-flow orchestration between ClickHouse, Prometheus, and cache. The tests must prove:

- ClickHouse fast path returns ClickHouse results without fallback
- APO-owned empty ClickHouse result falls back to Prometheus
- APO-owned timed-out ClickHouse lookup falls back to Prometheus
- APO-owned errored ClickHouse lookup falls back to Prometheus
- Non-APO `ServiceInstance` ownership suppresses Prometheus fallback
- Cache hits suppress repeated backend calls
- Timeout/error responses are not cached
- Cache key isolation prevents incorrect reuse

In addition, implementation is not considered validated until live verification shows that the representative AgentSRE `findings-data` request is no longer dominated by repeated waits on `/api/dataplane/instances`.

Alternatives considered:

- Only unit-test helper functions: too weak to prove the endpoint control flow.
- Only benchmark live traffic: too coarse to prevent regression in edge paths.

## Risks / Trade-offs

- [Prometheus fallback is not semantically identical to heartbeat app-info state] → Keep CH-first behavior and only use Prometheus on timeout, error, or empty result when the APO ownership gate allows it.
- [Prometheus fallback can violate non-APO data-plane ownership] → Preserve `CheckGetApoMetric()` gating before invoking Prometheus.
- [A short cache can temporarily hide newly appearing instances within the TTL window] → Keep TTL small and configuration-driven.
- [Shared APO callers beyond AgentSRE also use `/api/dataplane/instances`] → Preserve schema and healthy ClickHouse behavior; constrain the behavioral change to degraded ClickHouse paths.
- [Adding a budgeted sub-context could leak into broader core abstractions] → Keep any helper local to this endpoint path and do not redesign `core.Context` in this patch.
- [This stop-loss does not remove the true root cause in `originx_app_info`] → Record collector / ClickHouse current-state redesign as a separate follow-up rather than overloading this patch.

## Migration Plan

1. Add dataplane-scoped config for ClickHouse lookup budget and cache TTL.
2. Implement bounded ClickHouse lookup orchestration in `/api/dataplane/instances`.
3. Preserve the existing APO data-plane ownership gate before invoking Prometheus fallback.
4. Add short TTL caching for successful final instance results.
5. Add service-layer regression tests for fast path, APO-gated fallback paths, non-APO fallback suppression, cache hit, error non-cache behavior, and cache key isolation.
6. Deploy the APO backend change without changing endpoint schema.
7. Re-run the representative AgentSRE `findings-data` request and collect before/after evidence from:
   - APO `system.query_log`
   - AgentSRE logs
   - total request wall time

Rollback strategy:

- Disable the new behavior by setting the ClickHouse budget or cache TTL config to the safe values defined during implementation, or revert the APO backend patch.
- Because schema is unchanged, rollback does not require consumer-side compatibility handling.

## Open Questions

- What exact default values should APO use for ClickHouse budget and cache TTL in `apo.yml`?
- Should timed-out ClickHouse lookups be logged as warnings with a distinct marker so live verification can attribute fallback decisions quickly?
- Which APO callers, if any, require a stronger guarantee than the current APO-gated fallback semantics already provide?
