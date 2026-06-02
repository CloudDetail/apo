## Why

Custom dashboard stat widgets can display raw metric values such as bytes as very long numbers, which makes operational dashboards hard to scan. The current `display.unit` field is only a suffix, so the system cannot distinguish supported unit semantics from user-entered custom text.

## What Changes

- Make widget unit configuration support both selectable preset units and free-form custom suffixes.
- Show unit behavior controls only for selected preset units, not for custom suffixes.
- Add preset unit formatting for bytes, milliseconds, seconds, and percent ratio display.
- Preserve existing custom suffix behavior for user-entered units.
- Preserve compatibility for existing dashboards that only store `display.unit`.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `custom-dashboard-editor`: Unit configuration changes from a plain text input to a selectable/free-form control with preset-specific behavior switches.
- `custom-dashboard-viewer`: Widget numeric rendering uses preset unit behavior when configured and falls back to suffix formatting for custom units.
- `custom-dashboard-querying`: Preview/result presentation uses the same numeric display configuration while leaving query execution semantics unchanged.

## Impact

- Frontend custom dashboard editor, widget renderer, display formatting helpers, i18n, and behavior tests.
- Existing dashboard JSON remains readable; unsupported or custom units continue to render as numeric value plus suffix.
- No backend API or database schema migration is expected because widget display config is persisted as dashboard JSON.
