## MODIFIED Requirements

### Requirement: Viewer Controls
The system SHALL support variable selectors, the APO global time range, manual refresh, theme integration, and widget fullscreen viewing on the dashboard viewer. Widget preview refreshes in the viewer SHALL follow viewport-aware query scheduling so visible or near-visible widgets refresh before offscreen widgets. The system MUST NOT provide dashboard-level refresh interval configuration.

#### Scenario: Viewer changes time range
- **WHEN** a viewer changes the APO global time range
- **THEN** the system marks dashboard widgets stale, refreshes visible or near-visible widgets using the selected time range, and defers offscreen widget refresh until those widgets become eligible

#### Scenario: Viewer manually refreshes
- **WHEN** a viewer clicks manual refresh
- **THEN** the system marks dashboard widgets stale, refreshes visible or near-visible widgets using the current global time range and current dashboard variable selections, and defers offscreen widget refresh until those widgets become eligible

#### Scenario: Viewer opens fullscreen widget
- **WHEN** a viewer opens a widget in fullscreen mode
- **THEN** the system displays that widget with the current variables and global time range and immediately schedules its preview if the current preview is missing or stale

#### Scenario: Viewer looks for refresh interval
- **WHEN** a viewer opens the dashboard viewer
- **THEN** the system does not show dashboard-level refresh interval controls

### Requirement: Viewer States
The system SHALL distinguish deferred, loading, no permission, no data, query failure, retrying, and dashboard-not-found states, and retained widget display options MUST have visible runtime effects.

#### Scenario: Query fails
- **WHEN** a dashboard widget query fails
- **THEN** the widget displays the error type, reason, and suggested handling when available

#### Scenario: Viewer retries query failure
- **WHEN** a viewer retries a failed widget query
- **THEN** that widget enters a retrying or loading state and unrelated widgets keep their current state

#### Scenario: No data
- **WHEN** a widget query succeeds but returns no data
- **THEN** the widget indicates whether the likely cause is current time range, current variable selection, or no permitted data when the backend can distinguish it

#### Scenario: Widget query is deferred
- **WHEN** a widget preview query has not started because the widget is offscreen
- **THEN** the widget displays a deferred or pending state and does not present the widget as a successful no-data result

#### Scenario: Unsupported display text was configured previously
- **WHEN** a dashboard contains legacy empty-state or error-state text that is no longer supported
- **THEN** the viewer ignores the legacy field and displays the standard state content
