## MODIFIED Requirements

### Requirement: System Variables
The system SHALL support dashboard variables as system-suggested semantic filters derived from labels used by the dashboard's widget metric queries. The system SHALL also preserve imported Grafana-style static option variables, label-value variables, and `includeAll`/`allValue` metadata when those variables are represented in the APO dashboard definition. The system MUST NOT expose time range as a dashboard variable because time is controlled by the APO global time selector.

#### Scenario: Creator configures variables
- **WHEN** a creator configures dashboard variables
- **THEN** the system shows suggested semantic variables discovered from or declared by the dashboard's widget metric queries and allows the creator to enable or disable those variables

#### Scenario: Creator configures equivalent labels
- **WHEN** different dashboard metrics use different labels for the same filter concept
- **THEN** the system maps those labels, such as `svc` and `svc_name`, to one semantic dashboard variable without requiring creator label binding

#### Scenario: Creator looks for time variable
- **WHEN** a creator configures dashboard variables
- **THEN** the system does not offer time range as a configurable dashboard variable

#### Scenario: Creator deletes enabled variable
- **WHEN** a creator deletes or disables an enabled dashboard variable
- **THEN** the system removes that semantic variable from the dashboard definition and from subsequent query filtering

#### Scenario: Dashboard contains static option variable
- **WHEN** a dashboard definition contains a variable with fixed `options`
- **THEN** the viewer displays those options without calling the backend candidate API

#### Scenario: Dashboard contains label-value variable
- **WHEN** a dashboard definition contains a `label` variable with labels and a label-value query
- **THEN** the viewer requests candidates using that variable query and the backend resolves already selected upstream variables before querying Prometheus-compatible label values

#### Scenario: Dashboard contains include-all variable
- **WHEN** a dashboard definition contains a variable with `includeAll`
- **THEN** the viewer prepends an `All` candidate and uses the variable's `allValue`, defaulting to `.*` when no all value is declared

#### Scenario: Dashboard contains datasource variable
- **WHEN** an imported Grafana dashboard declares a datasource variable
- **THEN** APO does not expose it as a dashboard variable because custom dashboards use the backend-configured Prometheus/VictoriaMetrics data source

### Requirement: Variable Substitution
The system SHALL resolve enabled semantic dashboard variables into metric-label matchers for custom query expressions and preset query parameters, including multi-select values, imported Grafana variable tokens, and system-managed label aliases.

#### Scenario: Query can use semantic variable
- **WHEN** a widget query contains a label that matches an enabled semantic variable's alias mapping
- **THEN** the system resolves the selected variable values into a matcher for that query's matching label before backend query execution

#### Scenario: Query cannot use semantic variable
- **WHEN** a widget query does not contain any label matching an enabled semantic variable
- **THEN** the system does not inject that variable into the widget query

#### Scenario: Variable has multiple selected values
- **WHEN** a viewer selects multiple values for a dashboard variable
- **THEN** the system resolves the selection into a query matcher that includes all selected values for each applicable label alias

#### Scenario: Variable maps multiple label aliases
- **WHEN** a dashboard variable maps to multiple metric labels with the same meaning
- **THEN** the system applies the selected variable values to each relevant metric query using that query's matching label name

#### Scenario: Imported query uses Grafana variable syntax
- **WHEN** an imported widget query contains `$name`, `${name}`, `${name:regex}`, `${name:raw}`, or `{{ name }}`
- **THEN** the backend resolves those tokens before Prometheus-compatible query execution, with regex format escaping regular-expression metacharacters except the imported all value

### Requirement: Viewer Variable Controls
The system SHALL display dashboard variable selectors on the viewing page for enabled semantic variables, static option variables, and label-value variables, and refresh affected widgets when variable values change.

#### Scenario: Viewer changes variable value
- **WHEN** a viewer selects a different variable value
- **THEN** the system refreshes widgets affected by that semantic variable using the new value

#### Scenario: Viewer selects multiple values
- **WHEN** a viewer selects multiple values in a dashboard variable selector
- **THEN** the system refreshes affected widgets using all selected values

#### Scenario: Viewer has disabled variable
- **WHEN** a semantic variable is not enabled for the dashboard
- **THEN** the viewer does not display a selector for that variable and does not apply it to widget queries

### Requirement: Permission-Filtered Candidate Values
The system MUST filter variable candidate values by the current user's data group permissions, the dashboard's widget metric labels, and already selected upstream variable values.

#### Scenario: User views variable candidates
- **WHEN** a user opens variable candidate values
- **THEN** the system returns only candidates allowed by the user's data group permissions

#### Scenario: User has no candidates
- **WHEN** a user has no permitted candidate values for a dashboard variable
- **THEN** the system displays a no-available-data or no-permission state

#### Scenario: Namespace filters service candidates
- **WHEN** a viewer selects one or more namespace values and then opens service variable candidates
- **THEN** the system returns only service candidates associated with the selected namespace values and permitted by the user's data permissions

#### Scenario: Candidates are limited to dashboard labels
- **WHEN** a viewer opens candidates for a dashboard variable
- **THEN** the system returns values from labels used by the current dashboard's widget metric queries, not unrelated global labels

### Requirement: Variable Label Discovery
The system SHALL expose semantic variable suggestions and internal label metadata derived from the dashboard's widget metric queries.

#### Scenario: Creator opens variable configuration
- **WHEN** the dashboard contains widget metric queries with labels matching known semantic variable aliases
- **THEN** the system shows semantic variable suggestions for those labels instead of exposing raw label binding controls

#### Scenario: Dashboard has no metric labels
- **WHEN** no widget metric query can provide label metadata
- **THEN** the system explains that variables become available after configuring widget metric queries

#### Scenario: Dashboard has unknown labels
- **WHEN** widget metric queries contain labels that do not match known semantic variable aliases
- **THEN** the system does not expose those labels as manually configurable variables
