INSERT INTO router (router_id, router_to, hide_time_selector)
SELECT 1, "/service", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 1)
UNION ALL
SELECT 2, "/logs/fault-site", true WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 2)
UNION ALL
SELECT 3, "/logs/full", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 3)
UNION ALL
SELECT 4, "/system-dashboard", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 4)
UNION ALL
SELECT 5, "/basic-dashboard", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 5)
UNION ALL
SELECT 6, "/application-dashboard", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 6)
UNION ALL
SELECT 7, "/middleware-dashboard", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 7)
UNION ALL
SELECT 8, "/mysql-dashboard", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 8)
UNION ALL
SELECT 9, "/alerts", true WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 9)
UNION ALL
SELECT 10, "/config", true WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 10)
UNION ALL
SELECT 11, "/healthy-service", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 11)
UNION ALL
SELECT 12, "/system/user-manage", true WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 12)
UNION ALL
SELECT 13, "/service/info", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 13)
UNION ALL
SELECT 14, "/system/menu-manage", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 14)
UNION ALL
SELECT 15, "/trace/fault-site", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 15)
UNION ALL
SELECT 16, "/trace/full", false WHERE NOT EXISTS (SELECT 1 FROM router WHERE router_id = 16);