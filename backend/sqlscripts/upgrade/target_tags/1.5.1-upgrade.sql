-- id is an auto-increment primary key
-- 0 has a special meaning during execution and must not be used

UPDATE target_tags SET id = 100 WHERE id = 0;
UPDATE alert_enrich_rules SET target_tag_id = 100 WHERE target_tag_id = 0;