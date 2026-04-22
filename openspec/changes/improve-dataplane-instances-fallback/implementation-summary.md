## Final Scope

This change implemented a narrow owner-side stop-loss for `GET /api/dataplane/instances` in APO backend. The delivered behavior is:

- keep ClickHouse app-info as the preferred source
- run the ClickHouse lookup under a dataplane-scoped time budget
- treat timeout, error, and empty result as fallback candidates
- preserve the existing APO ownership gate before invoking Prometheus fallback
- cache successful final instance results for a short TTL

This change did not redesign `originx_app_info`, did not modify `apo-collector`, and did not broaden the behavior change to other dataplane endpoints.

## Actual Code Changes

- Added dataplane config fields in `backend/config/config.go` and defaults in `backend/config/apo.yml` for:
  - `service_instances_lookup_budget_seconds`
  - `service_instances_cache_ttl_seconds`
- Added `backend/pkg/services/dataplane/service_instances_lookup.go` to host:
  - the bounded ClickHouse lookup orchestration
  - the APO-gated Prometheus fallback decision
  - a local request-scoped cache key model and success-only TTL cache
  - write-path expiry sweeping so stale request-window keys do not accumulate until reread
  - a lightweight bounded `core.Context` wrapper for ClickHouse timeout control
- Updated `backend/pkg/services/dataplane/service.go` to initialize the new helper from config without changing external constructor signatures.
- Updated `backend/pkg/services/dataplane/service_get_service_instances.go` to delegate to the helper.
- Added `backend/pkg/services/dataplane/service_get_service_instances_test.go` with service-layer regression coverage for:
  - ClickHouse fast path
  - APO-gated empty/error/timeout fallback
  - non-APO fallback suppression
  - cache hit behavior
  - timeout/error non-cache behavior
  - cache key isolation
  - cache expiry cleanup on later writes without rereading the expired key

## Document Updates

- Updated change docs:
  - `openspec/changes/improve-dataplane-instances-fallback/tasks.md`
  - `openspec/changes/improve-dataplane-instances-fallback/implementation-summary.md`
- Updated system docs:
  - `docs/openspec-system/modules/dataplane-and-topology-governance/backend.md`

## Design Deviations

- No intentional design deviation from the latest change docs.
- Implementation kept the existing non-APO ownership behavior conservative:
  - empty ClickHouse results remain empty results
  - timeout/error remain schema-compatible failures
  - Prometheus is never substituted as a universal instance source

## Compatibility And Rollout

- The `GET /api/dataplane/instances` request/response schema remains unchanged.
- Healthy ClickHouse paths still return ClickHouse-derived instances first.
- Rollout depends only on APO backend deployment plus config availability in `apo.yml`.
- No OSS sync is included in this implementation round per current user direction.

## Live Verification Evidence

- Representative request was re-run through a temporary local AgentSRE instance pointed at the patched commercial APO backend code:
  - `GET http://127.0.0.1:8001/api/v1/conversations/findings-data?threadId=thread_1776677738509_8bhq8bym9&serviceName=ts-travel-service`
  - temporary AgentSRE config used `apo_backend_url=http://127.0.0.1:18080`
  - temporary APO backend config started the patched binary on `:18080`
- The patched APO backend startup log confirmed the intended runtime config:
  - `dataplane instances lookup config: lookupBudget=2s cacheTTL=10s`
- During the live request, APO backend logs showed repeated `GET /api/dataplane/instances` calls no longer waiting on unbounded ClickHouse latency:
  - one request hit the ClickHouse budget and returned in `2.05065875s` after `context deadline exceeded`, then fell back to Prometheus
  - one request hit a ClickHouse read timeout and returned in `2.0367157909999998s`, then fell back to Prometheus
  - one request hit the ClickHouse budget again and returned in `2.036769s`, then fell back to Prometheus
  - one request saw ClickHouse return empty in `366ms` and completed the full `/api/dataplane/instances` path in `0.464286709s` via Prometheus fallback
- The representative `findings-data` HTTP request completed with `200 OK` in `94.590438s`.
- Interpretation:
  - this change's scoped goal is met because the request is no longer dominated by minute-class ClickHouse waits inside APO `/api/dataplane/instances`
  - the remaining tail in this run stayed in AgentSRE-side runtime/LLM projection work, so this verification must not be misreported as a full `findings-data` SLA fix

## Follow-ups

- Default values `lookup budget = 2s` and `cache TTL = 10s` were implemented as approved working defaults; adjust later if production evidence suggests different thresholds.
- Local verification required a temporary AgentSRE config because the current long-running local AgentSRE dev process was pointed at `apo_backend_url=http://127.0.0.1:8000`, which would not have exercised the patched APO backend instance.
- If rollout evidence shows repeated suppressed fallbacks in non-APO ownership scenarios, evaluate whether additional observability or caller guidance is needed rather than broadening fallback semantics.
