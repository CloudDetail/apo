INSERT INTO target_tags (id,tag_name,"describe",field) VALUES
	 (0,'自定义tag','自定义TAG名','custom'),
	 (1,'服务名','服务名称,如APM系统中的service','serviceName'),
	 (2,'服务端点','提供服务的接口,如HTTP服务的URL','endpoint'),
	 (3,'命名空间','K8s Namespace','namespace'),
	 (4,'POD名','K8s PodName','pod'),
	 (5,'主机名','K8s Node / 服务器 hostname','node'),
	 (6,'进程PID','服务器上进程的pid','pid'),
	 (7,'告警类型','告警对象的类型,如 应用类型 主机类型等.','group'),
	 (8,'数据库服务URL','数据库用于提供服务的地址,如 host:3306','dbURL'),
	 (9,'数据库服务HOST','数据库用于提供服务的地址','dbHost'),
	 (10,'数据库服务IP','数据库用于提供服务的ip','dbIP'),
	 (11,'数据库服务Port','数据库用于提供服务的port','dbPort');

INSERT INTO alert_enrich_rules (enrich_rule_id,source_id,r_type,rule_order,from_field,from_regex,target_tag_id,custom_tag,"schema",schema_source) VALUES
	 ('88b38d0a-0d7d-4266-a4dd-76a920464da3','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.svc_name','',1,'','',''),
	 ('7d6df829-f363-4ab4-b12a-7591777bda77','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.content_key','',2,'','',''),
	 ('961c9a78-54bb-4db7-992b-f17272d87278','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.namespace','',3,'','',''),
	 ('1b8c4702-533e-4930-965d-544ab349b369','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.pod','',4,'','',''),
	 ('c93abcbd-48d7-4cbc-9641-9037a0b6fd45','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.pod_name','',4,'','',''),
	 ('34777e7d-fb49-4ad2-bd8a-b4826fd696d4','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.node','',5,'','',''),
	 ('91195c74-194d-4f90-ba86-d9727c17220e','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.node_name','',5,'','',''),
	 ('a4c0ed18-cba1-4002-988e-2ca9092e4573','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.pid','',6,'','',''),
	 ('257e44eb-09e4-4268-9b09-5a71641b783c','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.group','',7,'','',''),
	 ('abe2a5c1-630f-4b6a-86cc-d31e2a170712','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.db_url','',8,'','',''),
	 ('55df6363-90ef-48c6-9594-9a95b7836fed','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.net_host_name','([^(]+)$',9,'','',''),
	 ('feab2453-0ca8-4322-bcbf-f5f238067a27','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.net_host_name','(\d+\.\d+\.\d+\.\d+)',10,'','',''),
	 ('5622d50a-7aab-4bf6-8754-430995267aac','825079a8-4d05-3507-b347-1272a078f9ff','tagMapping',0,'.net_host_name','(\d+)',11,'','',''),
	 ('f09e6647-2609-4a09-a6ad-47f2120c1633','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.svc_name','',1,'','',''),
	 ('149c2ff0-f73d-4041-b075-0e5316705336','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.content_key','',2,'','',''),
	 ('28c085de-f986-472e-b279-a868ecde9b93','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.namespace','',3,'','',''),
	 ('a04ebf23-29e3-4406-bbd5-4f34fca9f8ac','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.pod','',4,'','',''),
	 ('90f504a1-2d07-4eba-a69e-15cf941e6438','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.pod_name','',4,'','',''),
	 ('89e4f2b7-b298-4675-a0ad-a6ac4df90a81','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.node','',5,'','',''),
	 ('e733c178-5d39-469b-a878-961fc508661b','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.node_name','',5,'','',''),
	 ('37d21136-a7bf-4207-96ff-450e6c2a0772','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.pid','',6,'','',''),
	 ('43e25169-5704-403f-a614-0ac7234e79d5','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.group','',7,'','',''),
	 ('d954c06a-0326-426c-98bc-bcc930ffb356','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.db_url','',8,'','',''),
	 ('a5bfca83-f816-47d9-bd2e-65db287aa1ad','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.net_host_name','([^(]+)$',9,'','',''),
	 ('5483fe02-c1f0-4579-a8db-1faad7af70c8','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.net_host_name','(\d+\.\d+\.\d+\.\d+)',10,'','',''),
	 ('245d5629-d90c-4fa3-a855-2e376cd87ca2','2213d3d5-41da-32a8-9026-22c2bf6aa448','tagMapping',0,'.net_host_name','(\d+)',11,'','','');

INSERT INTO alert_enrich_conditions (enrich_rule_id,source_id,from_field,operation,expr) VALUES
	 ('55df6363-90ef-48c6-9594-9a95b7836fed','825079a8-4d05-3507-b347-1272a078f9ff','.group','match','middleware'),
	 ('feab2453-0ca8-4322-bcbf-f5f238067a27','825079a8-4d05-3507-b347-1272a078f9ff','.group','match','middleware'),
	 ('5622d50a-7aab-4bf6-8754-430995267aac','825079a8-4d05-3507-b347-1272a078f9ff','.group','match','middleware'),
	 ('a5bfca83-f816-47d9-bd2e-65db287aa1ad','2213d3d5-41da-32a8-9026-22c2bf6aa448','.group','match','middleware'),
	 ('5483fe02-c1f0-4579-a8db-1faad7af70c8','2213d3d5-41da-32a8-9026-22c2bf6aa448','.group','match','middleware'),
	 ('245d5629-d90c-4fa3-a855-2e376cd87ca2','2213d3d5-41da-32a8-9026-22c2bf6aa448','.group','match','middleware');

INSERT INTO alert_sources (source_id,source_name,source_type) VALUES
	 ('825079a8-4d05-3507-b347-1272a078f9ff','APO_DEFAULT_ENRICH_RULE_PROMETHEUS','prometheus'),
	 ('2213d3d5-41da-32a8-9026-22c2bf6aa448','APO_DEFAULT_ENRICH_RULE_JSON','json');
