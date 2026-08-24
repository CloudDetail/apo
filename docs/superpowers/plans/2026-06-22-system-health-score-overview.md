# System Health Score Overview Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the `/healthy-service` system health scoring page with real backend data where each system is an `All` first-level data group.

**Architecture:** Add a new read-only overview API under `/api/system-health/overview` while preserving the existing report APIs. The backend resolves visible first-level data groups, reuses existing RYG service health metrics per group, and returns a UI-shaped response with fixed seven health metric keys. The frontend hides the global data group selector for `/healthy-service`, replaces mock data with API calls, and renders loading, empty, and error states.

**Tech Stack:** Go/Gin backend, existing `systemhealth` API package, existing `serviceoverview` RYG service, React/Ant Design frontend.

---

### Task 1: Backend Response Contract And Pure Aggregation

**Files:**
- Modify: `backend/pkg/model/request/system_health.go`
- Modify: `backend/pkg/model/response/system_health.go`
- Create: `backend/pkg/api/systemhealth/func_getoverview.go`
- Test: `backend/pkg/api/systemhealth/func_getoverview_test.go`

- [x] **Step 1: Write failing aggregation tests**

Create tests that call pure functions `firstLevelSystemGroups` and `buildSystemHealthSystem`. The tests must prove that nested groups are not treated as systems and that service RYG detail scores map into fixed health metrics.

Run: `cd backend && go test ./pkg/api/systemhealth`

Expected: FAIL with undefined helper functions.

- [x] **Step 2: Add request and response structs**

Add query request structs for overview/detail and response structs:

```go
type SystemHealthOverviewResp struct {
    Summary     SystemHealthOverviewSummary `json:"summary"`
    Systems     []SystemHealthSystem        `json:"systems"`
    Window      string                      `json:"window"`
    GeneratedAt string                      `json:"generatedAt"`
    Source      string                      `json:"source"`
}
```

- [x] **Step 3: Implement pure aggregation helpers**

Implement helper behavior:

```text
firstLevelSystemGroups(root) => root.SubGroups sorted by group_id
buildSystemHealthSystem(group, services) => response.SystemHealthSystem
```

Mapping:

```text
replica       -> resource
errorRate     -> errorRate
latency       -> latency
logErrorCount -> errorLog
alert         -> badRequests
```

Additional metrics:

```text
resourceShift -> cadvisor CPU/memory service-level 24h baseline shift
lockCost      -> Polaris futex lock wait cost per request
```

Metrics still return `gray` when the required service-level baseline or join data is unavailable.

- [x] **Step 4: Run backend package tests**

Run: `cd backend && go test ./pkg/api/systemhealth`

Expected: PASS.

### Task 2: Backend API And Routing

**Files:**
- Modify: `backend/pkg/api/systemhealth/handler.go`
- Modify: `backend/pkg/api/systemhealth/func_getoverview.go`
- Modify: `backend/pkg/router/router_api.go`

- [x] **Step 1: Add handler methods**

Add:

```go
GetOverview() core.HandlerFunc
GetSystemDetail() core.HandlerFunc
```

- [x] **Step 2: Implement API handlers**

`GET /api/system-health/overview` returns all visible first-level data-group systems.

`GET /api/system-health/overview/:groupId` returns one visible system or an error if not visible.

- [x] **Step 3: Register routes**

Register before legacy report routes:

```go
systemHealthApi.GET("/overview", systemHealthHandler.GetOverview())
systemHealthApi.GET("/overview/:groupId", systemHealthHandler.GetSystemDetail())
```

- [x] **Step 4: Run backend package tests**

Run: `cd backend && go test ./pkg/api/systemhealth`

Expected: PASS.

### Task 3: Frontend API Client And Route Behavior

**Files:**
- Modify: `frontend/src/pro/api/serviceHealthy.js`
- Modify: `frontend/src/pro/routes.js`

- [x] **Step 1: Add API client functions**

Add:

```js
export const getSystemHealthOverviewApi = (params) => get('/api/system-health/overview', params)
export const getSystemHealthSystemDetailApi = (groupId, params) => get(`/api/system-health/overview/${groupId}`, params)
```

- [x] **Step 2: Hide global data group selector**

Remove `showDataGroup: true` from the `/healthy-service` route.

### Task 4: Frontend Data Integration

**Files:**
- Modify: `frontend/src/pro/views/systemHealthOverview/index.jsx`

- [x] **Step 1: Remove mock data as runtime source**

The page must not render the hard-coded `systems` array.

- [x] **Step 2: Load overview from API**

Use current global time range as query input and default to `relativeTime=15m` when no valid time range is available.

- [x] **Step 3: Load detail from API**

When a system row is clicked, request `/api/system-health/overview/:groupId` and render the returned system.

- [x] **Step 4: Add UI states**

Render loading, empty, and error states without falling back to mock systems.

### Task 5: Verification

**Files:**
- No code files created.

- [x] **Step 1: Format backend**

Run: `gofmt -w backend/pkg/model/request/system_health.go backend/pkg/model/response/system_health.go backend/pkg/api/systemhealth/handler.go backend/pkg/api/systemhealth/func_getoverview.go backend/pkg/api/systemhealth/func_getoverview_test.go backend/pkg/router/router_api.go`

- [x] **Step 2: Run backend tests**

Run: `cd backend && go test ./pkg/api/systemhealth`

Expected: PASS.

- [x] **Step 3: Run frontend build or lint**

Run the repo-supported frontend verification command after inspecting `frontend/package.json`.

Expected: command completes without new errors from this change.
