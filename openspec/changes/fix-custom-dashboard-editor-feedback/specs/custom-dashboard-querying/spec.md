## MODIFIED Requirements

### Requirement: Preset Query Mode
The system SHALL provide preset queries grouped by scenario, with name, description, compatible chart types, parameters, parameter descriptions, and semantic variable label metadata when the preset can be filtered by dashboard variables.

#### Scenario: User selects preset query
- **WHEN** a user selects a preset query for a compatible widget type
- **THEN** the system displays only the parameters exposed by that preset query

#### Scenario: User attempts to edit preset definition
- **WHEN** a user configures a preset query
- **THEN** the system allows parameter values but does not allow changing the preset query definition

#### Scenario: Preset query supports semantic variable
- **WHEN** a preset query declares or contains labels matching enabled semantic dashboard variables
- **THEN** the query request applies those variable values to the preset query using the matching labels

### Requirement: Custom Query Mode
The system SHALL allow custom PromQL/VictoriaMetrics-compatible query expressions with system-resolved semantic dashboard variable filtering and legend behavior that is visibly reflected in widget rendering when retained.

#### Scenario: User enters custom query
- **WHEN** a user enters a custom metric query expression
- **THEN** the system stores the expression and applies enabled semantic dashboard variables whose label aliases match labels in the expression

#### Scenario: User configures unsupported literal variable
- **WHEN** a user configures a widget query
- **THEN** the system does not expose standalone literal variable fields such as `$service` outside the dashboard semantic-variable system

### Requirement: Query Preview
The system SHALL provide query preview using the current global time range, current dashboard variable values, and widget display formatting.

#### Scenario: Preview succeeds
- **WHEN** a user previews a valid query that returns data
- **THEN** the system displays a chart or table preview matching the widget type and applying configured decimals and unit display

#### Scenario: Preview has no data
- **WHEN** a user previews a valid query that returns no data
- **THEN** the system displays a no-data state rather than a query failure

#### Scenario: Global time changes before preview
- **WHEN** a user changes the APO global time selector and then previews a widget query
- **THEN** the preview request uses the selected global time range as the query window

#### Scenario: Variable changes before preview
- **WHEN** a user changes an enabled dashboard variable and then previews a widget query
- **THEN** the preview request uses the semantic variable values that apply to that widget query
