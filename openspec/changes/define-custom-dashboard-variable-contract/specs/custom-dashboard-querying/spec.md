## MODIFIED Requirements

### Requirement: Custom Query Mode
The system SHALL allow custom PromQL/VictoriaMetrics-compatible query expressions with label-based variable references and legend behavior that is visibly reflected in widget rendering when retained. Variable substitution MUST preserve Include All semantics for both injected label matchers and explicit variable tokens.

#### Scenario: All is selected for injected matcher
- **WHEN** a built-in resource variable is selected as All and the system injects a matcher for a matching label
- **THEN** the injected matcher does not narrow the query beyond a regex wildcard for that label

#### Scenario: All is selected in exact matcher variable token
- **WHEN** a query contains an exact matcher such as `namespace="$namespace"` and the `namespace` variable is selected as All
- **THEN** backend query execution treats that matcher as no effective namespace restriction rather than as a literal `namespace=".*"` comparison

#### Scenario: All is selected in exact matcher Grafana formatted token
- **WHEN** a query contains an exact matcher such as `pod="${pod:raw}"` or `namespace="${namespace:regex}"` and the variable is selected as All
- **THEN** backend query execution treats that matcher as a wildcard regex matcher rather than as a literal exact comparison

#### Scenario: All is selected in regex matcher variable token
- **WHEN** a query contains a regex matcher such as `namespace=~"$namespace"` and the `namespace` variable is selected as All
- **THEN** backend query execution substitutes the configured All regex value

#### Scenario: Concrete values are selected in variable token
- **WHEN** a query contains a variable token and one or more concrete values are selected
- **THEN** backend query execution substitutes escaped values appropriate to the matcher format
