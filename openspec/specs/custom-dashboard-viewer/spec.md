## Purpose
This specification defines the custom dashboard capability behavior.
## Requirements
### Requirement: Native Dashboard Viewer
The system SHALL provide an APO-native dashboard viewing page and MUST NOT use Grafana iframe embedding for custom dashboards.

#### Scenario: User opens dashboard
- **WHEN** a user opens a custom dashboard they can view
- **THEN** the system renders the dashboard in the APO-native viewer

### Requirement: Viewer Controls
The system SHALL support variable selectors, the APO global time range, manual refresh, theme integration, and widget fullscreen viewing on the dashboard viewer. The system MUST NOT provide dashboard-level refresh interval configuration.

#### Scenario: Viewer changes time range
- **WHEN** a viewer changes the APO global time range
- **THEN** the system refreshes dashboard widgets using the selected time range

#### Scenario: Viewer manually refreshes
- **WHEN** a viewer clicks manual refresh
- **THEN** the system refreshes dashboard widgets using the current global time range and current dashboard variable selections

#### Scenario: Viewer opens fullscreen widget
- **WHEN** a viewer opens a widget in fullscreen mode
- **THEN** the system displays that widget with the current variables and global time range

#### Scenario: Viewer looks for refresh interval
- **WHEN** a viewer opens the dashboard viewer
- **THEN** the system does not show dashboard-level refresh interval controls

### Requirement: Theme Support
The system SHALL support light and dark themes for dashboard viewing and editing, linked to the APO system theme.

#### Scenario: System theme changes
- **WHEN** the APO system theme changes
- **THEN** the dashboard viewer and editor update their visual theme consistently

### Requirement: Viewer States
The system SHALL distinguish no permission, no data, query failure, and dashboard-not-found states, and retained widget display options MUST have visible runtime effects.

#### Scenario: Query fails
- **WHEN** a dashboard widget query fails
- **THEN** the widget displays the error type, reason, and suggested handling when available

#### Scenario: No data
- **WHEN** a widget query succeeds but returns no data
- **THEN** the widget indicates whether the likely cause is current time range, current variable selection, or no permitted data when the backend can distinguish it

#### Scenario: Unsupported display text was configured previously
- **WHEN** a dashboard contains legacy empty-state or error-state text that is no longer supported
- **THEN** the viewer ignores the legacy field and displays the standard state content

### Requirement: Readable Operational Layout
The system SHALL keep dashboard widgets readable at common desktop resolutions and prevent widget title, legend, axis, and table content from incoherently overlapping.

#### Scenario: Dashboard renders on desktop
- **WHEN** a dashboard is displayed at a common desktop resolution
- **THEN** widget content remains readable and controls do not overlap incoherently
