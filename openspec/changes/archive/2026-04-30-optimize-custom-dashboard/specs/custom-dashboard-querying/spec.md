## MODIFIED Requirements

### Requirement: Custom Query Mode
The system SHALL allow custom PromQL/VictoriaMetrics-compatible query expressions with label-based variable references and legend behavior that is visibly reflected in widget rendering when retained.

#### Scenario: User enters custom query
- **WHEN** a user enters a custom metric query expression
- **THEN** the system stores the expression and any supported label-based variable references with the widget

#### Scenario: User configures unsupported literal variable
- **WHEN** a user configures a widget query
- **THEN** the system does not expose standalone literal variable fields such as `$service` outside the dashboard label-variable system

### Requirement: Query Preview
The system SHALL provide query preview using the current global time range and current dashboard variable values.

#### Scenario: Preview succeeds
- **WHEN** a user previews a valid query that returns data
- **THEN** the system displays a chart or table preview matching the widget type

#### Scenario: Preview has no data
- **WHEN** a user previews a valid query that returns no data
- **THEN** the system displays a no-data state rather than a query failure

#### Scenario: Global time changes before preview
- **WHEN** a user changes the APO global time selector and then previews a widget query
- **THEN** the preview request uses the selected global time range as the query window
