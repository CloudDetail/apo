## Context

Custom dashboard widgets currently store `display.unit` as a string suffix and `display.decimals` as numeric precision. The same formatter is used by stat widgets, table values, chart tooltips, and chart axes. This keeps the model simple, but it cannot tell whether `B` means a supported byte unit that can be scaled or a user-entered custom suffix.

The dashboard editor already persists display configuration inside dashboard JSON, so this change can extend that JSON without database migration. The target users are dashboard creators and operators who need stat values such as memory usage to be readable at a glance.

## Goals / Non-Goals

**Goals:**
- Distinguish supported preset unit semantics from custom suffix text.
- Format preset byte, duration, and percent values in Stat, table, chart tooltip, chart axis, and preview rendering.
- Keep custom units as plain suffixes with no hidden conversion.
- Keep old dashboard JSON compatible when only `display.unit` exists.

**Non-Goals:**
- No backend query changes or metric value mutation before persistence.
- No exhaustive Grafana unit catalog.
- No automatic unit inference from arbitrary custom text.
- No migration script for existing dashboard rows.

## Decisions

1. Store preset identity separately from the display suffix.

   New display config uses `unitPreset` to identify supported behavior, `unit` as the rendered suffix/legacy compatibility value, and `autoScale` for preset-specific conversion. A custom unit stores `unitPreset: "custom"` and uses `unit` as suffix text so hand-entered `B` or `%` can remain literal.

   Alternative considered: infer behavior from `unit` string. That would make hand-entered `B` or `%` behave differently from other custom input and would be hard to explain.

2. Use one behavior switch field but preset-specific labels.

   `autoScale` remains the stored boolean. The editor label changes by preset: bytes/duration use automatic scaling language; percent uses "input value is ratio" language. This avoids adding a second field that only applies to percent.

   Alternative considered: add `valueTransform: "ratioToPercent"`. That is clearer internally but larger than needed for the first unit set.

3. Normalize legacy units conservatively.

   Existing `display.unit` values that exactly match supported presets (`B`, `bytes`, `ms`, `s`, `%`) are treated as preset units during render. Other values remain custom suffixes. This improves existing dashboards while preserving unknown custom units.

   Alternative considered: require users to reselect presets. That avoids inference but leaves the current unreadable stat problem unresolved for existing widgets.

4. Format values after query execution only.

   Raw query values remain unchanged. Formatting applies only in frontend display helpers. For percent preset with `autoScale`, display multiplies by 100; threshold inputs remain in displayed units, and chart threshold rendering converts them back to raw values where needed.

## Risks / Trade-offs

- Legacy `%` dashboards might already query values in `0-100` form → Keep `autoScale` disabled by default for percent legacy normalization unless explicitly enabled by user.
- Axis labels may become less precise after scaling → Tooltips and Stat full value formatting keep decimal control, and chart axes already favor compact readability.
- Custom users may expect typed `B` to scale → The editor exposes a selectable `Bytes (B)` preset; typed custom text remains literal by design.
