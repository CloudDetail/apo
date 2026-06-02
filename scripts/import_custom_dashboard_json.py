#!/usr/bin/env python3
"""Import an APO custom dashboard JSON into the local SQLite database.

Input is the same dashboard JSON shape stored in custom_dashboard.config:
name, description, variables, widgets, visibility, and optional UI-only fields.
The script creates a normal dashboard row plus its dynamic menu, route, feature,
translations, and owner permission.
"""

from __future__ import annotations

import argparse
import copy
import json
import os
from pathlib import Path
from typing import Any

from import_grafana_dashboard import (
    VISIBILITY_VALUES,
    default_db_path,
    inline_static_matcher_variable,
    insert_dashboard_config,
    variable_query_label,
    variable_used_as_label_matcher,
)


def import_dashboard_json(
    dashboard_json_path: str | Path,
    db_path: str | Path,
    owner: str = "admin",
    name: str | None = None,
    description: str | None = None,
    visibility: str | None = None,
    dashboard_id: str | None = None,
):
    raw = json.loads(Path(dashboard_json_path).read_text(encoding="utf-8"))
    dashboard = normalize_dashboard_json(raw, name=name, description=description, visibility=visibility)
    return insert_dashboard_config(
        dashboard=dashboard,
        db_path=db_path,
        owner=owner,
        dashboard_id=dashboard_id,
    )


def normalize_dashboard_json(
    raw: dict[str, Any],
    name: str | None = None,
    description: str | None = None,
    visibility: str | None = None,
) -> dict[str, Any]:
    if not isinstance(raw, dict):
        raise ValueError("dashboard JSON must be an object")
    source = raw.get("dashboard") if isinstance(raw.get("dashboard"), dict) else raw
    dashboard = copy.deepcopy(source)

    if name is not None:
        dashboard["name"] = name
    if description is not None:
        dashboard["description"] = description
    if visibility is not None:
        dashboard["visibility"] = visibility

    dashboard["id"] = ""
    dashboard["ownerId"] = 0
    dashboard["creator"] = ""
    dashboard["visibility"] = dashboard.get("visibility") or "everyone"
    dashboard["teamIds"] = dashboard.get("teamIds") if isinstance(dashboard.get("teamIds"), list) else []
    dashboard["variables"] = dashboard.get("variables") if isinstance(dashboard.get("variables"), list) else []
    dashboard["widgets"] = dashboard.get("widgets") if isinstance(dashboard.get("widgets"), list) else []
    dashboard["timeRange"] = dashboard.get("timeRange") if isinstance(dashboard.get("timeRange"), dict) else {}
    dashboard["refreshSeconds"] = int(dashboard.get("refreshSeconds") or 0)
    dashboard["canEdit"] = False
    dashboard["canDelete"] = False
    dashboard["canCopy"] = False
    dashboard["createdAt"] = 0
    dashboard["updatedAt"] = 0
    dashboard["lastVisitedAt"] = 0
    dashboard["description"] = str(dashboard.get("description") or "")
    normalize_variable_labels(dashboard["variables"])
    inline_static_custom_matcher_variables(dashboard)
    return dashboard


def normalize_variable_labels(variables: list[dict[str, Any]]) -> None:
    for variable in variables:
        name = str(variable.get("name") or "")
        query_label = variable_query_label(variable)
        labels = variable.get("labels")
        if variable.get("type") == "label" and query_label and (not labels or labels == [name]):
            variable["labels"] = [query_label]


def inline_static_custom_matcher_variables(dashboard: dict[str, Any]) -> None:
    static_values = {
        name: value
        for variable in dashboard.get("variables") or []
        if (name := str(variable.get("name") or "").strip())
        if (value := static_custom_variable_value(variable))
    }
    if not static_values:
        return
    used = {
        name
        for name in static_values
        if any(variable_used_as_label_matcher(expression, name) for expression in dashboard_widget_expressions(dashboard))
    }
    if not used:
        return
    for widget in dashboard.get("widgets") or []:
        query = widget.get("query") if isinstance(widget, dict) else None
        if not isinstance(query, dict):
            continue
        if isinstance(query.get("expression"), str):
            query["expression"] = inline_static_values(query["expression"], static_values, used)
        for item in query.get("queries") or []:
            if isinstance(item, dict) and isinstance(item.get("expression"), str):
                item["expression"] = inline_static_values(item["expression"], static_values, used)
        remove_variable_dependency_references(query, used)
    dashboard["variables"] = [
        variable
        for variable in dashboard.get("variables") or []
        if str(variable.get("name") or "").strip() not in used
    ]


def static_custom_variable_value(variable: dict[str, Any]) -> str:
    if variable.get("type") != "custom" or variable.get("multi"):
        return ""
    options = variable.get("options")
    if isinstance(options, list):
        values = [
            str(option.get("value") if isinstance(option, dict) else option).strip()
            for option in options
            if (option.get("value") if isinstance(option, dict) else option) not in (None, "")
        ]
        if len(values) == 1:
            return values[0]
        if len(values) > 1:
            return ""
    query = str(variable.get("query") or "").strip()
    if query and "," not in query:
        return query
    return ""


def dashboard_widget_expressions(dashboard: dict[str, Any]) -> list[str]:
    expressions: list[str] = []
    for widget in dashboard.get("widgets") or []:
        query = widget.get("query") if isinstance(widget, dict) else None
        if not isinstance(query, dict):
            continue
        if isinstance(query.get("expression"), str):
            expressions.append(query["expression"])
        expressions.extend(
            item["expression"]
            for item in query.get("queries") or []
            if isinstance(item, dict) and isinstance(item.get("expression"), str)
        )
    return expressions


def inline_static_values(expression: str, static_values: dict[str, str], used: set[str]) -> str:
    result = expression
    for name in used:
        result = inline_static_matcher_variable(result, name, static_values[name])
    return result


def remove_variable_dependency_references(query: dict[str, Any], names: set[str]) -> None:
    if not names:
        return
    params = query.get("params")
    if not isinstance(params, dict):
        return
    dependencies = params.get("variableDependencies")
    if isinstance(dependencies, str):
        filtered = [
            dependency
            for dependency in dependencies.split()
            if variable_dependency_name(dependency) not in names
        ]
        if filtered:
            params["variableDependencies"] = " ".join(filtered)
        else:
            params.pop("variableDependencies", None)
    elif isinstance(dependencies, list):
        params["variableDependencies"] = [
            dependency
            for dependency in dependencies
            if variable_dependency_name(str(dependency)) not in names
        ]


def variable_dependency_name(dependency: str) -> str:
    token = dependency.strip()
    if token.startswith("${") and token.endswith("}"):
        token = token[2:-1]
    elif token.startswith("$"):
        token = token[1:]
    return token.split(":", 1)[0]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Import an APO custom dashboard JSON into custom_dashboard as a normal dashboard."
    )
    parser.add_argument("dashboard_json", help="Path to APO custom dashboard JSON")
    parser.add_argument(
        "--db",
        default=os.environ.get("APO_SQLITE_DB") or str(default_db_path()),
        help="SQLite database path. Defaults to APO_SQLITE_DB or backend/database-apo.db.",
    )
    parser.add_argument("--owner", default="admin", help="Dashboard owner username or user_id")
    parser.add_argument("--name", default=None, help="Override dashboard name")
    parser.add_argument("--description", default=None, help="Override dashboard description")
    parser.add_argument("--visibility", choices=sorted(VISIBILITY_VALUES), default=None)
    parser.add_argument("--dashboard-id", default=None, help="Optional UUID to use for the dashboard")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    result = import_dashboard_json(
        dashboard_json_path=Path(args.dashboard_json),
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
