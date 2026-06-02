## Context

APO already has frontend and backend modules for services, metrics, data groups, role permissions, and internationalized UI text. Current dashboard-like pages can embed or reference Grafana, but the custom dashboard capability must become an APO-native experience that owns navigation, dashboard configuration, metric query execution, permissions, and rendering.

The MVP is a cross-cutting feature: it adds new frontend pages and editor interactions, new backend CRUD/query APIs, new persistent dashboard configuration, and stricter server-side data permission controls. Existing system dashboards remain available; this change adds a separate "My Dashboards" capability for user-created metric dashboards.

## Goals / Non-Goals

**Goals:**

- Provide native dashboard list, create, view, edit, copy, delete, and save flows.
- Store dashboard definitions as structured configuration: metadata, visibility, variables, layout, widgets, query definitions, and style options.
- Execute all metric queries through backend APIs so data group permission, timeout, maximum points, concurrency, and audit controls are enforceable.
- Support preset queries for low-friction creation and custom PromQL/VictoriaMetrics-compatible expressions for advanced users.
- Keep the editor grid-based and operationally dense, with predictable layout constraints and readable chart/table states.
- Make built-in templates and preset queries versioned application assets rather than user-editable runtime records in MVP.

**Non-Goals:**

- Replacing or deleting existing system dashboards.
- Embedding Grafana iframe pages for this custom dashboard capability.
- Supporting logs, traces, alert events, SQL, mixed data sources, draft/publish workflows, version history, rollback, import/export, or command-center big-screen editing in MVP.
- Adding an online administration UI for preset queries or built-in templates.

## Decisions

### Native Dashboard Domain Model

Store dashboards as first-class APO entities with owner, visibility, metadata, variables, layout, and widgets. Widget configuration includes chart type, query mode, query definition, unit, decimals, legend, thresholds, and empty/error display options.

Alternative considered: persist a Grafana-compatible JSON model and render a subset. This was rejected because the product requirement is an APO-native workbench, and a Grafana-shaped model would leak external assumptions into permission, variable, and rendering behavior.

### Backend-Mediated Metric Query Execution

All preview and viewer queries go through backend services. The backend validates syntax where possible, resolves variables, injects data group constraints, enforces timeout/point/concurrency limits, audits query execution, and calls the configured PromQL/VictoriaMetrics-compatible endpoint.

Alternative considered: allow the frontend to call the metrics backend directly. This was rejected because frontend filtering cannot enforce data permissions and cannot reliably audit or cap resource usage.

### Separated Permission Layers

Implement create permission, dashboard visibility, and data permission as separate checks. Create permission controls new/copy operations, visibility controls list/open access, and data group permission controls variables and query results.

Alternative considered: make dashboard visibility imply data access. This was rejected because sharing a dashboard must never expand metric data access.

### Grid-Based Editor

Use a bounded grid layout model with integer x/y/w/h coordinates, minimum widget sizes, collision avoidance, and responsive desktop rendering. The editor should use established frontend grid/chart libraries already compatible with the project stack where practical, rather than a free canvas.

Alternative considered: freeform canvas with layers and absolute positioning. This was rejected because MVP targets an operational workbench, not a big-screen designer.

### Built-In Templates and Preset Queries

Ship template and preset query definitions with the application version. Creating a dashboard from a template copies the template structure into the user dashboard, so future template upgrades do not mutate existing user dashboards.

Alternative considered: store templates as fully user-managed database records from day one. This was deferred because MVP explicitly excludes administrator online maintenance of templates.

### API Shape

Add backend endpoints around the following resources:

- dashboards: list, get, create, update, delete, copy
- templates: list, get
- preset queries: list, get
- variables: candidate values for the current user and current dashboard context
- metric query: validate and preview/execute

The frontend should use these APIs from dedicated dashboard modules and keep dashboard editor state local until save. Save success replaces the canonical dashboard configuration immediately.

## Risks / Trade-offs

- Query fan-out or large result sets can overload the metrics backend -> enforce per-dashboard widget limits, per-widget query limits, timeouts, maximum returned points, and concurrency caps.
- Custom PromQL can accidentally bypass expected label filters -> inject server-side data group constraints after parsing/resolving variables and reject expressions that cannot be safely constrained.
- Different viewers can see different variable candidates and data for the same dashboard -> make no-data/no-permission states explicit and test with multiple data groups.
- Template changes after release can drift from copied user dashboards -> treat template-created dashboards as independent copies and document that template upgrades do not rewrite user dashboards.
- Grid editing can become awkward on small screens -> optimize MVP for common desktop resolutions and provide readable responsive viewing; complex mobile editing is out of MVP.
- Save failure could lose editor work -> keep unsaved state in the frontend after failed saves and show retryable error messages.

## Migration Plan

1. Add database tables or migrations for dashboard metadata, dashboard configuration, and query audit records.
2. Add role permission seed data for dashboard creation and update API permission metadata.
3. Add backend dashboard CRUD, templates, preset queries, variables, validation, and query execution services.
4. Add frontend routes, menu entry, list page, create flow, viewer, editor, widget renderers, and i18n text.
5. Roll out with existing system dashboards unchanged; the new "My Dashboards" entry is additive.
6. Rollback by hiding the new menu/route and leaving stored dashboard records unused; existing dashboards are unaffected.

## Open Questions

- Exact metric labels and label-injection strategy for data group, cluster, namespace, service, and instance/Pod scopes.
- Final maximums for dashboards per user, widgets per dashboard, queries per widget, returned points, and concurrent query execution.
- Whether instant query support is required in MVP or can be limited to time series query preview/execution.
- Final left-navigation behavior when a user has many dashboards.
