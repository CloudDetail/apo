## 1. Config and Service Wiring

- [x] 1.1 Add dataplane-scoped config fields in `backend/config/config.go` and `backend/config/apo.yml` for the ClickHouse lookup budget and instances cache TTL.
- [x] 1.2 Wire the new config into the dataplane service initialization in `backend/pkg/services/dataplane/service.go` without changing the `/api/dataplane/instances` API contract.

## 2. Bounded Lookup and Fallback

- [x] 2.1 Update `backend/pkg/services/dataplane/service_get_service_instances.go` so the ClickHouse app-info lookup runs under the configured time budget while keeping ClickHouse as the preferred source when it returns instances in time.
- [x] 2.2 Extend the same service flow so ClickHouse timeout, error, and empty-result cases enter the fallback decision path, while preserving the existing endpoint schema.
- [x] 2.3 Preserve the existing APO data-plane ownership gate before invoking `promRepo.GetActiveInstanceList()`, so active non-`apo` data planes with `ServiceInstance` capability do not get silently replaced by Prometheus results.
- [x] 2.4 Add a local dataplane helper or cache file under `backend/pkg/services/dataplane/` to store short-lived successful final instance results keyed by `(service, cluster, startTime, endTime)` without caching timeout or error responses.

## 3. Regression Coverage and Verification

- [x] 3.1 Add service-layer regression tests covering ClickHouse fast path, APO-gated empty-result fallback, APO-gated timeout fallback, APO-gated error fallback, non-APO fallback suppression, cache hits, timeout/error non-cache behavior, and cache key isolation.
- [x] 3.2 Run APO backend tests for the dataplane service path and confirm the new control flow passes without changing other dataplane endpoints.
- [x] 3.3 Perform live verification with the representative AgentSRE `findings-data` request and confirm `/api/dataplane/instances` is no longer dominated by the slow ClickHouse wait path.
