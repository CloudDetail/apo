INSERT INTO feature (feature_id, feature_name, parent_id)
SELECT 1, "服务概览", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 1)
UNION ALL
SELECT 2, "日志检索", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 2)
UNION ALL
SELECT 3, "故障现场日志", 2
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 3)
UNION ALL
SELECT 4, "全量日志", 2
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 4)
UNION ALL
SELECT 5, "链路追踪", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 5)
UNION ALL
SELECT 6, "故障现场链路", 5
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 6)
UNION ALL
SELECT 7, "全量链路", 5
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 7)
UNION ALL
SELECT 8, "全局资源大盘", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 8)
UNION ALL
SELECT 9, "应用基础设施大盘", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 9)
UNION ALL
SELECT 10, "应用指标大盘", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 10)
UNION ALL
SELECT 11, "中间件大盘", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 11)
UNION ALL
SELECT 12, "MySQL大盘", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 12)
UNION ALL
SELECT 13, "告警规则", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 13)
UNION ALL
SELECT 14, "配置中心", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 14)
UNION ALL
SELECT 15, "服务健康", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 15)
UNION ALL
SELECT 16, "系统管理", NULL
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 16)
UNION ALL
SELECT 17, "用户管理", 16
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 17)
UNION ALL
SELECT 18, "菜单管理", 16
    WHERE NOT EXISTS (SELECT 1 FROM feature WHERE feature_id = 18);
