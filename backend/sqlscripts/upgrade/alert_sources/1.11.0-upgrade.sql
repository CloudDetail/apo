UPDATE alert_sources SET params = '{}',enable_pull = false,last_pull_mill_ts = 0 WHERE params IS NULL;

INSERT INTO alert_sources (source_id,source_name,source_type,params,enable_pull,last_pull_mill_ts) VALUES
    ('563a44a3-839f-3e23-adff-e00ac8a3e18f','APO_DEFAULT_ENRICH_RULE_DATADOG','datadog','{}',false,0);

