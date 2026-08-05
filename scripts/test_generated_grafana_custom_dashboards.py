import json
import re
import sys
import tempfile
import unittest
from pathlib import Path

sys.path.insert(0, str(Path(__file__).resolve().parent))

import import_custom_dashboard_json
import generate_default_custom_dashboard_sql
import generate_grafana_custom_dashboards


class GeneratedGrafanaCustomDashboardsTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls):
        cls.root = Path(__file__).resolve().parents[1] / "backend/static/custom-dashboards"
        cls.manifest = json.loads((cls.root / "manifest.json").read_text(encoding="utf-8"))
        cls.files = sorted(path for path in cls.root.glob("*.json") if path.name != "manifest.json")

    def test_outputs_are_standalone_custom_dashboard_configs(self):
        self.assertEqual(16, len(self.files))
        self.assertEqual(16, self.manifest["generatedDashboardCount"])
        self.assertEqual(
            {item["file"] for item in self.manifest["dashboards"]},
            {path.name for path in self.files},
        )

        required_fields = {
            "id",
            "name",
            "description",
            "ownerId",
            "creator",
            "visibility",
            "teamIds",
            "variables",
            "widgets",
            "timeRange",
            "refreshSeconds",
            "canEdit",
            "canDelete",
            "canCopy",
            "createdAt",
            "updatedAt",
            "lastVisitedAt",
        }
        for path in self.files:
            dashboard = json.loads(path.read_text(encoding="utf-8"))
            self.assertTrue(required_fields.issubset(dashboard), path.name)
            self.assertNotIn("dashboard", dashboard, path.name)
            self.assertNotIn("category", dashboard, path.name)
            self.assertEqual("", dashboard["id"], path.name)
            self.assertEqual(dashboard, import_custom_dashboard_json.normalize_dashboard_json(dashboard), path.name)

    def test_generated_coverage_localization_and_variables(self):
        chinese = re.compile(r"[\u4e00-\u9fff]")
        token_pattern = re.compile(r"\$([A-Za-z_][A-Za-z0-9_]*)|\$\{([A-Za-z_][A-Za-z0-9_]*)(?::[^}]*)?\}")
        widget_count = 0
        query_count = 0
        for path in self.files:
            dashboard = json.loads(path.read_text(encoding="utf-8"))
            variables = {str(variable.get("name") or "") for variable in dashboard["variables"]}
            self.assertGreater(len(dashboard["widgets"]), 0, path.name)
            self.assertLessEqual(len(dashboard["widgets"]), 120, path.name)
            for variable in dashboard["variables"]:
                self.assertRegex(str(variable.get("label") or ""), chinese, f"{path.name}: {variable}")
            for widget in dashboard["widgets"]:
                widget_count += 1
                self.assertRegex(str(widget.get("title") or ""), chinese, f"{path.name}: {widget.get('title')}")
                for query in (widget.get("query") or {}).get("queries") or []:
                    query_count += 1
                    expression = str(query.get("expression") or "")
                    self.assertTrue(expression, f"{path.name}: {widget.get('title')}")
                    self.assertNotIn("SELECT ", expression.upper(), f"{path.name}: {widget.get('title')}")
                    used = {
                        match.group(1) or match.group(2)
                        for match in token_pattern.finditer(expression)
                    }
                    self.assertFalse(
                        used - variables - {"__rate_interval", "__interval"},
                        f"{path.name}: {widget.get('title')}",
                    )

        self.assertEqual(506, widget_count)
        self.assertEqual(937, query_count)
        self.assertEqual(widget_count, self.manifest["importedPanelCount"])
        self.assertEqual(query_count, self.manifest["promQLQueryCount"])
        self.assertEqual(70, self.manifest["unsupportedPanelCount"])
        self.assertEqual(3, len(self.manifest["unsupportedSources"]))

        node_dashboard = json.loads((self.root / "node-exporter.json").read_text(encoding="utf-8"))
        self.assertEqual("宿主机监控指标", node_dashboard["name"])
        self.assertEqual(116, len(node_dashboard["widgets"]))
        untranslated_legend = re.compile(
            r"\b(?:memory|process(?:es)?|used|available|waiting|sockets?|received|sent|"
            r"errors?|time|state|current|maximum|allocated|active|inactive|failed|read|"
            r"write|scrape|duration|success|online|connections?|messages?|datagrams?|"
            r"segments?|queue|packets?|operational|physical|critical|temperature|threads?|"
            r"load|open|files?|bytes)\b",
            re.IGNORECASE,
        )
        node_titles = {widget["title"] for widget in node_dashboard["widgets"]}
        self.assertIn("内核缓存（Slab）内存", node_titles)
        self.assertIn("块设备回弹缓冲区内存", node_titles)
        self.assertIn("节点指标采集状态", node_titles)
        for widget in node_dashboard["widgets"]:
            description = str(widget.get("description") or "")
            self.assertNotIn("Node Exporter", description, widget["title"])
            self.assertNotIn("Niced", description, widget["title"])
            for query in (widget.get("query") or {}).get("queries") or []:
                legend = re.sub(r"\{\{[^}]+\}\}", "", str(query.get("legend") or ""))
                self.assertNotRegex(legend, untranslated_legend, f"{widget['title']}: {legend}")
        self.assertNotIn("node-exporter-overview.json", {path.name for path in self.files})
        self.assertNotIn("node-exporter-system.json", {path.name for path in self.files})
        self.assertNotIn("node-exporter-storage-network.json", {path.name for path in self.files})

    def test_backend_initialization_sql_matches_bundled_dashboards(self):
        sql_root = self.root.parents[1] / "sqlscripts" / "default_custom_dashboards"
        with tempfile.TemporaryDirectory() as tmp:
            generated_root = Path(tmp)
            result = generate_default_custom_dashboard_sql.generate_sql(self.root, generated_root)

            self.assertEqual(16, result["dashboards"])
            for dialect in generate_default_custom_dashboard_sql.DIALECTS:
                filename = f"default_custom_dashboards-{dialect}.sql"
                checked_in = (sql_root / filename).read_text(encoding="utf-8")
                generated = (generated_root / filename).read_text(encoding="utf-8")
                self.assertEqual(checked_in, generated, filename)
                self.assertEqual(16, generated.count("INSERT INTO custom_dashboard"), filename)
                for item in self.manifest["dashboards"]:
                    dashboard_id = generate_default_custom_dashboard_sql.stable_dashboard_id(item["id"])
                    self.assertIn(dashboard_id, generated, f"{filename}: {item['id']}")

    def test_generator_removes_only_legacy_split_node_exporter_outputs(self):
        with tempfile.TemporaryDirectory() as tmp:
            output_root = Path(tmp)
            for filename in generate_grafana_custom_dashboards.LEGACY_NODE_EXPORTER_OUTPUTS:
                (output_root / filename).write_text("legacy", encoding="utf-8")
            preserved = output_root / "unrelated.json"
            preserved.write_text("keep", encoding="utf-8")

            generate_grafana_custom_dashboards.remove_legacy_node_exporter_outputs(output_root)

            for filename in generate_grafana_custom_dashboards.LEGACY_NODE_EXPORTER_OUTPUTS:
                self.assertFalse((output_root / filename).exists())
            self.assertEqual("keep", preserved.read_text(encoding="utf-8"))


if __name__ == "__main__":
    unittest.main()
