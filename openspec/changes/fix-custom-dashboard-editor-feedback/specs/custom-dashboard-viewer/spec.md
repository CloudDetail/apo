## MODIFIED Requirements

### Requirement: Viewer States
The system SHALL distinguish no permission, no data, query failure, unsupported widget type, and dashboard-not-found states, and retained widget display options MUST have visible runtime effects.

#### Scenario: Query fails
- **WHEN** a dashboard widget query fails
- **THEN** the widget displays the error type, reason, and suggested handling when available

#### Scenario: No data
- **WHEN** a widget query succeeds but returns no data
- **THEN** the widget indicates whether the likely cause is current time range, current variable selection, or no permitted data when the backend can distinguish it

#### Scenario: Unsupported display text was configured previously
- **WHEN** a dashboard contains legacy empty-state or error-state text that is no longer supported
- **THEN** the viewer ignores the legacy field and displays the standard state content

#### Scenario: Historical single-value widget is viewed
- **WHEN** a dashboard contains a historical single-value widget type that has been removed
- **THEN** the viewer displays an unsupported-widget state instead of rendering a single-value panel

### Requirement: Readable Operational Layout
The system SHALL keep dashboard widgets readable at common desktop resolutions and prevent widget title, legend, axis, and table content from incoherently overlapping.

#### Scenario: Dashboard renders on desktop
- **WHEN** a dashboard is displayed at a common desktop resolution
- **THEN** widget content remains readable and controls do not overlap incoherently

#### Scenario: Dashboard contains compact fixed cells
- **WHEN** a dashboard is displayed with fixed two-column cells using the compact 6:4 visual ratio
- **THEN** chart axes, tooltips, table cells, and titles remain readable within each widget

## ADDED Requirements

### Requirement: Widget Unit Display
The system SHALL apply widget unit configuration to numeric chart and table display.

#### Scenario: Chart widget has unit
- **WHEN** a line, area, or bar widget has a configured unit
- **THEN** the viewer displays the unit in chart y-axis labels and tooltip values

#### Scenario: Table widget has unit
- **WHEN** a table widget has a configured unit
- **THEN** the viewer displays the unit with numeric values in the table value column

#### Scenario: Widget value is non-numeric
- **WHEN** a widget result value is not numeric
- **THEN** the viewer leaves the value unmodified instead of appending a unit

### Requirement: Removed Widget Type Compatibility
The system SHALL preserve predictable viewing behavior for dashboards that contain widget types removed from new configuration.

#### Scenario: Historical TopN widget is viewed
- **WHEN** a dashboard contains a historical TopN widget
- **THEN** the viewer renders the widget through table-compatible display

#### Scenario: New dashboard is viewed
- **WHEN** a dashboard contains only supported widget types
- **THEN** the viewer renders line, area, bar, and table widgets without exposing single-value or TopN widget controls
