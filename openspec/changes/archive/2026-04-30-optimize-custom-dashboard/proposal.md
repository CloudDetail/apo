## Why

The current custom dashboard implementation shipped the core MVP, but real usage has exposed interaction, data-filtering, and information architecture issues that make dashboards hard to operate consistently. This change tightens the product contract around global time, fixed two-column layout, widget insertion/removal, menu behavior, and label-based variables so the feature behaves like a stable operational dashboard rather than a prototype editor.

## What Changes

- Use the APO global time selector as the single source of truth for dashboard viewing, editing preview, and metric query windows.
- Remove dashboard-level refresh interval configuration and widget-level time variables; time-related behavior is driven by the global time selector and manual refresh.
- Remove widget resize behavior and widget size editing from both the grid canvas and the right-side configuration panel.
- Replace the "add widget" button flow with a right-side widget palette shown when no widget is selected; each palette item shows the widget name and a mocked visual preview, and selecting it opens a real draggable widget for placement.
- Allow added widgets to be removed from dashboards and keep copy behavior only if it remains useful after the simplified placement model.
- Change dashboard layout to a fixed two-column grid where widgets snap into fixed cells when dragged.
- Improve drag feedback so only the intended widget/card is visually dragged and surrounding toolbar/header UI does not appear to move with it.
- Remove the "recently visited" and "updated time" columns from "My Dashboards"; make long descriptions fully readable through hover tooltip, popover, or equivalent detail display.
- Ensure newly created dashboards become independent menu items within the existing menu, route, and role permission control system so users can open them directly from navigation only when their permissions allow it.
- Redesign dashboard variable configuration around labels discoverable from the dashboard widgets' metric queries, including aliasing semantically equivalent labels across metrics such as `svc` and `svc_name`.
- Support viewer-side variable filtering against labels present in all dashboard metrics, including multi-select values and cascading candidates such as namespace-filtered service options.
- Keep preview mode reversible so clicking preview does not permanently leave or replace the editing experience.
- Remove widget configuration fields whose behavior is not visibly implemented or product-justified, including literal variable inputs such as `$service` and any legend, threshold, empty text, or error text fields that are not retained with clear runtime effects.

## Capabilities

### New Capabilities

- None.

### Modified Capabilities

- `custom-dashboard-management`: Change dashboard list metadata, long-description display, and newly created dashboard menu registration behavior within the existing permission control system.
- `custom-dashboard-editor`: Change widget add/remove interaction, remove widget resizing and size configuration, introduce fixed two-column snapping, improve drag feedback, and keep preview reversible.
- `custom-dashboard-querying`: Require preview and widget metric queries to use the global time selector and remove component-local time variable configuration.
- `custom-dashboard-variables`: Redesign variable configuration and viewer filtering around metric labels, label aliases, cascading candidates, and multi-select values.
- `custom-dashboard-viewer`: Remove dashboard refresh interval behavior, use global time for all viewer queries, and ensure retained widget display options have visible runtime effects.

## Impact

- Frontend custom dashboard pages: list, viewer, editor, widget renderer, widget palette, layout helpers, variable controls, and i18n text.
- Frontend navigation/menu data: created dashboards must be represented as independent menu items governed by the existing menu, route, and role permission control system.
- Backend custom dashboard APIs and persistence: dashboard metadata/list response shape, menu registration/update/delete handling, variable definitions, label alias mappings, widget layout schema, and removed display/time configuration fields.
- Metric query/preview APIs: request payloads must derive time windows from the global selector and variable values from label-based dashboard variable state.
- Tests: frontend behavior tests for list columns, description detail display, widget add/remove/drag/snap/preview flows, and backend/service tests for menu item lifecycle, variable candidate filtering, label aliasing, and query payload resolution.
