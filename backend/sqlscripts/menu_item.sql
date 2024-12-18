INSERT INTO menu_item (item_id, key, label, router_id, parent_id, abbreviation, insert_page_id, icon, en_label, en_abbreviation)
SELECT 1, "service", "服务概览", 1, NULL, NULL, NULL, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/service.svg", "Service Overview", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 1)
UNION ALL
SELECT 2, "logs", "日志检索", NULL, NULL, NULL, NULL, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/log.svg", "Log Search", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 2)
UNION ALL
SELECT 3, "faultSite", "故障现场日志", 2, 2, NULL, NULL, NULL, "Incident Logs", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 3)
UNION ALL
SELECT 4, "full", "全量日志", 3, 2, NULL, NULL, NULL, "Full Logs", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 4)
UNION ALL
SELECT 5, "trace", "链路追踪", NULL, NULL, NULL, NULL, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/trace.svg", "Traceability", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 5)
UNION ALL
SELECT 6, "faultSiteTrace", "故障现场链路", 15, 5, NULL, NULL, NULL, "Incident Trace", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 6)
UNION ALL
SELECT 7, "fullTrace", "全量链路", 16, 5, NULL, 6, NULL, "Full Trace", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 7)
UNION ALL
SELECT 8, "system", "全局资源大盘", 4, NULL, "全局资源", 4, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg", "Global Resource Dashboard", "Global Resource"
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 9)
UNION ALL
SELECT 9, "basic", "应用基础设施大盘", 5, NULL, "基础设施", 2, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg", "Application Infrastructure Dashboard", "Infrastructure"
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 9)
UNION ALL
SELECT 10, "application", "应用指标大盘", 6, NULL, "应用指标", 1, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg", "Application Metrics Dashboard", "Application Metrics"
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 10)
UNION ALL
SELECT 11, "middleware", "中间件大盘", 7, NULL, "中间件", 3, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg", "Middleware Dashboard", "Middleware"
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 11)
UNION ALL
SELECT 12, "mysql", "MySQL 大盘", 8, NULL, "MySQL", 5, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/dashboard.svg", "MySQL Dashboard", "MySQL"
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 12)
UNION ALL
SELECT 13, "alerts", "告警规则", 9, NULL, NULL, NULL, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/alert.svg", "Alert Rules", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 13)
UNION ALL
SELECT 14, "config", "配置中心", 10, NULL, NULL, NULL, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/setting.svg", "Configuration Center", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 14)
UNION ALL
SELECT 15, "healthy", "服务健康状态", 11, NULL, "服务健康", NULL, NULL, "Service Health Status", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 15)
UNION ALL
SELECT 16, "manage", "系统管理", NULL, NULL, NULL, NULL, "https://apo-front.oss-cn-hangzhou.aliyuncs.com/menu-icon/system.svg", "System Management", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 16)
UNION ALL
SELECT 17, "userManage", "用户管理", 12, 16, NULL, NULL, NULL, "User Management", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 17)
UNION ALL
SELECT 18, "menuManage", "菜单管理", 14, 16, NULL, NULL, NULL, "Menu Management", NULL
    WHERE NOT EXISTS (SELECT 1 FROM menu_item WHERE item_id = 18);
