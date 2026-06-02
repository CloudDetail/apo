## 1. Contract And Migration Foundation

- [x] 1.1 Audit current dashboard persistence, API DTOs, and frontend normalization for legacy `timeRange`, `refreshSeconds`, widget size, legend, threshold, empty text, error text, and literal variable fields
- [x] 1.2 Add dashboard definition normalization that maps legacy widget layouts into fixed two-column cells and ignores removed runtime/configuration fields
- [x] 1.3 Update backend dashboard create/update/detail/list contracts to stop requiring removed time, refresh, size, and unsupported display fields while tolerating old saved data
- [x] 1.4 Update frontend dashboard normalization and save payload construction to emit the simplified schema only

## 2. Dashboard Management And Menu Behavior

- [x] 2.1 Remove recently visited and updated time columns from the custom dashboard list UI and list behavior tests
- [x] 2.2 Add full-description display for truncated dashboard descriptions using tooltip, popover, or equivalent accessible interaction
- [x] 2.3 Implement dashboard menu item creation when a dashboard is created from a template, created blank, or copied
- [x] 2.4 Synchronize dashboard menu item label and visibility when dashboard name or visibility changes
- [x] 2.5 Remove or invalidate dashboard menu items when dashboards are deleted
- [x] 2.6 Integrate custom dashboard menu items with the existing menu, route, and role permission tables/APIs used by the current permission control system
- [x] 2.7 Add backend and frontend tests proving dashboard menu items respect menu permission, route permission, dashboard visibility, and do not expand metric data permissions

## 3. Global Time And Query Payloads

- [x] 3.1 Replace dashboard-local viewer time state with the APO global time selector for all dashboard widget queries
- [x] 3.2 Replace editor preview time state with the APO global time selector for query preview payloads
- [x] 3.3 Remove dashboard refresh interval controls from viewer, editor configuration, i18n text, and persisted save payloads
- [x] 3.4 Remove time range from dashboard variable configuration and from any query variable substitution path
- [x] 3.5 Add tests that preview, viewer refresh, and fullscreen widget requests use global time consistently

## 4. Editor Widget Palette And Lifecycle

- [x] 4.1 Replace the toolbar add-widget button with a right-side widget palette shown when no widget is selected
- [x] 4.2 Add palette entries for each supported widget type with localized widget names and mocked visual previews
- [x] 4.3 Implement selecting a palette entry to create a real widget in placement mode
- [x] 4.4 Implement valid drop placement for newly created widgets and preserve selection for subsequent configuration
- [x] 4.5 Fix widget deletion so added or existing widgets can be removed and their cells become reusable
- [x] 4.6 Keep widget copy behavior only with supported configuration fields and fixed-cell placement
- [x] 4.7 Add behavior tests for palette add, placement, delete, and copy flows

## 5. Fixed Two-Column Layout And Drag Feedback

- [x] 5.1 Replace freeform grid width/height layout editing with fixed two-column cell coordinates
- [x] 5.2 Remove resize handles and any width/height controls from the dashboard grid and right-side configuration panel
- [x] 5.3 Implement drag-to-snap behavior that places widgets in the nearest valid empty fixed cell
- [x] 5.4 Prevent invalid drops on occupied cells and retain the previous valid layout when placement fails
- [x] 5.5 Adjust drag styling so only the dragged widget or placeholder moves while toolbar, panel, and non-dragged widget headers remain stable
- [x] 5.6 Add responsive and interaction tests for two-column layout, cell snapping, invalid placement, and drag feedback

## 6. Widget Configuration Simplification

- [x] 6.1 Remove literal widget variable fields such as `$service` from widget configuration UI and save payloads
- [x] 6.2 Remove legend, threshold, empty-state text, and error-state text controls unless each retained field is wired to a visible viewer effect
- [x] 6.3 Update widget renderer to ignore unsupported legacy display fields and show standard empty/error states
- [x] 6.4 Update i18n resources and frontend tests for the simplified widget configuration panel

## 7. Label-Based Variable Configuration

- [x] 7.1 Add or update backend support for discovering label metadata from dashboard widget metric queries
- [x] 7.2 Redesign editor variable configuration to select labels from dashboard metric queries instead of fixed system variable names
- [x] 7.3 Implement variable alias binding so labels with the same meaning, such as `svc` and `svc_name`, can resolve through one dashboard variable
- [x] 7.4 Represent viewer variable selections as multi-select values and update frontend state, URL/state persistence if present, and query payload generation
- [x] 7.5 Resolve dashboard variables into metric-label matchers per widget query and per label alias before backend query execution

## 8. Cascading Candidate Filtering

- [x] 8.1 Update the variable candidate API to accept dashboard ID, variable ID or label binding, current selected variable values, and global time window when needed for label discovery
- [x] 8.2 Filter candidate values by current user's data group permissions and by labels used in the current dashboard's widget metric queries
- [x] 8.3 Implement cascading filtering so downstream candidates, such as service values, are constrained by upstream selections, such as namespace values
- [x] 8.4 Add empty/no-permission/limit states for variable candidates with no allowed values or expensive label queries
- [x] 8.5 Add backend tests for permission filtering, dashboard-label scoping, alias resolution, multi-select resolution, and namespace-to-service cascading

## 9. Preview And Navigation Flow

- [x] 9.1 Change preview mode so it does not replace or discard the editor state
- [x] 9.2 Add a clear path from preview back to editing with unsaved changes preserved
- [x] 9.3 Ensure editor leave prompts still trigger for real navigation away but not for reversible preview toggling
- [x] 9.4 Add frontend tests for preview-enter, preview-return, and unsaved-state preservation

## 10. Verification

- [x] 10.1 Run custom dashboard frontend behavior/unit tests and update snapshots only where behavior intentionally changed
- [x] 10.2 Run backend custom dashboard service and DAO tests covering menu lifecycle, permission-control integration, schema normalization, query payloads, and variable candidates
- [x] 10.3 Run OpenSpec validation/status for `optimize-custom-dashboard` and resolve any spec formatting or dependency issues
- [ ] 10.4 Manually verify the optimized dashboard flow: create dashboard, menu entry appears, add/delete widget, drag/snap, configure variables, preview/return, view with global time, and cascade-filter variables
