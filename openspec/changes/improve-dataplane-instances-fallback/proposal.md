## Why

`/api/dataplane/instances` currently blocks on a ClickHouse lookup against `originx_app_info` before it even considers the existing APO-gated Prometheus fallback. In the dev environment this lookup has degraded to tens or hundreds of seconds, and AgentSRE amplifies that latency by calling the endpoint repeatedly and serially during `findings-data` assembly.

The owner-side mitigation belongs in APO first. We need a narrow stop-loss change that cuts the request tail on `/api/dataplane/instances` without mixing in the larger collector / ClickHouse current-state redesign.

## What Changes

- Add a bounded ClickHouse lookup budget for `/api/dataplane/instances`, while keeping the endpoint schema unchanged.
- Keep ClickHouse as the preferred source, but treat ClickHouse timeout, error, or empty result as fallback candidates. The Prometheus fallback MUST remain gated by the existing APO data-plane ownership check and must not become a universal instance source when a non-`apo` data plane owns `ServiceInstance`.
- Add a short TTL cache for successful `/api/dataplane/instances` final results, keyed by `(service, cluster, startTime, endTime)`, so repeated lookups do not keep re-hitting ClickHouse or Prometheus within the cache window.
- Add service-layer regression tests for ClickHouse fast-path, APO-gated empty-result fallback, APO-gated timeout fallback, APO-gated error fallback, non-APO fallback suppression, cache hits, error non-cache behavior, and cache key isolation.
- Record explicit non-goals: this change does not redesign `originx_app_info`, does not change `apo-collector`, and does not alter other dataplane APIs.

## Capabilities

### New Capabilities
- `dataplane-service-instances-lookup`: Defines the bounded lookup, fallback, and cache behavior for `GET /api/dataplane/instances`.

### Modified Capabilities

None.

## Impact

- API:
  - `GET /api/dataplane/instances`
- Code:
  - `backend/config/config.go`
  - `backend/config/apo.yml`
  - `backend/pkg/services/dataplane/service.go`
  - `backend/pkg/services/dataplane/service_get_service_instances.go`
  - a new dataplane instances cache/helper file under `backend/pkg/services/dataplane/`
  - a new dataplane service regression test file under `backend/pkg/services/dataplane/`
- Dependencies:
  - ClickHouse `originx_app_info`
  - Prometheus / VictoriaMetrics instance lookup via `promRepo.GetActiveInstanceList()`
- Systems:
  - APO backend request latency for `/api/dataplane/instances`
  - upstream callers such as AgentSRE `findings-data`
  - other direct `/api/dataplane/instances` consumers such as Dify APO Select tooling
- Explicitly out of scope:
  - `apo-collector` writer behavior
  - `originx_app_info` table redesign, partitioning, or latest-state model changes
  - using Prometheus as a universal replacement for non-`apo` data planes that declare `ServiceInstance`
  - non-`/api/dataplane/instances` dataplane endpoints
