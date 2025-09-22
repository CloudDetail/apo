UPDATE data_plane_type SET capability_spec='["RedMetric","ServiceInstance","Topology","Log","Trace","Metrics"]' where type_name='signoz';
UPDATE data_plane SET capability_spec='["RedMetric","ServiceInstance","Topology","Log","Trace","Metrics"]' where typ='signoz';
UPDATE data_plane_type SET type_name='prometheusmetric' where type_name='prometheus';
UPDATE data_plane SET typ = 'prometheusmetric' where typ = 'prometheus'
