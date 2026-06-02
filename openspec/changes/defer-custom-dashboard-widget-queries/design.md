## Context

The custom dashboard viewer and editor currently load dashboard metadata and preset queries, initialize variables, and then refresh widget previews from the page. Existing refresh paths can build each widget preview request and dispatch all selected widgets with `Promise.all`. This is simple, but a dashboard with many widgets can issue a large burst of metric preview requests on initial open, time range changes, variable changes, data group changes, and manual refresh.

The backend already enforces query timeout, point count, concurrency, resource limits, data scope, and audit behavior. This change should reduce avoidable frontend request bursts before they reach the backend while keeping backend enforcement authoritative.

The related `align-custom-dashboard-resource-filters` change adds selected data group context to custom dashboard candidate and preview requests. This design treats that context as part of each preview request signature when present, the same way time range and dashboard variables are already part of stale-response protection.

## Goals / Non-Goals

**Goals:**

- Improve first-screen responsiveness by querying visible and near-visible widgets before offscreen widgets.
- Avoid dashboard-level query bursts by capping concurrent widget preview requests from a single viewer.
- Preserve correct refresh semantics when global time range, data group, dashboard variables, or manual refresh changes the request context.
- Keep each widget independently understandable with loading, deferred, stale, success, no-data, and error states.
- Keep fullscreen widgets correct even when their grid cell was not previously loaded.

**Non-Goals:**

- Change saved dashboard schemas, widget schemas, or backend preview query API contracts.
- Replace backend query concurrency and resource enforcement.
- Lazy-load dashboard metadata, preset query definitions, or variable candidate requests.
- Add dashboard-level automatic refresh intervals.
- Virtualize or unmount offscreen widget DOM as part of this change.

## Decisions

### Track viewport eligibility in the dashboard page, not in the renderer

The dashboard viewer/editor owns widget layout, refresh triggers, request signatures, and user actions, so the page should own viewport tracking and scheduling. Each widget shell will register with an `IntersectionObserver`; the renderer stays focused on displaying the current widget state and preview data.

Widgets are eligible for query dispatch when they are visible, within a preload margin before entering view, manually prioritized, or opened in fullscreen. A preload margin around 400-600px is a reasonable starting point because it hides most scroll latency without loading the whole dashboard. If viewport observation is unavailable, the viewer treats all widgets as eligible but still dispatches them through the dashboard-local scheduler rather than restoring eager `Promise.all` behavior.

Alternative considered: put lazy loading inside `WidgetRenderer`. That makes each widget self-contained, but it splits refresh invalidation and request dedupe across many component instances and makes manual refresh harder to reason about.

### Keep layout stable and defer data, not panels

All widget shells should still render in the dashboard grid so the user can scan the dashboard structure and scrolling remains stable. Lazy loading applies to preview data requests, not to dashboard layout or widget chrome.

Deferred widgets that have never loaded should show a lightweight pending state rather than a no-data state. Widgets with a previous successful preview can keep showing that data while marked stale until their next scheduled request finishes.

Alternative considered: do full DOM virtualization. That could reduce rendering cost, but it is a larger layout change and risks breaking chart sizing, legends, fullscreen entry points, and drag/editor assumptions.

### Use a dashboard-local preview scheduler

The viewer should route widget preview requests through a small dashboard-local scheduler. The scheduler maintains a queue of latest widget request signatures and runs a default maximum of 4 preview requests concurrently for one dashboard viewer. It should prioritize fullscreen widgets, visible widgets, then near-visible widgets.

For each widget, only the newest request signature matters. If the same widget is queued again with a newer signature, the older queued work is replaced. If an older in-flight response returns, the existing request signature guard ignores it. If a widget already has a successful or no-data preview for the current request signature and is not stale, entering the viewport again does not schedule a duplicate query.

The scheduler is scoped to one mounted dashboard page instance. When the dashboard id changes or the page unmounts, queued work is cleared and in-flight responses are ignored; requests are cancelled with `AbortController` where the existing request layer supports cancellation.

Alternative considered: rely only on `IntersectionObserver`. That prevents initial offscreen queries, but fast scrolling can still enqueue many requests together and create a second burst.

### Treat context changes as invalidation, not immediate full refresh

When the global time range, selected data group, dashboard variables, or manual refresh changes the request context, the viewer/editor should mark affected widgets stale. Visible and near-visible affected widgets are scheduled immediately. Offscreen affected widgets remain stale and are scheduled when they become eligible.

Variable changes should invalidate only widgets that use the changed variable when that can be determined, preserving the existing targeted refresh behavior. Manual refresh should invalidate every widget, but still dispatch through the scheduler instead of bypassing it.

Alternative considered: keep manual refresh as an eager all-widget refresh. That preserves the old mental model but reintroduces the request burst this change is meant to remove.

### Fullscreen gets priority

Opening a widget in fullscreen should immediately prioritize that widget's preview request if the current preview is missing or stale. The fullscreen modal can show the same loading or error state as the grid widget while the request runs.

Alternative considered: require the user to scroll the widget into view before fullscreen can load it. That is surprising because fullscreen is an explicit user action on the widget.

### Retry is widget-scoped

Retry should be scoped to the failed widget. Retrying a widget reuses the current dashboard request context, bypasses the normal "already failed" idle state for that widget, and schedules that widget through the same priority queue without refreshing unrelated widgets.

Alternative considered: rely only on manual refresh for retry. That is simpler, but it forces a whole-dashboard invalidation for one failed panel and works against the request-burst reduction goal.

### Keep the backend contract unchanged

The frontend scheduler reduces avoidable pressure, but it does not weaken backend enforcement. Preview requests still use the same endpoint, same request payload semantics, same data-scope enforcement, same audit behavior, and same query result states.

Alternative considered: add a batch preview endpoint. That could reduce HTTP overhead, but it would need new backend scheduling, partial-failure semantics, and audit decisions. It is unnecessary for this request.

## Risks / Trade-offs

- Offscreen widgets may not have fresh data when a user jumps far down the page -> Use preload margins and priority scheduling so newly eligible widgets start quickly.
- Users may expect manual refresh to start every widget request immediately -> Define manual refresh as invalidating all widgets while dispatch still follows viewport-aware scheduling.
- `IntersectionObserver` may be unavailable or behave unexpectedly in older embedded browsers -> Fall back to scheduling all widgets through the concurrency-limited queue.
- Fast scroll can make queued work stale before it starts -> Keep one latest queued request per widget and skip work whose request signature no longer matches.
- Re-entering the viewport can produce duplicate requests for already loaded widgets -> Store the last completed request signature per widget and schedule only when the current signature is missing, stale, failed with user retry, or changed.
- Dashboard navigation can leave old queued or in-flight work targeting an unmounted viewer -> Clear scheduler state on dashboard id changes and unmount, and ignore or cancel stale in-flight requests.
- Deferred widgets can be confused with no-data widgets -> Use a distinct deferred/loading state and only show no-data after a successful query reports no data.
- A fixed concurrency limit can underuse very fast backends or overrun slow ones -> Use a conservative default of 4 in the viewer and keep backend limits authoritative.
- Multiple active OpenSpec changes touch custom dashboard viewer/editor behavior -> During implementation, merge this refresh scheduling with the selected data group request-signature behavior from `align-custom-dashboard-resource-filters`.

## Migration Plan

1. Add viewer/editor-side widget visibility tracking and a preview scheduling helper without changing backend APIs.
2. Replace eager all-widget `Promise.all` refresh dispatch in the viewer/editor with scheduler-based dispatch.
3. Add widget query state tracking for deferred, queued, loading, stale, success, and error outcomes.
4. Update refresh invalidation for initial load, time range changes, data group changes, variable changes, manual refresh, fullscreen, and per-widget retry.
5. Update focused frontend behavior tests for scheduling, invalidation, stale response ignoring, duplicate-request prevention, fallback behavior, concurrency limit behavior, lifecycle cleanup, and retry.
6. Manually verify a dashboard with many widgets by opening it, scrolling, changing variables, changing time range, changing data group, manually refreshing, retrying a failed widget, navigating away, and opening fullscreen widgets.

Rollback is frontend-only: restore eager viewer dispatch while leaving backend enforcement unchanged.

## Open Questions

None.
