## ADDED Requirements

### Requirement: Preset Unit Value Formatting
The system SHALL format numeric widget values according to configured preset unit behavior in dashboard viewer, editor preview, table cells, stat widgets, chart tooltips, chart axes, and chart legend values. The system MUST render custom units as numeric values plus the custom suffix without automatic conversion.

#### Scenario: Viewer renders byte preset with automatic scaling
- **WHEN** a widget value is numeric and its display configuration uses the bytes preset with automatic scaling enabled
- **THEN** the system renders the value using a readable binary unit such as KiB, MiB, GiB, or TiB

#### Scenario: Viewer renders percent preset from ratio input
- **WHEN** a widget value is numeric and its display configuration uses the percent preset with ratio conversion enabled
- **THEN** the system multiplies the value by 100 for display and appends the percent suffix

#### Scenario: Viewer renders custom unit
- **WHEN** a widget value is numeric and its display configuration uses a custom unit suffix
- **THEN** the system appends the suffix without applying preset scaling or percent conversion

#### Scenario: Viewer renders non-numeric value
- **WHEN** a widget value is not numeric
- **THEN** the system leaves the value unchanged rather than applying unit conversion
