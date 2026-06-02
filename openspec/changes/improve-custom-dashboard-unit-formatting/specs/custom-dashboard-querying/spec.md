## ADDED Requirements

### Requirement: Query Preview Display Formatting
The system SHALL apply widget display unit formatting to query preview results after metric query execution. The system MUST NOT change the PromQL/VictoriaMetrics query expression, query request values, or raw metric values to implement display unit formatting.

#### Scenario: Preview formats returned numeric values
- **WHEN** a query preview returns numeric metric values for a widget with preset unit display configuration
- **THEN** the preview renders those values using the same formatter used by the dashboard viewer

#### Scenario: Preview preserves query semantics
- **WHEN** a widget uses preset or custom unit display configuration
- **THEN** the query request sent to the backend remains based on the widget query and dashboard variables rather than formatted display values
