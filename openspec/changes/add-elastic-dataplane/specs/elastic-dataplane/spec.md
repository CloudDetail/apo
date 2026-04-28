## ADDED Requirements

### Requirement: Elastic dataplane type SHALL be available in the dataplane catalog
The system SHALL expose `elastic` as a formal dataplane type in `/dataplane`, with `Elastic` as the display name and `elastic` as the canonical persisted type name.

#### Scenario: Elastic type is listed for configuration
- **WHEN** a user opens `/dataplane`
- **THEN** the dataplane type list SHALL include `Elastic`
- **THEN** selecting the type SHALL open the Elastic-specific configuration flow instead of the legacy integration metadata path

### Requirement: Elastic dataplane SHALL support Log-only onboarding
The system SHALL support configuring a single Elastic dataplane instance against either Elastic Cloud or a self-managed Elastic Stack for Log capability.

#### Scenario: User configures a log-only Elastic instance
- **WHEN** a user configures an Elastic instance with only log capability enabled
- **THEN** the system SHALL allow the instance to be tested and saved without requiring Trace, Topology, ServiceInstance, RedMetric, or Metrics inputs

#### Scenario: User configures Elastic from the APO main flow
- **WHEN** a user opens the Elastic configuration flow
- **THEN** the main form SHALL require only datasource name, Elasticsearch URL, and API Key
- **THEN** the saved instance SHALL use `Log` as the enabled capability
- **THEN** the saved instance SHALL default `authType` to `apiKey` and `logIndexPattern` to `logs-*`

### Requirement: Elastic Log test SHALL support AgentSRE log consumption
The system SHALL test Elastic Log connectivity without validating unenabled APM capabilities, and SHALL warn when logs cannot be confirmed as service-filterable for AgentSRE.

#### Scenario: Log-only test skips unenabled capabilities
- **WHEN** a user tests an Elastic instance with only `Log` enabled
- **THEN** the system SHALL validate only Elastic connection and Log query capability
- **THEN** missing Trace, Topology, ServiceInstance, RedMetric, or Metrics data SHALL NOT cause the test to fail

#### Scenario: Log query works but service filter fields are missing
- **WHEN** Elastic can return logs but recent samples do not include recognizable pod or namespace fields
- **THEN** the system SHALL return a warning rather than a hard failure
- **THEN** the warning SHALL explain that AgentSRE may not be able to filter service logs

### Requirement: Elastic field handling SHALL use common Log field defaults
The system SHALL default to ECS / OTel / Kubernetes log field conventions for Log queries and SHALL keep custom field mapping UI out of this change.

#### Scenario: Standard ECS fields are used by default
- **WHEN** an Elastic instance uses standard ECS, OTel, or Kubernetes log field names
- **THEN** the provider SHALL resolve service, host, container, namespace, pod, log level, and message fields without requiring custom mapping in the APO UI

#### Scenario: Non-standard deployment needs custom mapping
- **WHEN** an Elastic instance does not follow the default field conventions
- **THEN** custom mapping SHALL be treated as follow-up work outside this change
- **THEN** the current UI SHALL not expose a fully arbitrary field mapping system

### Requirement: Full Elastic Observability capabilities SHALL remain follow-up work
The system SHALL not treat `Trace`, `Metrics`, `Topology`, `ServiceInstance`, or `RedMetric` as delivered by this change.

#### Scenario: User expects full APM onboarding
- **WHEN** a user needs Elastic APM traces, metrics, topology, service instances, or RED metrics
- **THEN** the capability SHALL require a later change with UI, API contract, coverage, provider semantics, and end-to-end validation
