#!/usr/bin/env python3
"""Import every standalone APO custom dashboard JSON in a directory."""

from __future__ import annotations

import argparse
import json
import os
from pathlib import Path

from import_custom_dashboard_json import import_dashboard_json
from import_grafana_dashboard import default_db_path


def dashboard_files(directory: str | Path) -> list[Path]:
    root = Path(directory)
    if not root.is_dir():
        raise ValueError(f"dashboard directory does not exist: {root}")
    return sorted(path for path in root.glob("*.json") if path.name != "manifest.json")


def import_dashboard_directory(
    directory: str | Path,
    db_path: str | Path,
    owner: str = "admin",
) -> list[dict[str, object]]:
    files = dashboard_files(directory)
    if not files:
        raise ValueError(f"dashboard directory contains no dashboard JSON files: {directory}")
    results = []
    for path in files:
        result = import_dashboard_json(path, db_path=db_path, owner=owner)
        results.append(
            {
                "file": path.name,
                "dashboard_id": result.dashboard_id,
                "name": result.name,
                "widgets": result.widget_count,
                "variables": result.variable_count,
                "route": result.route,
            }
        )
    return results


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description="Import a directory of APO custom dashboard JSON files.")
    parser.add_argument("directory", type=Path, help="Directory containing standalone dashboard JSON files")
    parser.add_argument(
        "--db",
        default=os.environ.get("APO_SQLITE_DB") or str(default_db_path()),
        help="SQLite database path. Defaults to APO_SQLITE_DB or backend/database-apo.db.",
    )
    parser.add_argument("--owner", default="admin", help="Dashboard owner username or user_id")
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    results = import_dashboard_directory(args.directory, db_path=args.db, owner=args.owner)
    print(json.dumps({"imported": len(results), "dashboards": results}, ensure_ascii=False, indent=2))


if __name__ == "__main__":
    main()
