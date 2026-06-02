## 1. Widget Type And Compatibility Foundation

- [x] 1.1 Audit frontend and backend references to `single`, `topn`, widget unit display, variable definitions, and grid layout normalization.
- [x] 1.2 Update dashboard normalization so historical `topn` widgets use table-compatible rendering and historical `single` widgets are preserved as removed/unsupported.
- [x] 1.3 Remove `single` and `topn` from new widget creation, chart-type selection, preset compatibility filtering, i18n display options, and save payload generation.
- [x] 1.4 Add tests for supported widget type lists and historical `single`/`topn` compatibility behavior.

## 2. Widget Palette And Unit Display

- [x] 2.1 Replace generic palette preview markup/styles with distinct thumbnails for line, area, bar, and table widgets.
- [x] 2.2 Apply widget `display.unit` and `display.decimals` to chart y-axis labels and tooltip values.
- [x] 2.3 Apply widget `display.unit` and `display.decimals` to numeric table value cells while leaving non-numeric values unchanged.
- [x] 2.4 Add behavior or component tests covering palette thumbnails and unit formatting for chart and table widgets.

## 3. Fixed Grid Drag And Compaction

- [x] 3.1 Update fixed-cell sizing styles so editor and viewer cells use an approximate 6:4 visual ratio.
- [x] 3.2 Update grid pointer-to-cell calculation to support two columns and unlimited rows.
- [x] 3.3 Implement deterministic row-major layout compaction after add, copy, delete, and drag placement.
- [x] 3.4 Fix drag feedback so only the dragged widget or placeholder moves and toolbar, panel, and unrelated widget header UI remain stable.
- [x] 3.5 Add interaction tests for lower-row dragging, occupied-cell handling, compaction, delete compaction, and drag feedback state.

## 4. Semantic Dashboard Variables

- [x] 4.1 Define centralized semantic variable metadata and label alias mappings for dataGroup, cluster, namespace, service, and pod.
- [x] 4.2 Replace manual variable name/label editing controls with system-suggested semantic variables derived from dashboard widget and preset query labels.
- [x] 4.3 Add variable deletion or disable behavior that removes the semantic variable from editor state and save payloads.
- [x] 4.4 Update variable candidate and query payload generation so enabled semantic variables apply only to widget queries with matching labels.
- [x] 4.5 Add tests for variable suggestion, enable/delete behavior, alias matching, multi-select matching, and no-injection behavior for non-matching queries.

## 5. Query Preview And Viewer Integration

- [x] 5.1 Update editor preview requests to use enabled semantic variables and widget display formatting.
- [x] 5.2 Update viewer query refresh behavior so semantic variable changes refresh only affected widgets.
- [x] 5.3 Render unsupported historical `single` widgets with a clear unsupported state in viewer and editor.
- [x] 5.4 Render historical `topn` widgets through table-compatible display in viewer and editor.
- [x] 5.5 Add tests for preview variable application, viewer variable refresh, unsupported single-value rendering, and TopN table compatibility.

## 6. Verification

- [x] 6.1 Run custom dashboard frontend behavior/unit tests.
- [x] 6.2 Run backend custom dashboard service/query tests affected by widget normalization or variable payload changes.
- [x] 6.3 Run OpenSpec validation/status for `fix-custom-dashboard-editor-feedback` and resolve any spec or task formatting issues.
- [x] 6.4 Manually verify editor flow: add widgets, inspect distinct thumbnails, drag across multiple rows, delete widgets, confirm compact layout, enable/delete variables, preview, save, and view with unit display.
