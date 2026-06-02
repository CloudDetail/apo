## ADDED Requirements

### Requirement: Dashboard Variable Editing
The system SHALL allow creators to enable supported dashboard-local variables discovered from widget metric queries while excluding APO global page-context controls from dashboard variable configuration.

#### Scenario: Creator opens dashboard variable configuration
- **WHEN** a creator opens dashboard variable configuration
- **THEN** the system shows supported dashboard-local variable suggestions and does not show data group or time range as configurable variables

#### Scenario: Creator saves dashboard variables
- **WHEN** a creator saves a dashboard after editing variables
- **THEN** the saved dashboard definition excludes data group and time range variables

#### Scenario: Creator previews with selected data group
- **WHEN** a creator previews widgets from the editor
- **THEN** preview requests use the current APO top-level data group rather than a dashboard-local data group variable
