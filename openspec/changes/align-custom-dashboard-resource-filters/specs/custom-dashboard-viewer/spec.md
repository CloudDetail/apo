## MODIFIED Requirements

### Requirement: Viewer Controls
The system SHALL support dashboard-local variable selectors, the APO global time range, the APO top-level data group selector, manual refresh, theme integration, and widget fullscreen viewing on the dashboard viewer. The system MUST NOT provide dashboard-level refresh interval configuration or a dashboard-local data group selector.

#### Scenario: Viewer changes time range
- **WHEN** a viewer changes the APO global time range
- **THEN** the system refreshes dashboard widgets using the selected time range

#### Scenario: Viewer changes top-level data group
- **WHEN** a viewer changes the APO top-level data group selector
- **THEN** the system refreshes dashboard variable candidates and widgets using the selected data group

#### Scenario: Viewer manually refreshes
- **WHEN** a viewer clicks manual refresh
- **THEN** the system refreshes dashboard widgets using the current global time range, current top-level data group, and current dashboard-local variable selections

#### Scenario: Viewer opens fullscreen widget
- **WHEN** a viewer opens a widget in fullscreen mode
- **THEN** the system displays that widget with the current dashboard-local variables, current top-level data group, and global time range

#### Scenario: Viewer looks for refresh interval
- **WHEN** a viewer opens the dashboard viewer
- **THEN** the system does not show dashboard-level refresh interval controls

#### Scenario: Viewer looks for dashboard data group variable
- **WHEN** a viewer opens the dashboard viewer
- **THEN** the system does not show data group in the dashboard variable selector area
