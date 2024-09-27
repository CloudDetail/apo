package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

var metricData = response.GetMetricPQLResponse{
	AlertMetricsData: []model.AlertMetricsData{
		{
			Name:   "平均请求延时",
			PQL:    "sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_sum[1m]))/ sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_count[1m]))/1000000",
			Labels: []string{"content_key", "svc_name", "group", "severity"},
			Unit:   "s",
		},
		{
			Name:   "请求错误率",
			PQL:    "sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_count{is_error=\"true\"}[1m]))/ sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_count[1m]))",
			Labels: []string{"content_key", "svc_name"},
			Unit:   "%",
		},
		{
			Name:   "磁盘使用率",
			PQL:    "((node_filesystem_avail_bytes * 100) / node_filesystem_size_bytes and ON (instance_name, device, mountpoint) node_filesystem_readonly == 0) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Labels: []string{"device", "fstype", "instance_name", "job", "job_name", "mountpoint", "nodename"},
			Unit:   "%",
		},
		{
			Name:   "网络入吞吐量",
			PQL:    "(sum by (instance_name) (rate(node_network_receive_bytes_total[2m])) / 1024 / 1024) *  on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Labels: []string{"instance_name", "nodename"},
			Unit:   "MB",
		},
		{
			Name:   "磁盘读速度",
			PQL:    "(sum by (instance_name) (rate(node_disk_read_bytes_total[2m])) / 1024 / 1024) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Labels: []string{"instance_name", "nodename"},
			Unit:   "MB/s",
		},
		{
			Name:   "磁盘写速度",
			PQL:    "(sum by (instance_name) (rate(node_disk_written_bytes_total[2m])) / 1024 / 1024 ) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Labels: []string{"instance_name", "nodename"},
			Unit:   "MB/s",
		},
		{
			Name:   "CPU负载",
			PQL:    "(sum by (instance_name) (avg by (mode, instance_name) (rate(node_cpu_seconds_total{mode!=\"idle\"}[2m])))) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Labels: []string{"instance_name", "nodename"},
			Unit:   "core",
		},
		{
			Name:   "CPU IO Wait",
			PQL:    "(avg by (instance_name) (rate(node_cpu_seconds_total{mode=\"iowait\"}[5m])) * 100) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Labels: []string{"instance_name", "nodename"},
			Unit:   "%",
		},
		{
			Name:   "磁盘IO利用率",
			PQL:    "(rate(node_disk_io_time_seconds_total[1m])) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Labels: []string{"device", "instance", "instance_name", "job", "job_name", "nodename"},
			Unit:   "%",
		},
		{
			Name:   "网络RTT延时",
			PQL:    "kindling_network_rtt{} * 1000",
			Labels: []string{"container_id", "dst_ip", "instance", "job", "level", "namespace", "node", "node_ip", "node_name", "node_name", "pod", "src_ip", "src_node", "workload_kind", "workload_name"},
			Unit:   "ms",
		},
		{
			Name: "容器上次运行时间",
			PQL:  "container_last_seen",
			Labels: []string{
				"beta_kubernetes_io_arch",
				"beta_kubernetes_io_os",
				"container",
				"id",
				"image",
				"instance",
				"instance_name",
				"job",
				"job_name",
				"k8slens_edit_resource_version",
				"kubernetes_io_arch",
				"kubernetes_io_hostname",
				"kubernetes_io_os",
				"name",
				"namespace",
				"pod",
			},
			Unit: "s",
		},
		{
			Name:   "容器持久卷使用率",
			PQL:    "(sum(container_fs_inodes_free{name!=\"\"}) BY (instance_name) / sum(container_fs_inodes_total) BY (instance_name)) * 100",
			Labels: []string{"instance_name"},
			Unit:   "%",
		},
		{
			Name:   "容器CPU使用率",
			PQL:    "(sum(rate(container_cpu_usage_seconds_total{container!=\"\"}[5m])) by (pod, container) / sum(container_spec_cpu_quota{container!=\"\"}/container_spec_cpu_period{container!=\"\"}) by (pod, container) * 100)",
			Labels: []string{"pod", "instance_name"},
			Unit:   "%",
		},
		{
			Name:   "容器cpu_cfs_throttled",
			PQL:    "sum(increase(container_cpu_cfs_throttled_periods_total{container!=\"\"}[5m])) by (container, pod, namespace) / sum(increase(container_cpu_cfs_periods_total[5m])) by (container, pod, namespace)",
			Labels: []string{"container", "namespace", "pod"},
			Unit:   "%",
		},
		{
			Name:   "容器内存使用率",
			PQL:    "(sum(container_memory_working_set_bytes{name!=\"\"}) BY (instance_name, name) / sum(container_spec_memory_limit_bytes > 0) BY (instance_name, name) * 100)",
			Labels: []string{"instance_name", "name"},
			Unit:   "%",
		},
	},
}

func (s *service) GetMetricPQL() response.GetMetricPQLResponse {
	return metricData
}
