## 1. Scheduling State And Helpers

- [x] 1.1 Add behavior helpers for widget preview request eligibility, priority, stale state, last completed request signatures, and per-widget latest request signatures.
- [x] 1.2 Add a dashboard-local preview scheduler that caps concurrent widget preview requests at the default limit of four and keeps only the newest queued request per widget.
- [x] 1.3 Add fallback behavior that treats all widgets as eligible and schedules them through the concurrency-limited queue when viewport observation is unavailable.
- [x] 1.4 Clear queued scheduler work on dashboard id changes or viewer unmount, and ignore or cancel stale in-flight requests.

## 2. Viewer Visibility And Refresh Flow

- [x] 2.1 Add viewer-side widget visibility registration using `IntersectionObserver` with a preload margin.
- [x] 2.2 Replace eager `Promise.all` widget preview refresh dispatch with scheduler-based dispatch.
- [x] 2.3 Preserve stale response ignoring by comparing each widget's current preview request signature before applying results.
- [x] 2.4 Update initial dashboard open so only visible and near-visible widgets query immediately.
- [x] 2.5 Prevent duplicate preview requests when a widget re-enters the viewport with a current successful or no-data preview signature.
- [x] 2.6 Apply the same viewport scheduling and duplicate-request prevention to the custom dashboard editor page.

## 3. Refresh Invalidation

- [x] 3.1 Update global time range and data group changes to mark widgets stale and schedule only eligible widgets immediately.
- [x] 3.2 Update dashboard variable changes to invalidate only affected widgets when possible and schedule only eligible affected widgets.
- [x] 3.3 Update manual refresh to invalidate all widgets while dispatching through the scheduler.
- [x] 3.4 Prioritize fullscreen widget preview loading when the current preview is missing or stale.
- [x] 3.5 Add per-widget retry for failed previews that schedules only the retried widget with the current request context.
- [x] 3.6 Update editor global time range, data group, variable, and manual refresh paths to invalidate widgets while scheduling only eligible widgets immediately.

## 4. Widget States And UI Copy

- [x] 4.1 Track deferred, queued, loading, stale, success, no-data, failure, and retrying states per widget in the viewer.
- [x] 4.2 Update widget rendering so deferred widgets are not shown as successful no-data results.
- [x] 4.3 Preserve existing per-widget query failure, no-permission, no-data, loading, and fullscreen presentation.
- [x] 4.4 Ensure per-widget retry state and action do not refresh unrelated widgets.

## 5. Verification

- [x] 5.1 Add or update frontend behavior tests for the default four-request concurrency cap, queue replacement, priority, stale response ignoring, duplicate-request prevention, and scheduler cleanup.
- [x] 5.2 Add or update frontend behavior tests for initial open, scroll eligibility, variable changes, global time range changes, data group changes, manual refresh, fullscreen priority, viewport-observation fallback, and per-widget retry.
- [x] 5.3 Run `cd frontend && node --test src/pro/views/CustomDashboard/behavior.test.mjs`.
- [ ] 5.4 Manually verify a dashboard with many widgets for initial open, scrolling, duplicate scroll re-entry, variable changes, global time range changes, data group changes, manual refresh, per-widget retry, navigation away from the dashboard, and fullscreen.
- [x] 5.5 Run OpenSpec status or validation for `defer-custom-dashboard-widget-queries` and resolve formatting issues.
