import json
import sqlite3
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import import_custom_dashboard_json
import import_custom_dashboard_directory


class ImportCustomDashboardJSONTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.db_path = self.root / "apo.db"
        self.dashboard_path = self.root / "dashboard.json"
        self._init_db()
        self._write_dashboard_json()

    def tearDown(self):
        self.tmp.cleanup()

    def _init_db(self):
        conn = sqlite3.connect(self.db_path)
        try:
            conn.executescript(
                """
                CREATE TABLE user (user_id integer PRIMARY KEY, username text);
                CREATE TABLE custom_dashboard (
                  id varchar(64) PRIMARY KEY,
                  name varchar(128) NOT NULL,
                  description text,
                  owner_id integer NOT NULL,
                  visibility varchar(32) NOT NULL,
                  team_ids text,
                  config text,
                  last_visited_at datetime,
                  created_at datetime,
                  updated_at datetime,
                  deleted_at datetime
                );
                CREATE TABLE router (
                  router_id integer PRIMARY KEY AUTOINCREMENT,
                  router_to varchar(200),
                  custom numeric DEFAULT false,
                  hide_time_selector numeric
                );
                CREATE TABLE menu_item (
                  item_id integer PRIMARY KEY AUTOINCREMENT,
                  "key" varchar(20),
                  icon varchar(150),
                  parent_id integer,
                  router_id integer,
                  sort_order integer
                );
                CREATE TABLE feature (
                  feature_id integer PRIMARY KEY AUTOINCREMENT,
                  feature_name varchar(20),
                  parent_id integer,
                  custom numeric DEFAULT false
                );
                CREATE TABLE feature_mapping (
                  id integer PRIMARY KEY AUTOINCREMENT,
                  feature_id integer,
                  mapped_id integer,
                  mapped_type text
                );
                CREATE TABLE auth_permission (
                  id integer PRIMARY KEY AUTOINCREMENT,
                  type text,
                  subject_id integer,
                  subject_type varchar(10),
                  permission_id integer
                );
                CREATE TABLE i18n_translation (
                  id integer PRIMARY KEY AUTOINCREMENT,
                  language varchar(20),
                  translation varchar(50),
                  field_name varchar(20),
                  entity_id integer,
                  entity_type varchar(20)
                );
                INSERT INTO user (user_id, username) VALUES (100, 'admin');
                INSERT INTO router (router_to, custom, hide_time_selector)
                  VALUES ('/custom-dashboard', 0, 0);
                INSERT INTO menu_item ("key", icon, parent_id, router_id, sort_order)
                  VALUES ('customDashboard', 'customDashboard', NULL, 1, 25);
                INSERT INTO feature (feature_name, parent_id, custom)
                  VALUES ('我的大盘', NULL, 0);
                """
            )
            conn.commit()
        finally:
            conn.close()

    def _write_dashboard_json(self):
        dashboard = {
            "id": "ignored-source-id",
            "name": "APO Converted Dashboard",
            "description": "Converted by AI",
            "ownerId": 999,
            "creator": "ignored",
            "visibility": "personal",
            "variables": [
                {
                    "name": "cluster_id",
                    "label": "Cluster",
                    "type": "cluster",
                    "value": ".*",
                    "multi": True,
                    "required": False,
                    "defaultValue": ".*",
                    "labels": ["cluster_id"],
                    "includeAll": True,
                    "allValue": ".*",
                },
                {
                    "name": "diskdevices",
                    "label": "diskdevices",
                    "type": "custom",
                    "value": "[a-z]+|nvme[0-9]+n[0-9]+",
                    "multi": False,
                    "required": False,
                    "defaultValue": "[a-z]+|nvme[0-9]+n[0-9]+",
                    "options": [
                        {
                            "label": "[a-z]+|nvme[0-9]+n[0-9]+",
                            "value": "[a-z]+|nvme[0-9]+n[0-9]+",
                        }
                    ],
                    "query": "[a-z]+|nvme[0-9]+n[0-9]+",
                },
                {
                    "name": "related_pid",
                    "label": "related_pid",
                    "type": "label",
                    "value": ".*",
                    "multi": True,
                    "required": False,
                    "defaultValue": ".*",
                    "includeAll": True,
                    "allValue": ".*",
                    "query": 'label_values(kindling_span_trace_duration_nanoseconds_count{svc_name=~"$service_name"},pid)',
                    "labels": ["related_pid"],
                }
            ],
            "widgets": [
                {
                    "id": "w-cpu",
                    "title": "CPU",
                    "description": "",
                    "type": "line",
                    "layout": {"x": 0, "y": 0, "w": 2, "h": 1},
                    "query": {
                        "mode": "custom",
                        "presetId": "",
                        "expression": 'sum(rate(node_cpu_seconds_total{cluster_id=~"${cluster_id:regex}"}[$__rate_interval]))',
                        "params": {},
                        "legend": "CPU",
                        "queries": [
                            {
                                "id": "A",
                                "expression": 'sum(rate(node_cpu_seconds_total{cluster_id=~"${cluster_id:regex}"}[$__rate_interval]))',
                                "legend": "CPU",
                            }
                        ],
                    },
                    "display": {"unit": "%", "decimals": 2},
                },
                {
                    "id": "w-disk",
                    "title": "Disk Reads",
                    "description": "",
                    "type": "line",
                    "layout": {"x": 2, "y": 0, "w": 2, "h": 1},
                    "query": {
                        "mode": "custom",
                        "presetId": "",
                        "expression": 'irate(node_disk_reads_completed_total{cluster_id=~"${cluster_id:regex}",device=~"${diskdevices:regex}"}[$__rate_interval])',
                        "params": {
                            "variableDependencies": "$cluster_id $diskdevices $related_pid",
                        },
                        "legend": "Reads",
                        "queries": [
                            {
                                "id": "A",
                                "expression": 'irate(node_disk_reads_completed_total{cluster_id=~"${cluster_id:regex}",device=~"${diskdevices:regex}"}[$__rate_interval])',
                                "legend": "Reads",
                            }
                        ],
                    },
                    "display": {"unit": "iops", "decimals": 2},
                }
            ],
            "timeRange": {"from": "now-1h", "to": "now"},
            "refreshSeconds": 30,
            "canEdit": True,
            "canDelete": True,
            "createdAt": 123,
            "updatedAt": 456,
            "lastVisitedAt": 789,
        }
        self.dashboard_path.write_text(json.dumps(dashboard), encoding="utf-8")

    def test_imports_apo_dashboard_json_and_resets_runtime_fields(self):
        result = import_custom_dashboard_json.import_dashboard_json(
            dashboard_json_path=self.dashboard_path,
            db_path=self.db_path,
            owner="admin",
            visibility="everyone",
        )

        conn = sqlite3.connect(self.db_path)
        try:
            row = conn.execute(
                "SELECT name, description, owner_id, visibility, config FROM custom_dashboard WHERE id = ?",
                (result.dashboard_id,),
            ).fetchone()
            self.assertEqual(("APO Converted Dashboard", "Converted by AI", 100, "everyone"), row[:4])
            config = json.loads(row[4])
            self.assertEqual("", config["id"])
            self.assertEqual(0, config["ownerId"])
            self.assertEqual("", config["creator"])
            self.assertEqual("everyone", config["visibility"])
            self.assertFalse(config["canEdit"])
            self.assertFalse(config["canDelete"])
            self.assertEqual(0, config["createdAt"])
            self.assertEqual(["cluster_id", "related_pid"], [variable["name"] for variable in config["variables"]])
            self.assertEqual(["pid"], next(variable for variable in config["variables"] if variable["name"] == "related_pid")["labels"])
            self.assertEqual(2, len(config["widgets"]))
            self.assertEqual("CPU", config["widgets"][0]["title"])
            disk_expr = config["widgets"][1]["query"]["expression"]
            self.assertIn('device=~"[a-z]+|nvme[0-9]+n[0-9]+"', disk_expr)
            self.assertNotIn("diskdevices", disk_expr)
            self.assertEqual(
                "$cluster_id $related_pid",
                config["widgets"][1]["query"]["params"]["variableDependencies"],
            )

            route = conn.execute(
                "SELECT router_to FROM router WHERE router_id = ?",
                (result.router_id,),
            ).fetchone()[0]
            self.assertEqual(f"/custom-dashboard/{result.dashboard_id}", route)
            permission = conn.execute(
                "SELECT subject_id, subject_type FROM auth_permission WHERE permission_id = ?",
                (result.feature_id,),
            ).fetchone()
            self.assertEqual((100, "user"), permission)
        finally:
            conn.close()

    def test_imports_dashboard_directory_and_skips_manifest(self):
        dashboard_dir = self.root / "dashboards"
        dashboard_dir.mkdir()
        first = json.loads(self.dashboard_path.read_text(encoding="utf-8"))
        second = dict(first)
        second["name"] = "Second Dashboard"
        (dashboard_dir / "first.json").write_text(json.dumps(first), encoding="utf-8")
        (dashboard_dir / "second.json").write_text(json.dumps(second), encoding="utf-8")
        (dashboard_dir / "manifest.json").write_text(json.dumps({"dashboards": 2}), encoding="utf-8")

        results = import_custom_dashboard_directory.import_dashboard_directory(
            dashboard_dir,
            db_path=self.db_path,
            owner="admin",
        )

        self.assertEqual(["first.json", "second.json"], [result["file"] for result in results])
        conn = sqlite3.connect(self.db_path)
        try:
            self.assertEqual(2, conn.execute("SELECT COUNT(*) FROM custom_dashboard").fetchone()[0])
        finally:
            conn.close()

if __name__ == "__main__":
    unittest.main()
