import json
import sqlite3
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import import_grafana_dashboard


class ImportGrafanaDashboardTest(unittest.TestCase):
    def setUp(self):
        self.tmp = tempfile.TemporaryDirectory()
        self.root = Path(self.tmp.name)
        self.db_path = self.root / "apo.db"
        self.json_path = self.root / "grafana.json"
        self._init_db()
        self._write_grafana_json()

    def tearDown(self):
        self.tmp.cleanup()

    def _init_db(self):
        conn = sqlite3.connect(self.db_path)
        try:
            conn.executescript(
                """
                CREATE TABLE user (
                  user_id integer PRIMARY KEY,
                  username text
                );
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

    def _write_grafana_json(self):
        dashboard = {
            "title": "Cluster Overview",
            "uid": "k8s_views_global",
            "templating": {
                "list": [
                    {"name": "datasource", "label": "Data Source", "type": "datasource"},
                    {
                        "name": "cluster_id",
                        "label": "Cluster",
                        "type": "query",
                        "includeAll": True,
                        "allValue": ".*",
                        "current": {"value": "$__all"},
                    },
                    {
                        "name": "resolution",
                        "label": "Resolution",
                        "type": "custom",
                        "query": "1m, 5m",
                        "current": {"value": "1m"},
                    },
                    {
                        "name": "job_name",
                        "label": "Node Job",
                        "type": "query",
                        "query": {
                            "query": 'label_values(node_cpu_seconds_total{cluster_id=~"$cluster_id"}, job_name)'
                        },
                        "current": {"value": "integrations/unix"},
                    },
                    {
                        "name": "namespace",
                        "label": "Namespace",
                        "type": "query",
                        "query": {
                            "query": 'query_result(group by (namespace) (container_memory_rss{cluster_id=~"$cluster_id"}))'
                        },
                        "current": {"value": "default"},
                    },
                    {
                        "name": "pod",
                        "label": "Pod",
                        "type": "query",
                        "query": {
                            "query": 'query_result(group by (pod) (container_memory_rss{namespace="$namespace",cluster_id=~"$cluster_id"}))'
                        },
                        "current": {"value": "$__all"},
                        "includeAll": True,
                    },
                    {
                        "name": "service_name",
                        "label": "Service",
                        "type": "query",
                        "query": {
                            "query": 'query_result(group by (svc_name) (kindling_span_trace_duration_nanoseconds_count{namespace="$namespace",cluster_id=~"$cluster_id"}))'
                        },
                        "current": {"value": "checkout"},
                    },
                    {
                        "name": "related_nodes",
                        "label": "Related Nodes",
                        "type": "query",
                        "query": {
                            "query": 'query_result(group by (node_name) (container_memory_rss{pod=~"$pod",cluster_id=~"$cluster_id"}))'
                        },
                        "current": {"value": "$__all"},
                        "includeAll": True,
                    },
                ]
            },
            "panels": [
                {"id": 1, "title": "Overview", "type": "row"},
                {
                    "id": 2,
                    "title": "CPU Usage",
                    "type": "timeseries",
                    "gridPos": {"x": 0, "y": 1, "w": 12, "h": 8},
                    "fieldConfig": {"defaults": {"unit": "percentunit", "decimals": 1}},
                    "targets": [
                        {
                            "refId": "A",
                            "expr": 'avg(rate(node_cpu_seconds_total{cluster_id=~"$cluster_id",job_name="$job_name"}[$__rate_interval]))',
                            "legendFormat": "CPU",
                        },
                        {
                            "refId": "B",
                            "expr": "hidden_metric",
                            "legendFormat": "Hidden",
                            "hide": True,
                        },
                    ],
                },
                {
                    "id": 3,
                    "title": "Node Count",
                    "type": "stat",
                    "gridPos": {"x": 12, "y": 1, "w": 6, "h": 4},
                    "targets": [
                        {
                            "refId": "A",
                            "expr": 'count(up{cluster_id=~"$cluster_id"})',
                            "legendFormat": "",
                        }
                    ],
                },
                {
                    "id": 4,
                    "title": "JVM Instance",
                    "type": "row",
                    "panels": [
                        {
                            "id": 5,
                            "title": "Basic Info",
                            "type": "table",
                            "gridPos": {"x": 0, "y": 5, "w": 24, "h": 4},
                            "targets": [
                                {
                                    "refId": "A",
                                    "expr": 'jvm_cpu_count{instance=~"$jvm_instance", cluster_id=~"$cluster_id"}',
                                    "legendFormat": "__auto",
                                }
                            ],
                            "transformations": [
                                {
                                    "id": "organize",
                                    "options": {
                                        "excludeByName": {
                                            "Time": True,
                                            "Value": True,
                                            "__name__": True,
                                            "instance": True,
                                            "service_name": False,
                                            "service_version": False,
                                            "telemetry_sdk_language": False,
                                        },
                                        "indexByName": {
                                            "service_name": 0,
                                            "service_version": 1,
                                            "telemetry_sdk_language": 2,
                                            "container_id": 3,
                                        },
                                        "renameByName": {
                                            "service_name": "服务名",
                                            "service_version": "版本",
                                            "telemetry_sdk_language": "语言",
                                            "container_id": "Container Id",
                                        },
                                    },
                                },
                            ],
                        },
                    ],
                },
            ],
        }
        self.json_path.write_text(json.dumps(dashboard), encoding="utf-8")

    def test_imports_grafana_json_as_normal_dashboard(self):
        result = import_grafana_dashboard.import_dashboard(
            grafana_json_path=self.json_path,
            db_path=self.db_path,
            owner="admin",
        )

        conn = sqlite3.connect(self.db_path)
        try:
            row = conn.execute(
                "SELECT name, owner_id, visibility, config FROM custom_dashboard WHERE id = ?",
                (result.dashboard_id,),
            ).fetchone()
            self.assertEqual(("Cluster Overview", 100, "everyone"), row[:3])

            config = json.loads(row[3])
            self.assertEqual("Cluster Overview", config["name"])
            self.assertEqual(
                ["cluster_id", "resolution", "job_name", "namespace", "pod", "service_name", "related_nodes"],
                [v["name"] for v in config["variables"]],
            )
            variables_by_name = {variable["name"]: variable for variable in config["variables"]}
            self.assertEqual("cluster", variables_by_name["cluster_id"]["type"])
            self.assertEqual(["cluster", "cluster_id", "clusterId", "clusterID", "cluster_name"], variables_by_name["cluster_id"]["labels"])
            self.assertEqual("namespace", variables_by_name["namespace"]["type"])
            self.assertEqual(["namespace", "namespace_name", "namespaceName", "src_namespace", "dst_namespace"], variables_by_name["namespace"]["labels"])
            self.assertEqual(".*", variables_by_name["namespace"]["value"])
            self.assertEqual(".*", variables_by_name["namespace"]["defaultValue"])
            self.assertEqual("pod", variables_by_name["pod"]["type"])
            self.assertEqual(["pod", "pod_name", "src_pod", "dst_pod"], variables_by_name["pod"]["labels"])
            self.assertEqual("service", variables_by_name["service_name"]["type"])
            self.assertEqual(["svc", "svc_name", "service", "service_name"], variables_by_name["service_name"]["labels"])
            self.assertTrue(variables_by_name["service_name"]["includeAll"])
            self.assertEqual(".*", variables_by_name["service_name"]["allValue"])
            self.assertEqual(".*", variables_by_name["service_name"]["value"])
            self.assertEqual(".*", variables_by_name["service_name"]["defaultValue"])
            self.assertEqual("label", variables_by_name["related_nodes"]["type"])
            self.assertEqual(["node_name"], variables_by_name["related_nodes"]["labels"])
            self.assertTrue(variables_by_name["related_nodes"]["includeAll"])
            self.assertEqual(".*", variables_by_name["related_nodes"]["allValue"])
            self.assertEqual("label", variables_by_name["job_name"]["type"])
            self.assertEqual(["job_name"], variables_by_name["job_name"]["labels"])
            self.assertEqual(3, len(config["widgets"]))
            self.assertEqual(["line", "stat", "table"], [w["type"] for w in config["widgets"]])
            self.assertEqual(
                {"unit": "%", "unitPreset": "percent", "autoScale": True, "decimals": 1},
                {key: config["widgets"][0]["display"][key] for key in ("unit", "unitPreset", "autoScale", "decimals")},
            )
            self.assertEqual(1, len(config["widgets"][0]["query"]["queries"]))
            self.assertIn('cluster_id=~"${cluster_id:regex}"', config["widgets"][0]["query"]["expression"])
            self.assertIn("[$__rate_interval]", config["widgets"][0]["query"]["expression"])
            self.assertNotIn("* 100", config["widgets"][0]["query"]["expression"])
            self.assertEqual(
                [
                    {"name": "service_name", "label": "服务名"},
                    {"name": "service_version", "label": "版本"},
                    {"name": "telemetry_sdk_language", "label": "语言"},
                    {"name": "container_id", "label": "Container Id"},
                ],
                config["widgets"][2]["display"]["tableColumns"],
            )
            self.assertIn("instance", config["widgets"][2]["display"]["tableHiddenColumns"])
            self.assertFalse(config["widgets"][2]["display"]["showValueColumn"])
            self.assertIn('instance=~"${jvm_instance:regex}"', config["widgets"][2]["query"]["expression"])

            route = conn.execute(
                "SELECT router_to FROM router WHERE router_id = ?",
                (result.router_id,),
            ).fetchone()[0]
            self.assertEqual(f"/custom-dashboard/{result.dashboard_id}", route)

            menu = conn.execute(
                'SELECT "key", parent_id, router_id FROM menu_item WHERE item_id = ?',
                (result.menu_item_id,),
            ).fetchone()
            self.assertEqual(result.menu_key, menu[0])
            self.assertEqual(result.router_id, menu[2])

            mappings = conn.execute(
                "SELECT mapped_id, mapped_type FROM feature_mapping WHERE feature_id = ? ORDER BY mapped_type",
                (result.feature_id,),
            ).fetchall()
            self.assertEqual([(result.menu_item_id, "menu"), (result.router_id, "router")], mappings)

            permission = conn.execute(
                "SELECT type, subject_id, subject_type, permission_id FROM auth_permission WHERE permission_id = ?",
                (result.feature_id,),
            ).fetchone()
            self.assertEqual(("feature", 100, "user", result.feature_id), permission)
        finally:
            conn.close()

    def test_variable_mapping_contract(self):
        variables = import_grafana_dashboard.convert_variables([
            {"name": "endpoint_name", "label": "Endpoint", "type": "query"},
            {"name": "nodeName", "label": "Node", "type": "query"},
            {
                "name": "related_nodes",
                "label": "Related Nodes",
                "type": "query",
                "query": {"query": 'query_result(group by (node_name) (container_memory_rss{pod=~"$pod"}))'},
                "includeAll": True,
                "current": {"value": "$__all"},
            },
            {
                "name": "related_pid",
                "label": "PID",
                "type": "query",
                "query": {"query": 'label_values(kindling_span_trace_duration_nanoseconds_count{svc_name=~"$service_name"},pid)'},
            },
            {
                "name": "jvm_instance",
                "label": "JVM Instance",
                "type": "query",
                "regex": '/instance="(?<text>[^"]+)/g',
                "query": {"query": 'query_result(jvm_cpu_count{host_name=~"$pod"})'},
            },
        ])
        by_name = {variable["name"]: variable for variable in variables}

        self.assertEqual("endpoint", by_name["endpoint_name"]["type"])
        self.assertEqual(["endpoint", "endpoint_name", "content_key", "contentKey"], by_name["endpoint_name"]["labels"])
        self.assertEqual("instance", by_name["nodeName"]["type"])
        self.assertEqual(["instance", "instance_name", "node_name", "nodeName", "host_name", "hostName"], by_name["nodeName"]["labels"])
        self.assertEqual("label", by_name["related_nodes"]["type"])
        self.assertEqual(["node_name"], by_name["related_nodes"]["labels"])
        self.assertTrue(by_name["related_nodes"]["includeAll"])
        self.assertEqual("label", by_name["related_pid"]["type"])
        self.assertEqual(["pid"], by_name["related_pid"]["labels"])
        self.assertEqual("label", by_name["jvm_instance"]["type"])
        self.assertEqual(["instance"], by_name["jvm_instance"]["labels"])

    def test_static_custom_matcher_variables_are_inlined(self):
        dashboard = import_grafana_dashboard.convert_grafana_dashboard({
            "title": "Disk Dashboard",
            "templating": {
                "list": [
                    {
                        "name": "diskdevices",
                        "label": "Disk Devices",
                        "type": "custom",
                        "options": [
                            {
                                "selected": True,
                                "text": "[a-z]+|nvme[0-9]+n[0-9]+",
                                "value": "[a-z]+|nvme[0-9]+n[0-9]+",
                            }
                        ],
                        "current": {"value": "[a-z]+|nvme[0-9]+n[0-9]+"},
                    },
                    {
                        "name": "resolution",
                        "label": "Resolution",
                        "type": "custom",
                        "query": "1m, 5m",
                        "current": {"value": "1m"},
                    },
                ],
            },
            "panels": [
                {
                    "id": 7,
                    "title": "Disk Reads",
                    "type": "timeseries",
                    "targets": [
                        {
                            "refId": "A",
                            "expr": 'irate(node_disk_reads_completed_total{device=~"$diskdevices"}[$__rate_interval])',
                        }
                    ],
                }
            ],
        })

        self.assertEqual(["resolution"], [variable["name"] for variable in dashboard["variables"]])
        expression = dashboard["widgets"][0]["query"]["expression"]
        self.assertIn('device=~"[a-z]+|nvme[0-9]+n[0-9]+"', expression)
        self.assertNotIn("diskdevices", expression)

    def test_grafana_visual_types_and_thresholds_map_to_native_widgets(self):
        panels = [
            {
                "id": 1,
                "title": "Gauge",
                "type": "gauge",
                "fieldConfig": {
                    "defaults": {
                        "thresholds": {
                            "steps": [
                                {"color": "green", "value": None},
                                {"color": "yellow", "value": 70},
                                {"color": "red", "value": 90},
                            ]
                        }
                    }
                },
                "targets": [{"refId": "A", "expr": "up"}],
            },
            {
                "id": 2,
                "title": "Pie",
                "type": "piechart",
                "targets": [{"refId": "A", "expr": "sum by (status) (up)"}],
            },
            {
                "id": 3,
                "title": "Legacy Graph",
                "type": "graph",
                "targets": [{"refId": "A", "expr": "rate(requests_total[5m])"}],
            },
        ]

        widgets = import_grafana_dashboard.convert_panels(panels)

        self.assertEqual(["barGauge", "bar", "line"], [widget["type"] for widget in widgets])
        self.assertEqual(70.0, widgets[0]["display"]["warningThreshold"])
        self.assertEqual(90.0, widgets[0]["display"]["criticalThreshold"])

    def test_interval_and_grafana_value_variables_use_native_custom_options(self):
        variables = import_grafana_dashboard.convert_variables([
            {
                "name": "interval",
                "label": "Interval",
                "type": "interval",
                "query": "15s,1m,5m",
                "current": {"text": "1m", "value": "1m"},
            },
            {
                "name": "grouping",
                "label": "Grouping",
                "type": "custom",
                "options": [
                    {"text": "None", "value": ""},
                    {"text": "By transport", "value": ",transport"},
                ],
                "current": {"text": "None", "value": ""},
            },
            {"name": "adhoc", "type": "adhoc"},
        ])

        self.assertEqual(["interval", "grouping"], [variable["name"] for variable in variables])
        self.assertEqual("custom", variables[0]["type"])
        self.assertEqual(["15s", "1m", "5m"], [option["value"] for option in variables[0]["options"]])
        self.assertEqual("", variables[1]["options"][0]["value"])
        self.assertEqual(
            'rate(metric_total[${interval:raw}])',
            import_grafana_dashboard.normalize_query('rate(metric_total[${interval:value}])', ["interval"]),
        )


if __name__ == "__main__":
    unittest.main()
