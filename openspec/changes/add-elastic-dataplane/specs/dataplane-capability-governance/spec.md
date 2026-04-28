## ADDED Requirements

### Requirement: Elastic Log-only instances SHALL persist Log as the enabled capability
The system SHALL persist `Log` as the enabled capability for Elastic instances created through the current APO main flow.

#### Scenario: Elastic type template contains future capabilities
- **WHEN** the Elastic type template lists capabilities beyond `Log`
- **THEN** the current Elastic main flow SHALL save only `Log` for the instance
- **THEN** unimplemented Elastic APM capabilities SHALL NOT be treated as delivered by this change

### Requirement: Elastic Log capability SHALL follow APO query contracts
The system SHALL evaluate Elastic `Log` using APO and AgentSRE log query semantics rather than upstream product terminology.

#### Scenario: Log query is available
- **WHEN** an Elastic provider can query logs by time range
- **THEN** the system SHALL allow `Log` to be supported

#### Scenario: APM-like data exists but is not in scope
- **WHEN** Elastic contains trace, metric, or topology-like data
- **THEN** the current change SHALL NOT automatically mark `Trace`, `Metrics`, `Topology`, `ServiceInstance`, or `RedMetric` as delivered

### Requirement: Test datasource SHALL retain backward-compatible warning signaling
The dataplane test flow SHALL validate only Elastic Log for the current instance and SHALL retain backward-compatible overall success signaling.

#### Scenario: Unenabled capabilities do not participate in testing
- **WHEN** a user tests a dataplane instance with only `Log` enabled
- **THEN** the system SHALL validate only the log capability
- **THEN** missing Trace, Topology, ServiceInstance, RedMetric, or Metrics data SHALL NOT cause the test to fail

#### Scenario: Test distinguishes success warning and failure through compatible fields
- **WHEN** the datasource test completes
- **THEN** the system SHALL keep top-level `success` and `message` fields usable by existing APO callers
- **THEN** `success=true` with a non-empty message SHALL be treated by the APO UI as a warning

#### Scenario: No recent data becomes warning instead of immediate failure
- **WHEN** authentication and query wiring are valid but the current validation time window returns no relevant data
- **THEN** the system SHALL report a `warning`
- **THEN** the system SHALL NOT treat that condition as an automatic hard failure

### Requirement: Save behavior SHALL allow Log-only warnings
The system SHALL allow saving Elastic Log-only instances when Log validation returns a non-blocking warning.

#### Scenario: Failed Log capability blocks save
- **WHEN** Elastic Log test returns a hard failure
- **THEN** the dataplane instance SHALL NOT be saved

#### Scenario: Warning does not block save
- **WHEN** Elastic Log test returns `success=true` with a warning message
- **THEN** the dataplane instance SHALL remain saveable
- **THEN** the user SHALL be informed that the datasource is connected but some data is currently missing or degraded

### Requirement: Full capability governance SHALL remain follow-up work
The system SHALL not require this change to complete generic dataplane capability selection, coverage refresh, or structured save blocking for all providers.

#### Scenario: Later change restores full governance
- **WHEN** a later change adds instance-level capability selection and coverage governance
- **THEN** it SHALL define API contract, UI behavior, persistence, coverage aggregation, and tests separately
