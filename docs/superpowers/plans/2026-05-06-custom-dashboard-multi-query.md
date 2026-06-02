# Custom Dashboard Multi Query Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Support multiple custom PromQL query expressions in one custom dashboard widget and render all returned series in the same chart, table, stat, or bar gauge widget.

**Architecture:** Extend the existing widget query contract with a `queries` array while preserving `expression` and `legend` as the first query mirror for compatibility. Frontend preview requests and backend preview handling will normalize old and new shapes, execute each custom query independently, and merge returned series.

**Tech Stack:** React, Ant Design, Node test runner, Go service tests, existing custom dashboard API.

---

### Task 1: Frontend Query Normalization And Request Payload

**Files:**
- Modify: `frontend/src/pro/views/CustomDashboard/utils.mjs`
- Modify: `frontend/src/pro/views/CustomDashboard/behavior.mjs`
- Test: `frontend/src/pro/views/CustomDashboard/behavior.test.mjs`

- [x] **Step 1: Write failing frontend tests**

Add tests that assert:

```js
const dashboard = normalizeDashboard({
  widgets: [{ id: 'w-1', query: { mode: 'custom', expression: 'up', legend: 'Up' } }],
})
assert.deepEqual(dashboard.widgets[0].query.queries, [
  { id: 'q-1', expression: 'up', legend: 'Up' },
])

const request = buildPreviewRequest({
  widget: {
    id: 'w-1',
    type: 'line',
    query: {
      mode: 'custom',
      queries: [
        { id: 'q-a', expression: 'up{svc_name="old"}', legend: 'A' },
        { id: 'q-b', expression: 'rate(http_requests_total{svc="old"}[5m])', legend: 'B' },
      ],
    },
  },
  variables: { service: ['checkout'] },
  variableDefinitions: [{ name: 'service', type: 'service' }],
  timeRange: { startTime: 1000, endTime: 2000 },
})
assert.deepEqual(request.queries.map((query) => query.expression), [
  'up{svc_name=~"checkout"}',
  'rate(http_requests_total{svc=~"checkout"}[5m])',
])
```

- [x] **Step 2: Run failing frontend tests**

Run: `cd frontend && node --test src/pro/views/CustomDashboard/behavior.test.mjs`
Expected: failure because `query.queries` and request `queries` are not implemented.

- [x] **Step 3: Implement frontend normalization**

Add helpers in `utils.mjs` to normalize custom query items, mirror the first query to `expression` and `legend`, and ensure at least one query exists for custom mode.

- [x] **Step 4: Implement frontend request generation**

Update `buildPreviewRequest` and query-text helpers in `behavior.mjs` so variable discovery and preview payloads use all query items.

- [x] **Step 5: Run frontend tests**

Run: `cd frontend && node --test src/pro/views/CustomDashboard/behavior.test.mjs`
Expected: all tests pass.

### Task 2: Editor Multi Query Controls

**Files:**
- Modify: `frontend/src/pro/views/CustomDashboard/Editor.jsx`
- Modify: `frontend/public/locales/en/pro/customDashboard.json`
- Modify: `frontend/public/locales/zh/pro/customDashboard.json`
- Test: `frontend/src/pro/views/CustomDashboard/behavior.test.mjs`

- [x] **Step 1: Write failing editor source tests**

Assert the editor exposes add/delete query controls, renders one textarea per custom query, and disables deletion for the last query.

- [x] **Step 2: Run failing frontend tests**

Run: `cd frontend && node --test src/pro/views/CustomDashboard/behavior.test.mjs`
Expected: failure because editor controls do not exist.

- [x] **Step 3: Implement editor controls**

Add `addQuery`, `deleteQuery`, and `updateQueryItem` handlers. In custom mode, render each query item with expression and legend fields. Keep the existing preset mode unchanged.

- [x] **Step 4: Run frontend tests**

Run: `cd frontend && node --test src/pro/views/CustomDashboard/behavior.test.mjs`
Expected: all tests pass.

### Task 3: Stat And Bar Gauge Rendering

**Files:**
- Modify: `frontend/src/pro/views/CustomDashboard/utils.mjs`
- Modify: `frontend/src/pro/views/CustomDashboard/WidgetRenderer.jsx`
- Modify: `frontend/src/pro/views/CustomDashboard/index.scss`
- Modify: `frontend/public/locales/en/pro/customDashboard.json`
- Modify: `frontend/public/locales/zh/pro/customDashboard.json`
- Test: `frontend/src/pro/views/CustomDashboard/behavior.test.mjs`

- [x] **Step 1: Write failing frontend tests**

Assert `chartTypes` contains `stat` and `barGauge`, `WidgetRenderer.jsx` contains stat and bar gauge render branches, and formatting uses the last series value.

- [x] **Step 2: Run failing frontend tests**

Run: `cd frontend && node --test src/pro/views/CustomDashboard/behavior.test.mjs`
Expected: failure because the new widget types and render branches do not exist.

- [x] **Step 3: Implement rendering**

Add stat cards and bar gauge rows. For both types, use each returned series' last point, `formatWidgetValue`, and `seriesDisplayName`. Clamp bar gauge progress to `0-100` but keep the displayed value unchanged.

- [x] **Step 4: Run frontend tests**

Run: `cd frontend && node --test src/pro/views/CustomDashboard/behavior.test.mjs`
Expected: all tests pass.

### Task 4: Backend Multi Query Preview

**Files:**
- Modify: `backend/pkg/model/customdashboard/dashboard.go`
- Modify: `backend/pkg/model/request/custom_dashboard.go`
- Modify: `backend/pkg/services/customdashboard/service.go`
- Test: `backend/pkg/services/customdashboard/service_test.go`

- [x] **Step 1: Write failing backend tests**

Add tests that call `PreviewQuery` with two custom query items and assert two Prometheus queries are executed and returned series are merged. Add a syntax-error test for the second query.

- [x] **Step 2: Run failing backend tests**

Run: `cd backend && go test -count=1 ./pkg/services/customdashboard`
Expected: failure because request `Queries` and service multi-query execution are not implemented.

- [x] **Step 3: Implement backend models and normalization**

Add `QueryItem`/`WidgetQueryItem` structs with `id`, `expression`, and `legend`. Normalize old `Expression` into the first query.

- [x] **Step 4: Implement backend execution**

For custom mode, resolve all query items, validate each expression, enforce data scope on each, execute each, attach query legend metadata, and merge series.

- [x] **Step 5: Run backend tests**

Run: `cd backend && go test -count=1 ./pkg/services/customdashboard`
Expected: all tests pass.

### Task 5: Full Verification

**Files:**
- No new source files.

- [x] **Step 1: Run custom dashboard frontend tests**

Run: `cd frontend && node --test src/pro/views/CustomDashboard/behavior.test.mjs`
Expected: all tests pass.

- [x] **Step 2: Run navigation behavior tests**

Run: `cd frontend && node --test src/core/components/AppBreadcrumbBehavior.test.mjs src/core/components/AppSidebarBehavior.test.mjs`
Expected: all tests pass.

- [x] **Step 3: Run backend tests**

Run: `cd backend && go test -count=1 ./pkg/api/customdashboard ./pkg/services/customdashboard ./pkg/repository/database ./pkg/services/permission`
Expected: all packages pass.

- [x] **Step 4: Run frontend production build**

Run: `cd frontend && npm run build`
Expected: build succeeds; existing Sass/CoreUI and chunk-size warnings may remain.
