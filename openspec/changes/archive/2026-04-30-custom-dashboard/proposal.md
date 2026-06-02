## Why

APO currently relies on embedded Grafana dashboards for observability views, which creates a fragmented product experience and limits dashboard customization, permission control, and customer delivery polish. This change introduces a native APO custom metric dashboard capability so technical users and advanced customer users can build, edit, and view operational dashboards without depending on Grafana iframe pages.

## What Changes

- Add a first-level "My Dashboards" entry for listing dashboards the current user can view.
- Support dashboard lifecycle operations: create from template, create blank dashboard, view, edit, copy, delete, and save changes immediately.
- Replace the custom dashboard viewing experience with APO-native pages; Grafana iframe embedding is not used for this new capability.
- Add a grid-based dashboard editor with draggable/resizable widgets, widget add/copy/delete actions, and a right-side configuration panel.
- Support MVP widget types: line chart, area chart, bar chart, single value, table, and TopN ranking.
- Support metric queries through preset queries and custom PromQL/VictoriaMetrics-compatible query expressions.
- Add query preview, syntax validation, error messaging, timeout and result-size controls, and query audit requirements.
- Add dashboard variables, including data group, cluster, namespace, service, instance/Pod, and time range.
- Enforce three permission layers: create permission, dashboard visibility permission, and data group based metric data permission.
- Add built-in dashboard templates and preset queries that ship with the platform.
- Support light and dark themes, global time range, auto refresh, manual refresh, and unsaved-change prompts.
- Explicitly keep existing system dashboards in place and exclude logs, traces, alert events, SQL, mixed data sources, draft/publish workflow, version history, online preset-template administration, and command-center style big-screen editing from MVP scope.

## Capabilities

### New Capabilities

- `custom-dashboard-management`: Dashboard navigation, listing, lifecycle, visibility, and ownership rules.
- `custom-dashboard-editor`: Native grid editor, widgets, layout editing, configuration panels, and save behavior.
- `custom-dashboard-querying`: Metric query modes, preset queries, custom PromQL/VictoriaMetrics queries, preview, validation, and backend execution limits.
- `custom-dashboard-variables`: Dashboard variables, viewer-side variable selection, variable substitution, and permission-filtered candidate values.
- `custom-dashboard-permissions`: Create permission, dashboard visibility permission, and data permission enforcement.
- `custom-dashboard-templates`: Built-in templates and preset dashboard structures for quick dashboard creation.
- `custom-dashboard-viewer`: Native dashboard viewing page, refresh/time/theme behavior, fullscreen widget viewing, and empty/error states.

### Modified Capabilities

- None.

## Impact

- Frontend: new dashboard list, create flow, native viewer, grid editor, widget rendering, configuration panel, variable controls, theme integration, and unsaved-change handling.
- Backend APIs: dashboard CRUD, copy, templates, preset queries, variable candidate values, metric query execution, syntax validation, and query preview endpoints.
- Data model: dashboard metadata, visibility, layout, variables, widgets, query definitions, template definitions, and audit records.
- Permission system: new create-dashboard role permission and enforcement of ownership, visibility, and data group constraints.
- Observability data access: server-side PromQL/VictoriaMetrics-compatible query execution with data group scoping, timeout, maximum returned points, concurrency, and resource limits.
- Internationalization: all user-facing dashboard, editor, query, and error text must support Chinese and English.
