## MODIFIED Requirements

### Requirement: Widget Configuration
The system SHALL support widget title, description, chart type, query configuration, selectable or free-form unit configuration, and decimal places. The unit configuration SHALL distinguish system-provided preset units from user-entered custom suffixes. The system MUST display preset-specific unit behavior controls only when a selected preset supports them. The system MUST NOT expose widget size editing, literal dashboard variable fields such as `$service`, or legend, threshold, empty-state text, or error-state text fields unless those fields produce visible runtime behavior.

#### Scenario: Creator edits widget configuration
- **WHEN** the creator changes supported widget configuration in the side panel
- **THEN** the system updates the selected widget preview and marks the dashboard as unsaved

#### Scenario: Creator opens widget configuration
- **WHEN** the creator opens the selected-widget configuration panel
- **THEN** unsupported or non-effective fields are not displayed

#### Scenario: Creator selects preset unit
- **WHEN** the creator selects a system-provided unit preset such as bytes, duration, or percent
- **THEN** the editor stores the preset identity separately from the rendered unit suffix and displays any behavior switch supported by that preset

#### Scenario: Creator enters custom unit
- **WHEN** the creator enters a unit value that is not selected from the system-provided presets
- **THEN** the editor stores it as a custom suffix and does not display preset-only unit behavior controls
