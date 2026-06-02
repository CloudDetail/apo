## Why

Custom dashboard viewing and editing currently can refresh all widgets together when a dashboard opens, when the global time range changes, when variables change, or when the user manually refreshes. Dashboards with many widgets can therefore produce a burst of metric preview requests, increasing first-screen latency and backend query pressure even for widgets the user has not reached yet.

## What Changes

- Defer non-visible custom dashboard widget query requests until the widget is near or inside the current page viewport.
- Keep first-screen widgets eager so the dashboard remains useful immediately after opening.
- Add viewport-aware preloading so widgets begin loading before the user scrolls to them.
- Add dashboard-level request scheduling so lazy loading does not become another burst when many widgets enter view together.
- Keep manual refresh, time range changes, data group changes, and variable changes compatible with deferred widgets by refreshing visible widgets immediately and marking non-visible affected widgets stale until they become visible.
- Preserve per-widget loading, no-data, error, and fullscreen behavior, and support per-widget retry for failed preview requests.

No breaking changes are expected for saved dashboard definitions or backend query APIs.

## Capabilities

### New Capabilities

- `custom-dashboard-widget-query-scheduling`: Viewport-aware widget query scheduling, refresh invalidation, and request concurrency behavior for custom dashboard viewing and editing.

### Modified Capabilities

- `custom-dashboard-viewer`: Viewer and editor refresh behavior changes from refreshing every widget immediately to refreshing visible or near-visible widgets first and deferring offscreen widgets.

## Impact

- Frontend custom dashboard viewer/editor loading flow, widget visibility tracking, preview request dispatch, stale response handling, and per-widget state management.
- Frontend custom dashboard behavior helpers and behavior tests for scheduling, invalidation, and request signatures.
- Widget renderer loading and empty states may need small adjustments so deferred widgets are distinguishable from queried no-data widgets.
- Backend query APIs are not expected to change, but existing backend concurrency, timeout, resource limit, and audit behavior remain the final enforcement layer.
- Manual verification should cover dashboards with many widgets, scrolling and duplicate scroll re-entry, variable changes, global time range changes, data group changes, manual refresh, per-widget retry, dashboard navigation, and fullscreen widgets.
