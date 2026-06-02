## Context

Custom dashboards already have list, viewer, editor, query preview, variable, permission, and template implementations. The current code still carries MVP assumptions that conflict with actual operating expectations: dashboard-local time and refresh controls duplicate the global time selector, widget resizing creates unstable layouts, variables are configured as fixed names instead of metric-label driven filters, and the editor uses an add-button/config-panel flow that does not show useful widget previews.

This change is cross-cutting because the dashboard definition, menu integration, frontend editor/viewer behavior, metric query payloads, and variable candidate APIs all need to agree on the same simplified model.

## Goals / Non-Goals

**Goals:**

- Make global time the only time source used by custom dashboard query execution and preview.
- Simplify layout to a fixed two-column, fixed-cell model with drag-to-snap placement and no resize controls.
- Replace button-driven widget creation with a palette-driven widget selection and placement flow.
- Make variable configuration derive from labels used by the dashboard's widget metric queries, with support for label aliases, multi-select values, and cascading candidate filtering.
- Ensure dashboard list and menu behavior match product usage: concise list columns, full description visibility, and newly created dashboards appearing as navigation items governed by the current permission control system.
- Remove configuration fields that do not have a visible or defensible runtime effect.

**Non-Goals:**

- Adding new widget types beyond the existing MVP set.
- Supporting logs, traces, alert events, SQL, mixed data sources, or Grafana iframe behavior.
- Building a general dashboard marketplace or online template administration system.
- Introducing a user-defined refresh interval replacement; manual refresh and global time changes are sufficient for this change.

## Decisions

### Global time is owned outside dashboard definitions

Dashboard records SHALL no longer own refresh interval or default time range behavior for runtime querying. Viewer and editor preview requests will read the current APO global time selector and send the resolved query window to backend query APIs.

Alternative considered: keep dashboard default time as a fallback. This keeps two competing sources of truth and is the source of the current inconsistency, so it is removed from runtime behavior. Existing persisted time fields can be ignored during migration and removed after compatibility is no longer needed.

### Layout uses fixed two-column cells

Widgets will store deterministic cell coordinates such as column and row, not arbitrary width/height editing. Each widget occupies one fixed dashboard cell unless a later spec introduces explicitly named sizes. Dragging snaps to the nearest valid empty cell and does not resize neighboring widgets.

Alternative considered: keep react-grid-style width/height with resize disabled. That reduces UI controls but leaves hidden size state that can create inconsistent old dashboards. Normalizing to fixed cells gives simpler persistence, preview, tests, and future menu rendering.

### Widget insertion uses a side palette

When no widget is selected, the right panel becomes a widget palette. Palette items show the widget name and a mock rendering for visual recognition. Selecting an item creates a real widget instance in placement mode so the creator can drag it into the dashboard grid before configuring details.

Alternative considered: keep a toolbar add button with a dropdown. This is less discoverable and does not provide the requested visual previews.

### Variables are label-based filters, not hard-coded names

Dashboard variables will be configured from labels discovered or declared by widget metric queries. The backend will expose label metadata and candidate values for labels present in the dashboard's metrics. A variable definition can bind multiple label aliases to one semantic filter, for example mapping `svc` and `svc_name` into a single service variable.

Alternative considered: keep fixed variables such as namespace, service, and pod. Fixed names are too brittle because different metrics use different label names for the same concept.

### Cascading variable values are resolved server-side

Variable candidate requests SHALL include already selected variable values, and the backend will filter candidates against the current user's data permissions and the current dashboard's metric labels. Multi-select values are represented as arrays and resolved into query matchers appropriate for each bound label alias.

Alternative considered: compute cascades in the browser from all possible values. This risks exposing unpermitted label values and does not scale for high-cardinality metrics.

### Menu items use the existing permission control system

Creating a dashboard registers an independent menu item through the existing APO menu, route, and role permission infrastructure. Renaming, visibility changes, copying, and deleting dashboards update or remove corresponding menu entries. The menu entry must not grant data access and must still respect dashboard visibility checks when opened.

Alternative considered: leave dashboards only in the list page. The requested behavior requires direct navigation, and existing menu permission infrastructure can carry this without expanding metric permissions.

## Risks / Trade-offs

- Persisted dashboards may contain old width, height, refresh, timeRange, legend, threshold, empty text, or error text fields -> normalize old records when loaded and ignore unsupported fields on save.
- Label aliasing can be ambiguous when metrics use the same label names for different meanings -> require variables to bind to selected query labels and display the metric/query context during configuration.
- Cascading variable queries can become expensive on high-cardinality labels -> enforce existing backend limits, cache short-lived candidate results when safe, and return understandable limit states.
- Menu permission synchronization can create stale or overexposed items if dashboard update/delete partially fails -> perform dashboard, menu, route, and role-permission persistence in one transaction where supported, otherwise add repair logic on list/menu load.
- Fixed two-column layout reduces flexibility -> this is intentional for operational readability and easier implementation; future size variants should be explicit spec additions rather than freeform resizing.

## Migration Plan

1. Add compatibility normalization for existing dashboard definitions: map legacy widget layouts into two-column cells, drop runtime use of dashboard refresh/time fields, and ignore removed widget display fields.
2. Add backend support for dashboard menu item lifecycle and label-based variable candidate resolution while keeping old request fields tolerated during rollout.
3. Update frontend list, editor, viewer, and query preview flows to use the new behavior.
4. Backfill or lazily create menu items for existing custom dashboards the first time dashboards are listed or a migration runs.
5. After rollout, save dashboards in the simplified schema and stop emitting removed fields from APIs.

## Open Questions

- Should fixed cells all have the same height, or should height be derived from widget type while still removing user-controlled resizing?
- Should copied dashboards create a separate menu item immediately, or wait until the copied dashboard is first saved by the new owner?
