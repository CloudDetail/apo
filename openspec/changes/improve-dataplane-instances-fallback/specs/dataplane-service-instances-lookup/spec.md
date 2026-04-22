## ADDED Requirements

### Requirement: Bounded ClickHouse Lookup for Service Instances
The `GET /api/dataplane/instances` capability SHALL keep ClickHouse app-info as its preferred source, but the ClickHouse lookup MUST execute under an explicit, dataplane-scoped time budget. When the lookup completes successfully within that budget and returns one or more instances, the endpoint SHALL return the ClickHouse-derived instance list without invoking the Prometheus fallback path.

#### Scenario: ClickHouse returns instances within budget
- **WHEN** `/api/dataplane/instances` queries ClickHouse within the configured lookup budget and the query returns one or more instances
- **THEN** the service SHALL return the ClickHouse instance list
- **AND** the service SHALL NOT call `promRepo.GetActiveInstanceList()` for that request

#### Scenario: ClickHouse exceeds the configured budget
- **WHEN** `/api/dataplane/instances` starts a ClickHouse lookup and the lookup does not complete before the configured budget expires
- **THEN** the service SHALL stop waiting on the ClickHouse result for that request
- **AND** the service SHALL continue request processing through the APO data-plane fallback decision path

### Requirement: Prometheus Fallback on ClickHouse Degradation
The `GET /api/dataplane/instances` capability SHALL treat ClickHouse timeout, error, or empty result as fallback candidates. It SHALL invoke the existing `promRepo.GetActiveInstanceList()` path only when the APO data-plane ownership check allows APO metric fallback. The endpoint response schema MUST remain unchanged while this fallback behavior is added.

#### Scenario: ClickHouse returns no instances
- **WHEN** the ClickHouse lookup completes successfully but returns an empty instance list
- **AND** the APO data-plane ownership check allows APO metric fallback
- **THEN** the service SHALL call `promRepo.GetActiveInstanceList()`
- **AND** the service SHALL return the resulting Prometheus-derived instance list using the existing endpoint response schema

#### Scenario: ClickHouse returns an error
- **WHEN** the ClickHouse lookup fails with an error before producing a usable instance list
- **AND** the APO data-plane ownership check allows APO metric fallback
- **THEN** the service SHALL call `promRepo.GetActiveInstanceList()`
- **AND** the service SHALL return the resulting Prometheus-derived instance list using the existing endpoint response schema

#### Scenario: ClickHouse times out and APO fallback is allowed
- **WHEN** the ClickHouse lookup exceeds the configured budget
- **AND** the APO data-plane ownership check allows APO metric fallback
- **THEN** the service SHALL call `promRepo.GetActiveInstanceList()`
- **AND** the service SHALL return the resulting Prometheus-derived instance list using the existing endpoint response schema

#### Scenario: Non-APO ServiceInstance data plane suppresses Prometheus fallback
- **WHEN** the ClickHouse lookup times out, returns an error, or returns no instances
- **AND** an active non-`apo` data plane declares the `ServiceInstance` capability
- **THEN** the service SHALL NOT call `promRepo.GetActiveInstanceList()`
- **AND** the service SHALL return using the existing endpoint response schema without substituting Prometheus instances for the non-`apo` data-plane source

#### Scenario: Fallback does not change the API contract
- **WHEN** the service uses the Prometheus fallback path because ClickHouse timed out, errored, or returned no instances
- **THEN** the `GET /api/dataplane/instances` request and response schema SHALL remain unchanged for callers

### Requirement: Short-Lived Final Result Cache
The `GET /api/dataplane/instances` capability SHALL cache successful final returned instance lists for a short, configuration-driven TTL. The cache key MUST include `(service, cluster, startTime, endTime)` so repeated identical lookups can reuse the last result without re-querying ClickHouse or Prometheus during the cache window. Timeout and error responses MUST NOT be cached.

#### Scenario: Repeated identical lookup hits cache
- **WHEN** two or more `/api/dataplane/instances` requests use the same `service`, `cluster`, `startTime`, and `endTime` within the configured cache TTL
- **THEN** the service SHALL return the cached final instance list for the later requests
- **AND** the service SHALL NOT repeat the ClickHouse or Prometheus lookup while the cache entry remains valid

#### Scenario: Error response is not cached
- **WHEN** a `/api/dataplane/instances` request returns a timeout or error response that cannot use APO fallback
- **THEN** the service SHALL NOT store that response in the instances cache
- **AND** a later identical request SHALL execute the normal lookup and fallback decision flow again

#### Scenario: Different lookup parameters do not reuse cache
- **WHEN** `/api/dataplane/instances` receives a request whose `service`, `cluster`, `startTime`, or `endTime` differs from an existing cache entry
- **THEN** the service SHALL treat it as a different cache key
- **AND** the service SHALL execute the normal lookup and fallback flow for that request
