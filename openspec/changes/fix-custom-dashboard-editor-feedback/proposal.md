## Why

Manual testing of the optimized custom dashboard flow found editor regressions and product gaps that make the APO dashboard experience fall short of a practical Grafana dashboard replacement. The editor needs clearer widget selection, stable two-column placement, simpler dashboard variables, and consistent numeric unit display before the feature is reliable for daily dashboard building.

## What Changes

- Fix widget palette previews so each supported widget type has a distinct thumbnail that matches its runtime rendering shape.
- Add deletion for dashboard variables in the editor.
- Change new widget cells to the previous compact visual ratio, approximately 6:4, instead of oversized cards.
- Fix grid dragging so widgets can move across a fixed two-column, multi-row layout and the layout auto-compacts to remove empty cells between widgets.
- Simplify variable configuration to a Grafana-inspired dashboard-level filter model: the system suggests semantic variables from dashboard metric queries and users only choose which variables to enable.
- Fix drag feedback so only the dragged widget or placement indicator moves; toolbar, side panel, and unrelated header UI remain stable.
- Remove the single-value metric widget type and TopN widget type from new dashboard configuration.
- Keep widget unit configuration and apply it to chart axes, chart tooltips, and table value display.
- Preserve compatibility for existing dashboards by rendering historical TopN widgets as tables and marking historical single-value widgets as removed/unsupported until the user deletes or changes them.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `custom-dashboard-editor`: Widget palette previews, supported widget types, variable editing controls, fixed-cell sizing, drag placement, drag feedback, and unsupported historical widget behavior change.
- `custom-dashboard-variables`: Variable setup changes from manual name/label binding to system-suggested semantic dashboard variables selected by the user.
- `custom-dashboard-querying`: Variable resolution and numeric unit formatting affect query requests and query result presentation.
- `custom-dashboard-viewer`: Viewer rendering changes for removed historical widget types, TopN table compatibility, and unit display in charts and tables.

## Impact

- Frontend custom dashboard editor, viewer, renderer, behavior helpers, styles, i18n, and behavior tests.
- Backend dashboard normalization and custom dashboard query handling if persisted widget types or variable definitions require server-side compatibility.
- Existing dashboard definitions that contain `single` or `topn` widget types.
- No new external dependencies are expected.
