#!/usr/bin/env python3
"""Import a Grafana dashboard JSON as a normal APO custom dashboard.

This writes directly to the custom dashboard tables. It does not create or
modify dashboard templates.
"""

from __future__ import annotations

import argparse
import json
import math
import os
import re
import sqlite3
import uuid
from dataclasses import dataclass
from datetime import datetime
from pathlib import Path
from typing import Any


VISIBILITY_VALUES = {"everyone", "personal", "team"}
CUSTOM_DASHBOARD_PARENT_KEY = "customDashboard"
CUSTOM_DASHBOARD_PARENT_FEATURE = "我的大盘"
RESOURCE_VARIABLES = {
    "cluster": ("cluster", ["cluster", "cluster_id", "clusterId", "clusterID", "cluster_name"]),
    "cluster_id": ("cluster", ["cluster", "cluster_id", "clusterId", "clusterID", "cluster_name"]),
    "clusterId": ("cluster", ["cluster", "cluster_id", "clusterId", "clusterID", "cluster_name"]),
    "clusterID": ("cluster", ["cluster", "cluster_id", "clusterId", "clusterID", "cluster_name"]),
    "cluster_name": ("cluster", ["cluster", "cluster_id", "clusterId", "clusterID", "cluster_name"]),
    "namespace": ("namespace", ["namespace", "namespace_name", "namespaceName", "src_namespace", "dst_namespace"]),
    "namespace_name": ("namespace", ["namespace", "namespace_name", "namespaceName", "src_namespace", "dst_namespace"]),
    "namespaceName": ("namespace", ["namespace", "namespace_name", "namespaceName", "src_namespace", "dst_namespace"]),
    "src_namespace": ("namespace", ["namespace", "namespace_name", "namespaceName", "src_namespace", "dst_namespace"]),
    "dst_namespace": ("namespace", ["namespace", "namespace_name", "namespaceName", "src_namespace", "dst_namespace"]),
    "pod": ("pod", ["pod", "pod_name", "src_pod", "dst_pod"]),
    "pod_name": ("pod", ["pod", "pod_name", "src_pod", "dst_pod"]),
    "src_pod": ("pod", ["pod", "pod_name", "src_pod", "dst_pod"]),
    "dst_pod": ("pod", ["pod", "pod_name", "src_pod", "dst_pod"]),
    "service": ("service", ["svc", "svc_name", "service", "service_name"]),
    "service_name": ("service", ["svc", "svc_name", "service", "service_name"]),
    "svc": ("service", ["svc", "svc_name", "service", "service_name"]),
    "svc_name": ("service", ["svc", "svc_name", "service", "service_name"]),
    "endpoint": ("endpoint", ["endpoint", "endpoint_name", "content_key", "contentKey"]),
    "endpoint_name": ("endpoint", ["endpoint", "endpoint_name", "content_key", "contentKey"]),
    "content_key": ("endpoint", ["endpoint", "endpoint_name", "content_key", "contentKey"]),
    "contentKey": ("endpoint", ["endpoint", "endpoint_name", "content_key", "contentKey"]),
    "instance": ("instance", ["instance", "instance_name", "node_name", "nodeName", "host_name", "hostName"]),
    "instance_name": ("instance", ["instance", "instance_name", "node_name", "nodeName", "host_name", "hostName"]),
    "node_name": ("instance", ["instance", "instance_name", "node_name", "nodeName", "host_name", "hostName"]),
    "nodeName": ("instance", ["instance", "instance_name", "node_name", "nodeName", "host_name", "hostName"]),
    "host_name": ("instance", ["instance", "instance_name", "node_name", "nodeName", "host_name", "hostName"]),
    "hostName": ("instance", ["instance", "instance_name", "node_name", "nodeName", "host_name", "hostName"]),
}


@dataclass(frozen=True)
class ImportResult:
    dashboard_id: str
    name: str
    widget_count: int
    variable_count: int
    route: str
    menu_key: str
    router_id: int
    menu_item_id: int
    feature_id: int


def import_dashboard(
    grafana_json_path: str | Path,
    db_path: str | Path,
    owner: str = "admin",
    name: str | None = None,
    description: str | None = None,
    visibility: str = "everyone",
    dashboard_id: str | None = None,
) -> ImportResult:
    grafana_path = Path(grafana_json_path)
    database_path = Path(db_path)
    grafana = json.loads(grafana_path.read_text(encoding="utf-8"))
    dashboard = convert_grafana_dashboard(
        grafana,
        name=name,
        description=description,
        visibility=visibility,
    )
    if not dashboard["widgets"]:
        raise ValueError("Grafana JSON did not contain any importable panels")

    return insert_dashboard_config(
        dashboard=dashboard,
        db_path=database_path,
        owner=owner,
        dashboard_id=dashboard_id,
    )


def insert_dashboard_config(
    dashboard: dict[str, Any],
    db_path: str | Path,
    owner: str = "admin",
    dashboard_id: str | None = None,
) -> ImportResult:
    database_path = Path(db_path)
    visibility = str(dashboard.get("visibility") or "everyone")
    if visibility not in VISIBILITY_VALUES:
        raise ValueError(f"visibility must be one of: {', '.join(sorted(VISIBILITY_VALUES))}")
    if not dashboard.get("name"):
        raise ValueError("dashboard JSON must include a name")
    if not isinstance(dashboard.get("widgets"), list):
        raise ValueError("dashboard JSON must include a widgets array")
    if not isinstance(dashboard.get("variables"), list):
        dashboard["variables"] = []

    dashboard_id = dashboard_id or str(uuid.uuid4())
    validate_dashboard_id(dashboard_id)
    route = f"/custom-dashboard/{dashboard_id}"
    menu_key = custom_dashboard_menu_key(dashboard_id)
    now = datetime.now().strftime("%Y-%m-%d %H:%M:%S")
    config_json = json.dumps(dashboard, ensure_ascii=False, separators=(",", ":"))

    conn = sqlite3.connect(database_path)
    try:
        conn.execute("BEGIN")
        owner_id = resolve_owner_id(conn, owner)
        parent_menu_id, parent_order = resolve_parent_menu(conn)
        parent_feature_id = resolve_parent_feature(conn)
        ensure_route_available(conn, route)
        ensure_menu_key_available(conn, menu_key)
        ensure_dashboard_id_available(conn, dashboard_id)

        conn.execute(
            """
            INSERT INTO custom_dashboard
              (id, name, description, owner_id, visibility, team_ids, config, created_at, updated_at)
            VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
            """,
            (
                dashboard_id,
                dashboard["name"],
                dashboard["description"],
                owner_id,
                visibility,
                "[]",
                config_json,
                now,
                now,
            ),
        )
        conn.execute(
            "INSERT INTO router (router_to, custom, hide_time_selector) VALUES (?, ?, ?)",
            (route, 1, 0),
        )
        router_id = int(
            conn.execute("SELECT router_id FROM router WHERE router_to = ?", (route,)).fetchone()[0]
        )
        conn.execute(
            """
            INSERT INTO menu_item ("key", icon, parent_id, router_id, sort_order)
            VALUES (?, ?, ?, ?, ?)
            """,
            (menu_key, "customDashboard", parent_menu_id, router_id, parent_order + 1),
        )
        menu_item_id = int(
            conn.execute('SELECT item_id FROM menu_item WHERE "key" = ?', (menu_key,)).fetchone()[0]
        )
        conn.execute(
            "INSERT INTO feature (feature_name, parent_id, custom) VALUES (?, ?, ?)",
            (menu_key, parent_feature_id, 1),
        )
        feature_id = int(
            conn.execute("SELECT feature_id FROM feature WHERE feature_name = ?", (menu_key,)).fetchone()[0]
        )
        conn.execute(
            "INSERT INTO feature_mapping (feature_id, mapped_id, mapped_type) VALUES (?, ?, ?)",
            (feature_id, menu_item_id, "menu"),
        )
        conn.execute(
            "INSERT INTO feature_mapping (feature_id, mapped_id, mapped_type) VALUES (?, ?, ?)",
            (feature_id, router_id, "router"),
        )
        conn.execute(
            "INSERT INTO auth_permission (type, subject_id, subject_type, permission_id) VALUES (?, ?, ?, ?)",
            ("feature", owner_id, "user", feature_id),
        )
        insert_translations(conn, feature_id, menu_item_id, dashboard["name"])
        conn.commit()
    except Exception:
        conn.rollback()
        raise
    finally:
        conn.close()

    return ImportResult(
        dashboard_id=dashboard_id,
        name=dashboard["name"],
        widget_count=len(dashboard["widgets"]),
        variable_count=len(dashboard["variables"]),
        route=route,
        menu_key=menu_key,
        router_id=router_id,
        menu_item_id=menu_item_id,
        feature_id=feature_id,
    )


def convert_grafana_dashboard(
    grafana: dict[str, Any],
    name: str | None = None,
    description: str | None = None,
    visibility: str = "everyone",
) -> dict[str, Any]:
    title = name or str(grafana.get("title") or "Imported Grafana Dashboard")
    desc = description
    if desc is None:
        source = grafana.get("uid") or grafana.get("id") or title
        desc = f"Imported from Grafana dashboard {source}"
    raw_variables = grafana.get("templating", {}).get("list", [])
    static_custom_values = static_custom_matcher_variables(raw_variables, grafana.get("panels", []))
    variables = convert_variables(raw_variables, exclude_names=set(static_custom_values))
    widgets = convert_panels(grafana.get("panels", []), static_custom_values=static_custom_values)
    return {
        "id": "",
        "name": title,
        "description": desc,
        "ownerId": 0,
        "creator": "",
        "visibility": visibility,
        "teamIds": [],
        "variables": variables,
        "widgets": widgets,
        "timeRange": {},
        "refreshSeconds": 0,
        "canEdit": False,
        "canDelete": False,
        "canCopy": False,
        "createdAt": 0,
        "updatedAt": 0,
        "lastVisitedAt": 0,
    }


def convert_variables(raw_variables: list[dict[str, Any]], exclude_names: set[str] | None = None) -> list[dict[str, Any]]:
    variables: list[dict[str, Any]] = []
    excluded = exclude_names or set()
    for raw in raw_variables:
        name = str(raw.get("name") or "").strip()
        if not name or raw.get("type") in {"datasource", "adhoc"} or name in excluded:
            continue
        current_value = variable_current_value(raw)
        all_value = str(raw.get("allValue") or ".*")
        if current_value in {"$__all", "__all"}:
            current_value = all_value
        var_type = variable_type(raw)
        is_resource_variable = name in RESOURCE_VARIABLES
        if is_resource_variable:
            current_value = all_value

        variable: dict[str, Any] = {
            "name": name,
            "label": str(raw.get("label") or name),
            "type": var_type,
            "value": current_value,
            "multi": True if is_resource_variable else bool(raw.get("multi")),
            "required": False,
            "defaultValue": current_value,
        }
        if is_resource_variable or raw.get("includeAll"):
            variable["includeAll"] = True
            variable["allValue"] = all_value
            if not variable["value"]:
                variable["value"] = all_value
                variable["defaultValue"] = all_value

        options = variable_options(raw)
        if options:
            variable["options"] = options
        query = variable_query(raw)
        if query:
            variable["query"] = normalize_query(query, [name])
        labels = labels_for_variable(raw)
        if labels:
            variable["labels"] = labels
        variables.append(variable)
    return variables


def static_custom_matcher_variables(raw_variables: list[dict[str, Any]], raw_panels: list[dict[str, Any]]) -> dict[str, str]:
    candidates = {
        str(raw.get("name") or "").strip(): value
        for raw in raw_variables
        if (value := static_custom_variable_value(raw))
    }
    if not candidates:
        return {}
    used = variable_names_used_as_label_matchers(raw_panels, set(candidates))
    return {name: value for name, value in candidates.items() if name in used}


def static_custom_variable_value(raw: dict[str, Any]) -> str:
    name = str(raw.get("name") or "").strip()
    if not name or raw.get("type") != "custom" or raw.get("multi"):
        return ""
    options = variable_options(raw)
    if len(options) == 1:
        return str(options[0]["value"]).strip()
    if len(options) > 1:
        return ""
    query = variable_query(raw).strip()
    if query and "," not in query:
        return query
    return ""


def variable_current_value(raw: dict[str, Any]) -> str:
    current = raw.get("current") or {}
    value = current.get("value")
    if isinstance(value, list):
        return ",".join(str(item) for item in value)
    if value is None or value == "":
        value = current.get("text")
    if isinstance(value, list):
        return ",".join(str(item) for item in value)
    return str(value or "")


def variable_type(raw: dict[str, Any]) -> str:
    name = str(raw.get("name") or "")
    resource = RESOURCE_VARIABLES.get(name)
    if resource:
        return resource[0]
    if raw.get("type") in {"custom", "interval"}:
        return "custom"
    return "label"


def variable_options(raw: dict[str, Any]) -> list[dict[str, str]]:
    if isinstance(raw.get("options"), list) and raw["options"]:
        result = []
        for option in raw["options"]:
            value = option.get("value", option.get("text")) if isinstance(option, dict) else option
            text = option.get("text", value) if isinstance(option, dict) else option
            if value is None:
                continue
            result.append({"label": str(text), "value": str(value)})
        return result
    query = raw.get("query")
    if raw.get("type") not in {"custom", "interval"} or not isinstance(query, str):
        return []
    return [{"label": item.strip(), "value": item.strip()} for item in query.split(",") if item.strip()]


def variable_query(raw: dict[str, Any]) -> str:
    query = raw.get("query")
    if isinstance(query, dict):
        return str(query.get("query") or "")
    return str(query or "")


def labels_for_variable(raw: dict[str, Any]) -> list[str]:
    name = str(raw.get("name") or "")
    resource = RESOURCE_VARIABLES.get(name)
    if resource:
        return resource[1]
    if name in {"job_name", "cadvisor_job_name"}:
        return ["job_name"]
    query_label = variable_query_label(raw)
    if query_label:
        return [query_label]
    if raw.get("type") == "query":
        return [name]
    return []


def variable_query_label(raw: dict[str, Any]) -> str:
    query = variable_query(raw)
    if not query:
        return ""
    label_values = re.search(r"label_values\((.+),\s*([A-Za-z_][A-Za-z0-9_]*)\s*\)\s*$", query)
    if label_values:
        return label_values.group(2)
    group_by = re.search(r"group\s+by\s*\(\s*([A-Za-z_][A-Za-z0-9_]*)\s*\)", query)
    if group_by:
        return group_by.group(1)
    regex = str(raw.get("regex") or "")
    regex_label = re.search(r"([A-Za-z_][A-Za-z0-9_]*)=", regex)
    if regex_label:
        return regex_label.group(1)
    return ""


def convert_panels(raw_panels: list[dict[str, Any]], static_custom_values: dict[str, str] | None = None) -> list[dict[str, Any]]:
    panels = list(iter_importable_panels(raw_panels))
    variables = variable_names_from_panels(panels)
    widgets = []
    for panel in panels:
        queries = panel_queries(panel, variables, static_custom_values=static_custom_values or {})
        if not queries:
            continue
        first = queries[0]
        widgets.append(
            {
                "id": f"grafana-{panel.get('id') or len(widgets) + 1}",
                "title": str(panel.get("title") or f"Panel {len(widgets) + 1}"),
                "description": str(panel.get("description") or ""),
                "type": widget_type(panel.get("type")),
                "layout": widget_layout(panel.get("gridPos") or {}),
                "query": {
                    "mode": "custom",
                    "presetId": "",
                    "expression": first["expression"],
                    "params": {},
                    "legend": first["legend"],
                    "queries": queries,
                },
                "display": widget_display(panel),
            }
        )
    return widgets


def iter_importable_panels(raw_panels: list[dict[str, Any]]):
    for panel in raw_panels or []:
        if not isinstance(panel, dict):
            continue
        children = panel.get("panels")
        if panel.get("type") == "row":
            if isinstance(children, list):
                yield from iter_importable_panels(children)
            continue
        yield panel
        if isinstance(children, list):
            yield from iter_importable_panels(children)


def variable_names_from_panels(raw_panels: list[dict[str, Any]]) -> list[str]:
    names: set[str] = set()
    token_pattern = re.compile(r"\$([A-Za-z_][A-Za-z0-9_]*)|\$\{([A-Za-z_][A-Za-z0-9_]*)(?::[^}]*)?\}")
    for panel in raw_panels:
        for target in panel.get("targets") or []:
            expr = str(target.get("expr") or "")
            for match in token_pattern.finditer(expr):
                name = match.group(1) or match.group(2)
                if name and not name.startswith("__"):
                    names.add(name)
    return sorted(names, key=len, reverse=True)


def variable_names_used_as_label_matchers(raw_panels: list[dict[str, Any]], variable_names: set[str]) -> set[str]:
    used: set[str] = set()
    if not variable_names:
        return used
    for panel in iter_importable_panels(raw_panels):
        for target in panel.get("targets") or []:
            expression = str(target.get("expr") or "")
            for name in variable_names:
                if variable_used_as_label_matcher(expression, name):
                    used.add(name)
    return used


def variable_used_as_label_matcher(expression: str, name: str) -> bool:
    token = re.escape(name)
    return bool(re.search(rf'(!=|=~|!~|=)\s*"\$(?:{token}\b|\{{{token}(?::[^}}]*)?\}})"', expression))


def panel_queries(panel: dict[str, Any], variable_names: list[str], static_custom_values: dict[str, str] | None = None) -> list[dict[str, str]]:
    targets = [target for target in panel.get("targets") or [] if target.get("expr") and target.get("hide") is not True]
    if not targets:
        targets = [target for target in panel.get("targets") or [] if target.get("expr")]
    queries = []
    for index, target in enumerate(targets):
        expression = normalize_query(str(target.get("expr") or ""), variable_names, static_custom_values=static_custom_values or {})
        queries.append(
            {
                "id": str(target.get("refId") or f"q-{index + 1}"),
                "expression": expression,
                "legend": str(target.get("legendFormat") or ""),
            }
        )
    return queries


def normalize_query(query: str, variable_names: list[str], static_custom_values: dict[str, str] | None = None) -> str:
    expression = query.strip()
    static_values = static_custom_values or {}
    for name, value in static_values.items():
        expression = inline_static_matcher_variable(expression, name, value)
    for name in variable_names:
        if name in static_values:
            continue
        expression = re.sub(
            rf'=~\s*"\${re.escape(name)}"',
            f'=~"${{{name}:regex}}"',
            expression,
        )
        expression = re.sub(
            rf'=~\s*"\$\{{{re.escape(name)}\}}"',
            f'=~"${{{name}:regex}}"',
            expression,
        )
        expression = expression.replace(f"${{{name}:value}}", f"${{{name}:raw}}")
    return expression


def inline_static_matcher_variable(query: str, name: str, value: str) -> str:
    literal = promql_string_literal_value(value)
    token = re.escape(name)
    return re.sub(
        rf'(!=|=~|!~|=)\s*"\$(?:{token}\b|\{{{token}(?::[^}}]*)?\}})"',
        lambda match: f'{match.group(1)}"{literal}"',
        query,
    )


def promql_string_literal_value(value: str) -> str:
    return str(value).replace("\\", "\\\\").replace('"', '\\"')


def widget_type(grafana_type: Any) -> str:
    return {
        "timeseries": "line",
        "graph": "line",
        "heatmap": "line",
        "nodeGraph": "line",
        "piechart": "bar",
        "gauge": "barGauge",
        "bargauge": "barGauge",
        "stat": "stat",
        "table": "table",
        "barchart": "bar",
    }.get(str(grafana_type or ""), "line")


def widget_layout(grid: dict[str, Any]) -> dict[str, int]:
    raw_w = int(grid.get("w") or 12)
    raw_x = int(grid.get("x") or 0)
    raw_y = int(grid.get("y") or 0)
    width = max(1, min(4, round(raw_w / 6)))
    x = max(0, min(3, math.floor(raw_x / 6)))
    if x + width > 4:
        x = max(0, 4 - width)
    return {"x": x, "y": max(0, math.floor(raw_y / 4)), "w": width, "h": 1}


def widget_display(panel: dict[str, Any]) -> dict[str, Any]:
    display = {
        **display_unit_config(panel_unit(panel)),
        "decimals": panel_decimals(panel),
        "emptyText": "",
        "errorText": "",
    }
    display.update(panel_thresholds(panel))
    table_config = table_transform_display(panel)
    if table_config:
        display.update(table_config)
    return display


def panel_thresholds(panel: dict[str, Any]) -> dict[str, float]:
    thresholds = panel_defaults(panel).get("thresholds") or {}
    steps = thresholds.get("steps") if isinstance(thresholds, dict) else []
    values = [
        float(step["value"])
        for step in steps or []
        if isinstance(step, dict) and isinstance(step.get("value"), (int, float))
    ]
    if not values:
        return {}
    result = {"warningThreshold": values[0]}
    if len(values) > 1:
        result["criticalThreshold"] = values[-1]
    return result


def panel_defaults(panel: dict[str, Any]) -> dict[str, Any]:
    return (panel.get("fieldConfig") or {}).get("defaults") or {}


def panel_unit(panel: dict[str, Any]) -> str:
    return str(panel_defaults(panel).get("unit") or "")


def panel_decimals(panel: dict[str, Any]) -> int:
    value = panel_defaults(panel).get("decimals")
    if isinstance(value, int) and value >= 0:
        return value
    if panel.get("type") == "stat" and not panel_unit(panel):
        return 0
    return 2


def display_unit(unit: str) -> str:
    return {
        "percentunit": "%",
        "bytes": "B",
        "none": "",
    }.get(unit, unit)


def display_unit_config(unit: str) -> dict[str, Any]:
    normalized = str(unit or "")
    if normalized == "percentunit":
        return {"unit": "%", "unitPreset": "percent", "autoScale": True}
    if normalized == "bytes":
        return {"unit": "B", "unitPreset": "bytes", "autoScale": True}
    if normalized == "ms":
        return {"unit": "ms", "unitPreset": "milliseconds", "autoScale": False}
    if normalized == "s":
        return {"unit": "s", "unitPreset": "seconds", "autoScale": False}
    display = display_unit(normalized)
    return {"unit": display, "unitPreset": "custom" if display else "", "autoScale": False}


def table_transform_display(panel: dict[str, Any]) -> dict[str, Any]:
    if panel.get("type") != "table":
        return {}
    organize = next(
        (
            transform.get("options") or {}
            for transform in panel.get("transformations") or []
            if transform.get("id") == "organize"
        ),
        None,
    )
    if not organize:
        return {}
    exclude_by_name = organize.get("excludeByName") or {}
    index_by_name = organize.get("indexByName") or {}
    rename_by_name = organize.get("renameByName") or {}
    hidden_names = sorted(
        name
        for name, excluded in exclude_by_name.items()
        if excluded is True and name not in {"Time", "Value"}
    )
    candidates = set(index_by_name) | set(rename_by_name) | {
        name for name, excluded in exclude_by_name.items() if excluded is False
    }
    columns = []
    for name in sorted(candidates, key=lambda item: (index_by_name.get(item, 10_000), item)):
        if name in {"Time", "Value"} or exclude_by_name.get(name) is True:
            continue
        label = str(rename_by_name.get(name) or name)
        if not name or not label:
            continue
        columns.append({"name": str(name), "label": label})
    config: dict[str, Any] = {}
    if columns:
        config["tableColumns"] = columns
    if hidden_names:
        config["tableHiddenColumns"] = hidden_names
    if exclude_by_name.get("Value") is True:
        config["showValueColumn"] = False
    return config


def resolve_owner_id(conn: sqlite3.Connection, owner: str) -> int:
    if owner.isdigit():
        row = conn.execute("SELECT user_id FROM user WHERE user_id = ?", (int(owner),)).fetchone()
    else:
        row = conn.execute("SELECT user_id FROM user WHERE username = ?", (owner,)).fetchone()
    if row:
        return int(row[0])
    fallback = conn.execute("SELECT user_id FROM user ORDER BY user_id LIMIT 1").fetchone()
    if fallback:
        return int(fallback[0])
    raise ValueError("No user found in database; cannot assign dashboard owner")


def resolve_parent_menu(conn: sqlite3.Connection) -> tuple[int, int]:
    row = conn.execute(
        'SELECT item_id, sort_order FROM menu_item WHERE "key" = ?',
        (CUSTOM_DASHBOARD_PARENT_KEY,),
    ).fetchone()
    if not row:
        raise ValueError('Parent menu "customDashboard" was not found')
    return int(row[0]), int(row[1] or 0)


def resolve_parent_feature(conn: sqlite3.Connection) -> int:
    row = conn.execute(
        "SELECT feature_id FROM feature WHERE feature_name = ?",
        (CUSTOM_DASHBOARD_PARENT_FEATURE,),
    ).fetchone()
    if not row:
        raise ValueError(f'Parent feature "{CUSTOM_DASHBOARD_PARENT_FEATURE}" was not found')
    return int(row[0])


def ensure_route_available(conn: sqlite3.Connection, route: str) -> None:
    if conn.execute("SELECT 1 FROM router WHERE router_to = ?", (route,)).fetchone():
        raise ValueError(f"Router already exists: {route}")


def ensure_menu_key_available(conn: sqlite3.Connection, menu_key: str) -> None:
    if conn.execute('SELECT 1 FROM menu_item WHERE "key" = ?', (menu_key,)).fetchone():
        raise ValueError(f"Menu key already exists: {menu_key}")


def ensure_dashboard_id_available(conn: sqlite3.Connection, dashboard_id: str) -> None:
    if conn.execute("SELECT 1 FROM custom_dashboard WHERE id = ?", (dashboard_id,)).fetchone():
        raise ValueError(f"Dashboard id already exists: {dashboard_id}")


def insert_translations(conn: sqlite3.Connection, feature_id: int, menu_item_id: int, label: str) -> None:
    for language in ("zh", "en"):
        conn.execute(
            """
            INSERT INTO i18n_translation
              (language, translation, field_name, entity_id, entity_type)
            VALUES (?, ?, ?, ?, ?)
            """,
            (language, label, "feature_name", feature_id, "feature"),
        )
        for field_name in ("label", "abbreviation"):
            conn.execute(
                """
                INSERT INTO i18n_translation
                  (language, translation, field_name, entity_id, entity_type)
                VALUES (?, ?, ?, ?, ?)
                """,
                (language, label, field_name, menu_item_id, "menu"),
            )


def validate_dashboard_id(dashboard_id: str) -> None:
    try:
        uuid.UUID(dashboard_id)
    except ValueError as exc:
        raise ValueError("--dashboard-id must be a UUID") from exc


def custom_dashboard_menu_key(dashboard_id: str) -> str:
    return "cd-" + dashboard_id.replace("-", "")[:17]


def default_db_path() -> Path:
    repo_root = Path(__file__).resolve().parents[1]
    candidate = repo_root / "backend" / "database-apo.db"
    if candidate.exists():
        return candidate
    return Path("database-apo.db")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Import a Grafana dashboard JSON into APO custom_dashboard as a normal dashboard."
    )
    parser.add_argument("grafana_json", help="Path to the Grafana dashboard JSON export")
    parser.add_argument(
        "--db",
        default=os.environ.get("APO_SQLITE_DB") or str(default_db_path()),
        help="SQLite database path. Defaults to APO_SQLITE_DB or backend/database-apo.db.",
    )
    parser.add_argument("--owner", default="admin", help="Dashboard owner username or user_id")
    parser.add_argument("--name", default=None, help="Override dashboard name")
    parser.add_argument("--description", default=None, help="Override dashboard description")
    parser.add_argument(
        "--visibility",
        default="everyone",
        choices=sorted(VISIBILITY_VALUES),
        help="Dashboard visibility",
    )
    parser.add_argument("--dashboard-id", default=None, help="Optional UUID to use for the dashboard")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    result = import_dashboard(
        grafana_json_path=Path(args.grafana_json),
        db_path=Path(args.db),
        owner=args.owner,
        name=args.name,
        description=args.description,
        visibility=args.visibility,
        dashboard_id=args.dashboard_id,
    )
    print(
        json.dumps(
            {
                "dashboard_id": result.dashboard_id,
                "name": result.name,
                "widgets": result.widget_count,
                "variables": result.variable_count,
                "route": result.route,
                "db": str(Path(args.db).resolve()),
            },
            ensure_ascii=False,
            indent=2,
        )
    )


if __name__ == "__main__":
    main()
