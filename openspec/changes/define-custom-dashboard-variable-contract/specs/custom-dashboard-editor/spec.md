## ADDED Requirements

### Requirement: Dashboard Variable Add List
The system SHALL show a grouped add-variable list in the dashboard information modal. The list SHALL include unenabled built-in resource variables independent of widget query content and SHALL supplement them with query-discovered custom labels.

#### Scenario: Creator opens add-variable list on a blank dashboard
- **WHEN** the dashboard has no widget query labels
- **THEN** the add-variable list shows all unenabled built-in resource variables

#### Scenario: Creator opens add-variable list with mixed labels
- **WHEN** widget queries contain both built-in resource aliases and custom labels
- **THEN** the add-variable list groups unenabled built-in resource variables separately from query-discovered custom labels

#### Scenario: Built-in resource is already enabled
- **WHEN** a built-in resource variable is already enabled, including through an imported label-shaped definition
- **THEN** the add-variable list does not show that built-in resource again

#### Scenario: Custom label is already enabled by alias
- **WHEN** an enabled variable name or label alias matches a query-discovered custom label
- **THEN** the add-variable list does not show that custom label again
