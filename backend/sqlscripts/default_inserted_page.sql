INSERT INTO insert_page (page_id, url, type)
SELECT 1, 'grafana/d/b0102ebf-9e5e-4f21-80aa-9c2565cd3dcb/originx-polaris-metrics-service-level', 'grafana'
    WHERE NOT EXISTS (SELECT 1 FROM insert_page WHERE page_id = 1)
UNION ALL
SELECT 2, 'grafana/d/adst2iva9181se/e59fba-e7a180-e8aebe-e696bd-e68385-e586b5', 'grafana'
    WHERE NOT EXISTS (SELECT 1 FROM insert_page WHERE page_id = 2)
UNION ALL
SELECT 3, 'grafana/dashboards/f/edwu5b9rkv94wb/', 'grafana'
    WHERE NOT EXISTS (SELECT 1 FROM insert_page WHERE page_id = 3)
UNION ALL
SELECT 4, 'grafana/d/k8s_views_global/e99b86-e7bea4-e680bb-e8a788', 'grafana'
    WHERE NOT EXISTS (SELECT 1 FROM insert_page WHERE page_id = 4)
UNION ALL
SELECT 5, 'grafana/d/0D6dTg3Zk/mysql-e68c87-e6a087', 'grafana'
    WHERE NOT EXISTS (SELECT 1 FROM insert_page WHERE page_id = 5)
UNION ALL
SELECT 6, '/jaeger/search', 'jaeger'
    WHERE NOT EXISTS (SELECT 1 FROM insert_page WHERE page_id = 6);
