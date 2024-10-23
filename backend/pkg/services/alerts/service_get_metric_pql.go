package alerts

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

var metricData = response.GetMetricPQLResponse{
	AlertMetricsData: []model.AlertMetricsData{
		{
			Name:  "平均请求延时",
			PQL:   "sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_sum[1m]))/ sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_count[1m]))/1000000",
			Unit:  "s",
			Group: "app",
		},
		{
			Name:  "请求错误率",
			PQL:   "sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_count{is_error=\"true\"}[1m]))/ sum by (svc_name, content_key) (increase(kindling_span_trace_duration_nanoseconds_count[1m]))",
			Unit:  "%",
			Group: "app",
		},
		// 主机相关
		{
			Name:  "磁盘使用率",
			PQL:   "((node_filesystem_avail_bytes * 100) / node_filesystem_size_bytes and ON (instance_name, device, mountpoint) node_filesystem_readonly == 0) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Unit:  "%",
			Group: "infra",
		},
		{
			Name:  "网络入吞吐量",
			PQL:   "(sum by (instance_name) (rate(node_network_receive_bytes_total[2m])) / 1024 / 1024) *  on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Unit:  "MB",
			Group: "infra",
		},
		{
			Name:  "磁盘读速度",
			PQL:   "(sum by (instance_name) (rate(node_disk_read_bytes_total[2m])) / 1024 / 1024) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Unit:  "MB/s",
			Group: "infra",
		},
		{
			Name:  "磁盘写速度",
			PQL:   "(sum by (instance_name) (rate(node_disk_written_bytes_total[2m])) / 1024 / 1024 ) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Unit:  "MB/s",
			Group: "infra",
		},
		{
			Name:  "CPU负载",
			PQL:   "(sum by (instance_name) (avg by (mode, instance_name) (rate(node_cpu_seconds_total{mode!=\"idle\"}[2m])))) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Unit:  "core",
			Group: "infra",
		},
		{
			Name:  "CPU IO Wait",
			PQL:   "(avg by (instance_name) (rate(node_cpu_seconds_total{mode=\"iowait\"}[5m])) * 100) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Unit:  "%",
			Group: "infra",
		},
		{
			Name:  "磁盘IO利用率",
			PQL:   "(rate(node_disk_io_time_seconds_total[1m])) * on(instance_name) group_left (nodename) node_uname_info{nodename=~\".+\"}",
			Unit:  "%",
			Group: "infra",
		},
		{
			Name:  "网络RTT延时",
			PQL:   "kindling_network_rtt{} * 1000",
			Unit:  "ms",
			Group: "infra",
		},
		// 容器相关
		{
			Name:  "容器上次运行时间",
			PQL:   "container_last_seen",
			Unit:  "s",
			Group: "container",
		},
		{
			Name:  "容器持久卷使用率",
			PQL:   "(sum(container_fs_inodes_free{name!=\"\"}) BY (instance_name) / sum(container_fs_inodes_total) BY (instance_name)) * 100",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器CPU使用率",
			PQL:   "(sum(rate(container_cpu_usage_seconds_total{container!=\"\"}[5m])) by (pod, container) / sum(container_spec_cpu_quota{container!=\"\"}/container_spec_cpu_period{container!=\"\"}) by (pod, container) * 100)",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器cpu_cfs_throttled",
			PQL:   "sum(increase(container_cpu_cfs_throttled_periods_total{container!=\"\"}[5m])) by (container, pod, namespace) / sum(increase(container_cpu_cfs_periods_total[5m])) by (container, pod, namespace)",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器内存使用率",
			PQL:   "(sum(container_memory_working_set_bytes{name!=\"\"}) BY (instance_name, name) / sum(container_spec_memory_limit_bytes > 0) BY (instance_name, name) * 100)",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 CPU 利用率（system）",
			PQL:   "irate(container_cpu_system_seconds_total{image!=\"\", image!~\".*pause.*\"}[3m]) * 100",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 CPU 利用率（user）",
			PQL:   "irate(container_cpu_user_seconds_total{image!=\"\", image!~\".*pause.*\"}[3m]) * 100",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 CPU 利用率（整体，值不会大于 100）",
			PQL:   "sum( irate(container_cpu_usage_seconds_total{image!=\"\", image!~\".*pause.*\"}[3m]) ) by (pod,namespace,container,image) / sum( container_spec_cpu_quota/container_spec_cpu_period ) by (pod,namespace,container,image)",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 CPU 利用率（整体，值可能会大于 100）",
			PQL:   "irate(container_cpu_usage_seconds_total{image!=\"\", image!~\".*pause.*\"}[3m]) * 100",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 CPU 每秒有多少 period",
			PQL:   "irate(container_cpu_cfs_periods_total{}[3m])",
			Unit:  "period/s",
			Group: "container",
		},
		{
			Name:  "容器 CPU 每秒被 throttle 的 period 量",
			PQL:   "irate(container_cpu_cfs_throttled_periods_total{}[3m])",
			Unit:  "period/s",
			Group: "container",
		},
		{
			Name:  "容器 CPU 被 throttle 的比例",
			PQL:   "irate(container_cpu_cfs_throttled_periods_total{}[3m]) / irate(container_cpu_cfs_periods_total{}[3m]) * 100",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 filesystem 使用率",
			PQL:   "container_fs_usage_bytes / container_fs_limit_bytes * 100",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 filesystem 使用量",
			PQL:   "container_fs_usage_bytes",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 filesystem 当前 IO 次数",
			PQL:   "container_fs_io_current",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 filesystem 当前 IO 次数",
			PQL:   "container_fs_io_current",
			Unit:  "number",
			Group: "container",
		},
		{
			Name:  "容器 filesystem 总量",
			PQL:   "container_fs_limit_bytes",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 inode free 量",
			PQL:   "container_fs_inodes_free",
			Unit:  "SI short",
			Group: "container",
		},
		{
			Name:  "容器 inode total 量",
			PQL:   "container_fs_inodes_total",
			Unit:  "SI short使用 SI 标准换算, 比如 1K=1000",
			Group: "container",
		},
		{
			Name:  "容器 inode 使用率",
			PQL:   "100 - container_fs_inodes_free / container_fs_inodes_total * 100",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 IO 每秒写入 byte 量",
			PQL:   "sum(irate(container_fs_writes_bytes_total[3m])) by (namespace, pod)",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 IO 每秒读取 byte 量",
			PQL:   "sum(irate(container_fs_reads_bytes_total[3m])) by (namespace, pod)",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 memory cache 量",
			PQL:   "container_memory_cache{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 memory 使用率（Usage）",
			PQL:   "100 * container_memory_usage_bytes/container_spec_memory_limit_bytes and container_spec_memory_limit_bytes != 0",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 memory 使用率（Working Set）",
			PQL:   "100 * container_memory_working_set_bytes/container_spec_memory_limit_bytes and container_spec_memory_limit_bytes != 0",
			Unit:  "%",
			Group: "container",
		},
		{
			Name:  "容器 memory 使用量（mapped_file）",
			PQL:   "container_memory_mapped_file{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 memory 使用量（RSS）",
			PQL:   "container_memory_rss{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 memory 使用量（RSS）",
			PQL:   "",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 memory 使用量（Swap）",
			PQL:   "container_memory_swap{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 memory 使用量（Usage）",
			PQL:   "container_memory_usage_bytes{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 memory 使用量（Working Set）",
			PQL:   "container_memory_working_set_bytes{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 memory 分配失败次数（每秒）",
			PQL:   "rate(container_memory_failures_total{}[3m])",
			Unit:  "number",
			Group: "container",
		},
		{
			Name:  "容器 memory 限制量",
			PQL:   "container_spec_memory_limit_bytes{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器 net 每秒发送 bit 量",
			PQL:   "sum(irate(container_network_transmit_bytes_total[3m])) by (namespace, pod) * 8",
			Unit:  "bit/s",
			Group: "container",
		},
		{
			Name:  "容器 net 每秒发送数据包数量",
			PQL:   "irate(container_network_transmit_packets_total[3m])",
			Unit:  "SI short使用 SI 标准换算, 比如 1K=1000",
			Group: "container",
		},
		{
			Name:  "容器 net 每秒发送时 drop 包数量",
			PQL:   "irate(container_network_transmit_packets_dropped_total[3m])",
			Unit:  "SI short使用 SI 标准换算, 比如 1K=1000",
			Group: "container",
		},
		{
			Name:  "容器 net 每秒发送错包数",
			PQL:   "irate(container_network_transmit_errors_total[3m])",
			Unit:  "SI short使用 SI 标准换算, 比如 1K=1000",
			Group: "container",
		},
		{
			Name:  "容器 net 每秒接收 bit 量",
			PQL:   "sum(irate(container_network_receive_bytes_total[3m])) by (namespace, pod) * 8",
			Unit:  "bit/s",
			Group: "container",
		},
		{
			Name:  "容器 net 每秒接收数据包数量",
			PQL:   "irate(container_network_receive_packets_total[3m])",
			Unit:  "SI short使用 SI 标准换算, 比如 1K=1000",
			Group: "container",
		},
		{
			Name:  "容器 net 每秒接收时 drop 包数量",
			PQL:   "irate(container_network_receive_packets_dropped_total[3m])",
			Unit:  "SI short使用 SI 标准换算, 比如 1K=1000\n",
			Group: "container",
		},
		{
			Name:  "容器 net 每秒接收错包数",
			PQL:   "irate(container_network_receive_errors_total[3m])",
			Unit:  "SI short使用 SI 标准换算, 比如 1K=1000",
			Group: "container",
		},
		{
			Name:  "容器允许运行的最大线程数",
			PQL:   "container_threads_max{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "number",
			Group: "container",
		},
		{
			Name:  "容器内 1 号进程 soft ulimit 值",
			PQL:   "container_ulimits_soft{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "number",
			Group: "container",
		},
		{
			Name:  "容器已经运行的时间",
			PQL:   "container_start_time_seconds{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "s",
			Group: "container",
		},
		{
			Name:  "容器当前打开套接字数量",
			PQL:   "container_sockets{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "number",
			Group: "container",
		},
		{
			Name:  "容器当前打开文件句柄数量",
			PQL:   "container_file_descriptors{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "number",
			Group: "container",
		},
		{
			Name:  "容器当前运行的线程数",
			PQL:   "container_threads{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "number",
			Group: "container",
		},
		{
			Name:  "容器当前运行的进程数",
			PQL:   "container_processes{image!=\"\", image!~\".*pause.*\"}",
			Unit:  "number",
			Group: "container",
		},
		{
			Name:  "容器正在使用的 GPU 加速卡内存量",
			PQL:   "container_accelerator_memory_used_bytes",
			Unit:  "byte",
			Group: "container",
		},
		{
			Name:  "容器总 GPU 加速卡可用内存量",
			PQL:   "container_accelerator_memory_total_bytes",
			Unit:  "byte",
			Group: "container",
		},
	},
}

func (s *service) GetMetricPQL() response.GetMetricPQLResponse {
	return metricData
}
