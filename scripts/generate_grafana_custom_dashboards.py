#!/usr/bin/env python3
"""Generate import-ready APO custom dashboard JSON files.

Each generated dashboard is a standalone JSON object in the same shape stored
in custom_dashboard.config. Coverage metadata and unsupported Grafana panels
are written separately to manifest.json.
"""

from __future__ import annotations

import argparse
import copy
import json
import re
from pathlib import Path
from typing import Any

import import_grafana_dashboard
import import_custom_dashboard_json


SOURCE_REVISION = "30de16b"
MAX_WIDGETS = 60

SOURCES = [
    ("cluster-overview", "zh/default/ClusterOverview.json", "集群总览", "cluster"),
    ("infrastructure", "zh/default/Infrastructure.json", "基础设施情况", "infrastructure"),
    ("node-exporter", "zh/default/Node-Exporter-Full.json", "宿主机监控指标", "infrastructure"),
    ("originx-data-monitor", "zh/default/Originx-Data-Monitor.json", "Originx 数据自监控", "observability"),
    ("originx-polaris-container", "zh/default/Originx-Polaris-Metrics-Container.json", "Originx 北极星指标（容器层级）", "application"),
    ("originx-polaris-service", "zh/default/Originx-Polaris-Metrics.json", "Originx 北极星指标（服务层级）", "application"),
    ("service-middleware", "zh/default/Service-MiddleWare-Metrics.json", "服务调用中间件", "middleware"),
    ("thread-polaris-container", "zh/default/Thread Polaris Metrics-zh.json", "线程级北极星指标（容器）", "application"),
    ("thread-polaris-vm", "zh/default/Thread-Polaris-Metrics-VM-zh.json", "线程级北极星指标（虚拟机）", "application"),
    ("elasticsearch", "zh/middleware/es.json", "Elasticsearch 指标", "middleware"),
    ("kafka", "zh/middleware/kafka.json", "Kafka 指标", "middleware"),
    ("mysql", "zh/middleware/mysql-temp-zh-cn.json", "MySQL 指标", "middleware"),
    ("rabbitmq-exporter", "zh/middleware/rabbitmq-exporter.json", "RabbitMQ Exporter 指标", "middleware"),
    ("rabbitmq-official", "zh/middleware/rabbitmq-official.json", "RabbitMQ 官方指标", "middleware"),
    ("redis", "zh/middleware/redis.json", "Redis 指标", "middleware"),
    ("otel-collector", "en/default/Originx-Otel-Collector.json", "OpenTelemetry Collector 指标", "observability"),
]

UNSUPPORTED_SOURCES = [
    {
        "path": "zh/default/Java-Agent-Monitor.json",
        "dashboard": "Java 探针监控",
        "panels": 4,
        "reason": "面板使用 ClickHouse SQL，当前我的面板仅支持 PromQL 查询",
    },
    {
        "path": "zh/middleware/nginx.json",
        "dashboard": "NGINX 请求日志分析",
        "panels": 65,
        "reason": "面板使用 ClickHouse SQL 和动态日志表，当前我的面板仅支持 PromQL 查询",
    },
    {
        "path": "en/default/Originx-Otel-Collector.json",
        "dashboard": "OpenTelemetry Collector 指标",
        "panels": 1,
        "reason": "Grafana 文本说明面板不属于当前我的面板组件类型",
    },
]

PANEL_TITLES = {
    "Pressure": "资源压力",
    "CPU Busy": "CPU 繁忙率",
    "Sys Load": "系统负载",
    "RAM Used": "内存使用率",
    "SWAP Used": "交换空间使用率",
    "Root FS Used": "根文件系统使用率",
    "CPU Cores": "CPU 核心数",
    "Uptime": "运行时间",
    "RootFS Total": "根文件系统总容量",
    "RAM Total": "内存总容量",
    "SWAP Total": "交换空间总容量",
    "CPU Basic": "CPU 基础指标",
    "Memory Basic": "内存基础指标",
    "Network Traffic Basic": "网络流量基础指标",
    "Disk Space Used Basic": "磁盘空间基础指标",
    "CPU": "CPU 指标",
    "Memory Stack": "内存构成",
    "Network Traffic": "网络流量",
    "Disk Space Used": "磁盘空间使用率",
    "Disk IOps": "磁盘 IOPS",
    "I/O Usage Read / Write": "磁盘读写吞吐",
    "I/O Utilization": "磁盘 I/O 利用率",
    "CPU spent seconds in guests (VMs)": "虚拟机占用 CPU 时间",
    "Memory Active / Inactive": "活跃与非活跃内存",
    "Memory Committed": "已提交内存",
    "Memory Active / Inactive Detail": "活跃与非活跃内存明细",
    "Memory Writeback and Dirty": "回写与脏页内存",
    "Memory Shared and Mapped": "共享与映射内存",
    "Memory Slab": "Slab 内存",
    "Memory Vmalloc": "虚拟内存分配",
    "Memory Bounce": "Bounce 内存",
    "Memory Anonymous": "匿名内存",
    "Memory Kernel / CPU": "内核与 CPU 内存",
    "Memory HugePages Counter": "大页数量",
    "Memory HugePages Size": "大页容量",
    "Memory DirectMap": "直接映射内存",
    "Memory Unevictable and MLocked": "不可回收与锁定内存",
    "Memory NFS": "NFS 内存",
    "Memory Pages In / Out": "内存页面换入换出",
    "Memory Pages Swap In / Out": "交换页面换入换出",
    "Memory Page Faults": "内存缺页",
    "OOM Killer": "OOM 终止进程",
    "Time Synchronized Drift": "时间同步偏差",
    "Time PLL Adjust": "时间锁相环调整",
    "Time Synchronized Status": "时间同步状态",
    "Time Misc": "其他时间指标",
    "Processes Status": "进程状态",
    "Processes State": "进程运行状态",
    "Processes  Forks": "进程创建次数",
    "Processes Memory": "进程内存",
    "PIDs Number and Limit": "PID 数量与上限",
    "Process schedule stats Running / Waiting": "进程调度运行与等待",
    "Threads Number and Limit": "线程数量与上限",
    "Context Switches / Interrupts": "上下文切换与中断",
    "System Load": "系统负载",
    "CPU Frequency Scaling": "CPU 频率调节",
    "Pressure Stall Information": "资源压力停顿信息",
    "Interrupts Detail": "中断明细",
    "Schedule timeslices executed by each cpu": "各 CPU 调度时间片",
    "Entropy": "系统熵",
    "CPU time spent in user and system contexts": "用户态与系统态 CPU 时间",
    "File Descriptors": "文件描述符",
    "Hardware temperature monitor": "硬件温度监控",
    "Throttle cooling device": "节流散热设备",
    "Power supply": "电源状态",
    "Systemd Sockets": "Systemd 套接字",
    "Systemd Units State": "Systemd 单元状态",
    "Disk IOps Completed": "磁盘已完成 IOPS",
    "Disk R/W Data": "磁盘读写数据量",
    "Disk Average Wait Time": "磁盘平均等待时间",
    "Average Queue Size": "平均队列长度",
    "Disk R/W Merged": "磁盘合并读写次数",
    "Time Spent Doing I/Os": "磁盘 I/O 耗时",
    "Instantaneous Queue Size": "瞬时队列长度",
    "Disk IOps Discards completed / merged": "磁盘丢弃完成与合并次数",
    "Filesystem space available": "文件系统可用空间",
    "File Nodes Free": "可用文件节点",
    "File Descriptor": "文件描述符数量",
    "File Nodes Size": "文件节点总量",
    "Filesystem in ReadOnly / Error": "只读或异常文件系统",
    "Network Traffic by Packets": "网络包流量",
    "Network Traffic Errors": "网络传输错误",
    "Network Traffic Drop": "网络丢包",
    "Network Traffic Compressed": "网络压缩包",
    "Network Traffic Multicast": "网络组播包",
    "Network Traffic Fifo": "网络 FIFO 错误",
    "Network Traffic Frame": "网络帧错误",
    "Network Traffic Carrier": "网络载波错误",
    "Network Traffic Colls": "网络冲突",
    "NF Conntrack": "连接跟踪",
    "ARP Entries": "ARP 条目",
    "MTU": "MTU 配置",
    "Speed": "网卡速率",
    "Queue Length": "发送队列长度",
    "Softnet Packets": "Softnet 数据包",
    "Softnet Out of Quota": "Softnet 超限",
    "Network Operational Status": "网络运行状态",
    "Sockstat TCP": "TCP 套接字统计",
    "Sockstat UDP": "UDP 套接字统计",
    "Sockstat FRAG / RAW": "FRAG 与 RAW 套接字统计",
    "Sockstat Memory Size": "套接字内存",
    "Sockstat Used": "已用套接字",
    "Netstat IP In / Out Octets": "IP 收发字节",
    "Netstat IP Forwarding": "IP 转发",
    "ICMP In / Out": "ICMP 收发",
    "ICMP Errors": "ICMP 错误",
    "UDP In / Out": "UDP 收发",
    "UDP Errors": "UDP 错误",
    "TCP In / Out": "TCP 收发",
    "TCP Errors": "TCP 错误",
    "TCP Connections": "TCP 连接",
    "TCP SynCookie": "TCP SYN Cookie 状态",
    "TCP SYN Cookie": "TCP SYN Cookie 状态",
    "TCP Direct Transition": "TCP 直接状态转换",
    "TCP Stat": "TCP 状态统计",
    "Node Exporter Scrape Time": "Node Exporter 采集耗时",
    "Node Exporter Scrape": "Node Exporter 采集状态",
    "Futex": "Futex 等待",
    "Epoll": "Epoll 等待",
    "File": "文件 I/O",
    "Other": "其他等待",
    "Runq": "运行队列",
    "Idle": "空闲",
    "On CPU": "CPU 执行时间",
    "On CPU P90": "CPU 执行时间 P90",
    "File P90": "文件 I/O P90",
    "Net": "网络 I/O",
    "Net P90": "网络 I/O P90",
    "Epoll P90": "Epoll 等待 P90",
    "Futex P90": "Futex 等待 P90",
    "Other P90": "其他等待 P90",
    "Runqueue Latency": "运行队列延迟",
    "Runq P90": "运行队列延迟 P90",
    "Brokers": "Broker 数量",
    "Topics": "Topic 数量",
    "Replicas": "副本数量",
    "In Sync Replicas": "同步副本数量",
    "Partitions not on preferred node": "不在首选节点上的分区",
    "RabbitMQ Server": "RabbitMQ 服务",
    "Node up Stats": "节点存活状态",
    "Exchanges": "交换机",
    "Channels": "通道",
    "Consumers": "消费者",
    "Connections": "连接",
    "Queues": "队列",
    "Messages/host": "每主机消息数",
    "Messages / Queue": "每队列消息数",
    "Memory": "内存",
    "FIle descriptors": "文件描述符",
    "Sockets": "套接字",
    "Ready messages": "就绪消息数",
    "Incoming messages / s": "每秒进入消息数",
    "Publishers": "发布者",
    "Unacknowledged messages": "未确认消息数",
    "Outgoing messages / s": "每秒发出消息数",
    "Nodes": "节点",
    "Memory available before publishers blocked": "发布者阻塞前可用内存",
    "Disk space available before publishers blocked": "发布者阻塞前可用磁盘空间",
    "File descriptors available": "可用文件描述符",
    "TCP sockets available": "可用 TCP 套接字",
    "Messages ready to be delivered to consumers": "待投递给消费者的消息",
    "Messages pending consumer acknowledgement": "等待消费者确认的消息",
    "Messages published / s": "每秒发布消息数",
    "Messages confirmed to publishers / s": "每秒向发布者确认消息数",
    "Messages routed to queues / s": "每秒路由到队列的消息数",
    "Messages unconfirmed to publishers / s": "每秒未向发布者确认的消息数",
    "Unroutable messages dropped / s": "每秒丢弃的不可路由消息数",
    "Unroutable messages returned to publishers / s": "每秒退回发布者的不可路由消息数",
    "Messages delivered / s": "每秒投递消息数",
    "Messages redelivered / s": "每秒重新投递消息数",
    "Messages delivered with manual ack / s": "每秒手动确认模式投递消息数",
    "Messages delivered auto ack / s": "每秒自动确认模式投递消息数",
    "Messages acknowledged / s": "每秒确认消息数",
    "Polling operations with auto ack / s": "每秒自动确认轮询次数",
    "Polling operations that yield no result / s": "每秒无结果轮询次数",
    "Polling operations with manual ack / s": "每秒手动确认轮询次数",
    "Total queues": "队列总数",
    "Queues declared / s": "每秒声明队列数",
    "Queues created / s": "每秒创建队列数",
    "Queues deleted / s": "每秒删除队列数",
    "Total channels": "通道总数",
    "Channels opened / s": "每秒打开通道数",
    "Channels closed / s": "每秒关闭通道数",
    "Total connections": "连接总数",
    "Connections opened / s": "每秒新建连接数",
    "Connections closed / s": "每秒关闭连接数",
    "Spans ${metric:text}": "Span 数量 ${metric:text}",
    "Metric Points ${metric:text}": "指标点 ${metric:text}",
    "Log Records ${metric:text}": "日志记录 ${metric:text}",
    "Batch Send Size Heatmap": "批量发送大小分布",
    "Batch Metrics 1": "批处理指标一",
    "Batch Metrics 2": "批处理指标二",
    "Exporter Queue Size": "导出器队列长度",
    "Exporter Queue Capacity": "导出器队列容量",
    "Exporter Queue Usage": "导出器队列使用率",
    "Total RSS Memory": "RSS 内存总量",
    "Total Runtime Sys Memory": "运行时系统内存总量",
    "Total Runtime Heap Memory": "运行时堆内存总量",
    "CPU Usage": "CPU 使用率",
    "Service Instance Count": "服务实例数量",
    "Uptime by Service Instance": "服务实例运行时间",
    "Service Instance Details": "服务实例明细",
    "Spans Flow": "Span 流量",
    "Metric Points Flow": "指标点流量",
    "Log Records Flow": "日志记录流量",
    "RPC server responses by GRPC status code (receivers)": "接收器 RPC 服务端响应（按 gRPC 状态码）",
    "RPC client responses by GRPC status code (exporters)": "导出器 RPC 客户端响应（按 gRPC 状态码）",
    "RPC server duration (receivers)": "接收器 RPC 服务端耗时",
    "RPC client duration (exporters)": "导出器 RPC 客户端耗时",
    "RPC server request size (receivers)": "接收器 RPC 服务端请求大小",
    "RPC client request size (exporters)": "导出器 RPC 客户端请求大小",
    "Namespaces": "命名空间数量",
    "Pods": "Pod 数量",
}

ROW_TITLES = {
    "Quick CPU / Mem / Disk": "CPU、内存与磁盘概览",
    "Basic CPU / Mem / Net / Disk": "CPU、内存、网络与磁盘基础指标",
    "CPU / Memory / Net / Disk": "CPU、内存、网络与磁盘",
    "Memory Meminfo": "内存详细信息",
    "Memory Vmstat": "内存虚拟统计",
    "System Timesync": "系统时间同步",
    "System Processes": "系统进程",
    "System Misc": "其他系统指标",
    "Hardware Misc": "其他硬件指标",
    "Systemd": "Systemd 指标",
    "Storage Disk": "磁盘存储",
    "Storage Filesystem": "文件系统存储",
    "Network Traffic": "网络流量",
    "Network Sockstat": "网络套接字统计",
    "Network Netstat": "网络协议统计",
    "Node Exporter": "Node Exporter 状态",
}

VARIABLE_LABELS = {
    "datasource": "数据源",
    "cluster": "集群",
    "cluster_id": "集群",
    "job": "采集任务",
    "job_name": "采集任务",
    "cadvisor_job_name": "容器采集任务",
    "instance": "实例",
    "node": "主机",
    "node_name": "节点",
    "nodeName": "节点",
    "host_name": "主机",
    "namespace": "命名空间",
    "pod": "Pod 实例",
    "service": "服务",
    "service_name": "服务",
    "endpoint": "请求端点",
    "endpoint_name": "请求端点",
    "container_id": "容器 ID",
    "tid": "线程",
    "pid": "进程 ID",
    "related_pid": "关联进程 ID",
    "diskdevices": "磁盘设备",
    "name": "名称",
    "topic": "主题",
    "consumergroup_hide": "消费组过滤",
    "rabbitmq_cluster": "RabbitMQ 集群",
    "iid": "实例 ID",
    "minstep": "最小步长",
    "metric": "基础指标",
    "receiver": "接收器",
    "processor": "处理器",
    "exporter": "导出器",
    "grouping": "附加分组",
    "suffix": "后缀",
    "adhoc": "临时过滤",
    "resolution": "精度",
    "interval": "采样间隔",
}

VARIABLE_OPTION_LABELS = {
    "auto": "自动",
    "Rate": "速率",
    "Count": "增量",
    "None (basic metrics)": "不分组（基础指标）",
    "By transport (detailed metrics)": "按传输类型分组（详细指标）",
    "By service instance id": "按服务实例 ID 分组",
}

LEGEND_REPLACEMENTS = [
    ("Received", "接收"), ("Transmitted", "发送"), ("Receive", "接收"), ("Transmit", "发送"),
    ("Inbound", "入站"), ("Outbound", "出站"), ("Read", "读取"), ("Write", "写入"),
    ("Used", "已使用"), ("Free", "可用"), ("Total", "总量"), ("Limit", "上限"),
    ("Available", "可用"), ("Idle", "空闲"), ("Busy", "繁忙"), ("Errors", "错误"),
    ("Error", "错误"), ("Requests", "请求"), ("Connections", "连接"), ("Messages", "消息"),
]


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    parser.add_argument("--source-root", type=Path, required=True)
    parser.add_argument(
        "--output-dir",
        type=Path,
        default=Path(__file__).resolve().parents[1] / "backend/static/custom-dashboards",
    )
    return parser.parse_args()


def localize_title(value: Any) -> str:
    text = str(value or "").strip()
    if not text:
        return "未命名指标"
    return PANEL_TITLES.get(text, ROW_TITLES.get(text, text))


def localize_legend(value: Any) -> str:
    text = str(value or "")
    if text in PANEL_TITLES:
        return PANEL_TITLES[text]
    for source, target in LEGEND_REPLACEMENTS:
        text = re.sub(rf"\b{re.escape(source)}\b", target, text, flags=re.IGNORECASE)
    return text


def custom_description(descriptions: dict[str, str], path: list[str]) -> str:
    prefix = " - ".join(path)
    values = []
    for key, value in descriptions.items():
        if key == prefix or key.startswith(prefix + " - "):
            if value and value not in values:
                values.append(value)
    return "；".join(values[:3])


def localize_dashboard(grafana: dict[str, Any], name: str, descriptions: dict[str, str]) -> dict[str, Any]:
    localized = copy.deepcopy(grafana)
    source_title = str(grafana.get("title") or name)
    localized["title"] = name

    def walk(source_panels: list[dict[str, Any]], target_panels: list[dict[str, Any]], rows: list[str]) -> None:
        for source_panel, target_panel in zip(source_panels or [], target_panels or []):
            source_panel_title = str(source_panel.get("title") or "")
            if source_panel.get("type") == "row":
                target_panel["title"] = localize_title(source_panel_title)
                walk(source_panel.get("panels") or [], target_panel.get("panels") or [], rows + [source_panel_title])
                continue
            target_panel["title"] = localize_title(source_panel_title)
            description = custom_description(descriptions, [source_title, *rows, source_panel_title])
            target_panel["description"] = description or f"{target_panel['title']}。"
            for target in target_panel.get("targets") or []:
                if isinstance(target, dict) and target.get("legendFormat"):
                    target["legendFormat"] = localize_legend(target["legendFormat"])

    walk(grafana.get("panels") or [], localized.get("panels") or [], [])
    for variable in (localized.get("templating") or {}).get("list") or []:
        variable_name = str(variable.get("name") or "")
        current_label = str(variable.get("label") or variable_name)
        variable["label"] = VARIABLE_LABELS.get(variable_name, VARIABLE_LABELS.get(current_label.lower(), current_label))
        for option in variable.get("options") or []:
            if isinstance(option, dict):
                option_text = str(option.get("text") or "")
                if option_text in VARIABLE_OPTION_LABELS:
                    option["text"] = VARIABLE_OPTION_LABELS[option_text]
    return localized


def compact_layout(widgets: list[dict[str, Any]]) -> None:
    if not widgets:
        return
    min_y = min(int((widget.get("layout") or {}).get("y") or 0) for widget in widgets)
    for widget in widgets:
        layout = widget.get("layout") or {}
        layout["y"] = max(0, int(layout.get("y") or 0) - min_y)
        widget["layout"] = layout


def dashboard_config(grafana: dict[str, Any], name: str, source_id: str) -> dict[str, Any]:
    dashboard = import_grafana_dashboard.convert_grafana_dashboard(
        grafana,
        name=name,
        description=f"从 Grafana 面板迁移并汉化，已转换为“我的面板”原生格式。来源：{source_id}。",
        visibility="everyone",
    )
    compact_layout(dashboard["widgets"])
    return import_custom_dashboard_json.normalize_dashboard_json(dashboard, visibility="everyone")


def split_node_dashboard(grafana: dict[str, Any]) -> list[tuple[str, str, dict[str, Any]]]:
    panels = grafana.get("panels") or []
    definitions = [
        ("node-exporter-overview", "宿主机监控指标 - 概览、CPU 与内存", panels[:35]),
        ("node-exporter-system", "宿主机监控指标 - 系统与硬件", panels[35:40]),
        ("node-exporter-storage-network", "宿主机监控指标 - 存储与网络", panels[40:]),
    ]
    result = []
    for source_id, name, part_panels in definitions:
        part = copy.deepcopy(grafana)
        part["title"] = name
        part["panels"] = copy.deepcopy(part_panels)
        result.append((source_id, name, part))
    return result


def apply_source_query_fixes(source_id: str, grafana: dict[str, Any]) -> None:
    def walk(panels: list[dict[str, Any]]):
        for panel in panels or []:
            yield panel
            yield from walk(panel.get("panels") or [])

    if source_id == "thread-polaris-vm":
        variables = (grafana.setdefault("templating", {})).setdefault("list", [])
        variables.append(
            {
                "name": "job",
                "label": "采集任务",
                "type": "query",
                "query": {
                    "query": 'label_values(namedprocess_namegroup_cpu_seconds_total{cluster_id=~"$cluster_id"},job_name)'
                },
                "current": {"text": "integrations/process", "value": "integrations/process"},
            }
        )
        replacements = {
            'rate(container_cpu_usage_seconds_total{name=~"$container_id.*", cluster_id=~"$cluster_id"}[2m])':
                'sum by (instance_name,pid) (rate(namedprocess_namegroup_cpu_seconds_total{job_name=~"$job",instance_name=~"$node",pid=~"$pid",cluster_id=~"$cluster_id"}[$__rate_interval]))',
            'container_memory_rss{name=~"$container_id.*", cluster_id=~"$cluster_id"}':
                'sum by (instance_name,pid) (namedprocess_namegroup_memory_bytes{job_name=~"$job",instance_name=~"$node",pid=~"$pid",cluster_id=~"$cluster_id",memtype="resident"})',
        }
        for panel in walk(grafana.get("panels") or []):
            for target in panel.get("targets") or []:
                expression = target.get("expr") if isinstance(target, dict) else None
                if expression in replacements:
                    target["expr"] = replacements[expression]
    if source_id == "mysql":
        for panel in walk(grafana.get("panels") or []):
            for target in panel.get("targets") or []:
                if isinstance(target, dict) and target.get("expr"):
                    target["expr"] = str(target["expr"]).replace("$host", "$instance")


def validate_dashboard(dashboard: dict[str, Any]) -> None:
    widgets = dashboard.get("widgets") or []
    if not widgets:
        raise ValueError(f"dashboard has no widgets: {dashboard.get('name')}")
    if len(widgets) > MAX_WIDGETS:
        raise ValueError(f"dashboard exceeds {MAX_WIDGETS} widgets: {dashboard.get('name')} ({len(widgets)})")
    variable_names = {str(variable.get("name") or "") for variable in dashboard.get("variables") or []}
    for variable in dashboard.get("variables") or []:
        if not re.search(r"[\u4e00-\u9fff]", str(variable.get("label") or "")):
            raise ValueError(f"variable label was not localized: {dashboard.get('name')} / {variable.get('name')}")
    used_variables: set[str] = set()
    token_pattern = re.compile(r"\$([A-Za-z_][A-Za-z0-9_]*)|\$\{([A-Za-z_][A-Za-z0-9_]*)(?::[^}]*)?\}")
    for widget in widgets:
        if not re.search(r"[\u4e00-\u9fff]", str(widget.get("title") or "")):
            raise ValueError(f"widget title was not localized: {dashboard.get('name')} / {widget.get('title')}")
        for query in (widget.get("query") or {}).get("queries") or []:
            expression = str(query.get("expression") or "")
            if not expression or "SELECT " in expression.upper():
                raise ValueError(f"unsupported query in {dashboard.get('name')} / {widget.get('title')}")
            for match in token_pattern.finditer(expression):
                used_variables.add(match.group(1) or match.group(2))
    unresolved = used_variables - variable_names - {"__rate_interval", "__interval"}
    if unresolved:
        raise ValueError(f"unresolved variables in {dashboard.get('name')}: {sorted(unresolved)}")


def main() -> None:
    args = parse_args()
    descriptions = json.loads((args.source_root / "script/custom-describe.json").read_text(encoding="utf-8"))
    dashboard_root = args.source_root / "dashboards"
    dashboards = []
    source_widget_count = 0
    query_count = 0
    args.output_dir.mkdir(parents=True, exist_ok=True)
    for source_id, relative_path, name, category in SOURCES:
        grafana = json.loads((dashboard_root / relative_path).read_text(encoding="utf-8"))
        apply_source_query_fixes(source_id, grafana)
        localized = localize_dashboard(grafana, name, descriptions)
        parts = split_node_dashboard(localized) if source_id == "node-exporter" else [(source_id, name, localized)]
        for part_id, part_name, part in parts:
            dashboard = dashboard_config(part, part_name, part_id)
            validate_dashboard(dashboard)
            widget_count = len(dashboard["widgets"])
            dashboard_query_count = sum(
                len((widget.get("query") or {}).get("queries") or [])
                for widget in dashboard["widgets"]
            )
            source_widget_count += widget_count
            query_count += dashboard_query_count
            filename = f"{part_id}.json"
            (args.output_dir / filename).write_text(
                json.dumps(dashboard, ensure_ascii=False, indent=2) + "\n",
                encoding="utf-8",
            )
            dashboards.append(
                {
                    "id": part_id,
                    "file": filename,
                    "name": part_name,
                    "category": category,
                    "source": relative_path,
                    "widgets": widget_count,
                    "queries": dashboard_query_count,
                }
            )

    manifest = {
        "version": 1,
        "sourceRepository": "apo-init-dashboards",
        "sourceRevision": SOURCE_REVISION,
        "sourceDashboardCount": len(SOURCES) + 2,
        "generatedDashboardCount": len(dashboards),
        "importedPanelCount": source_widget_count,
        "promQLQueryCount": query_count,
        "unsupportedPanelCount": sum(item["panels"] for item in UNSUPPORTED_SOURCES),
        "unsupportedSources": UNSUPPORTED_SOURCES,
        "dashboards": dashboards,
    }
    (args.output_dir / "manifest.json").write_text(
        json.dumps(manifest, ensure_ascii=False, indent=2) + "\n",
        encoding="utf-8",
    )
    print(
        json.dumps(
            {
                key: manifest[key]
                for key in (
                    "generatedDashboardCount",
                    "importedPanelCount",
                    "promQLQueryCount",
                    "unsupportedPanelCount",
                )
            },
            ensure_ascii=False,
        )
    )


if __name__ == "__main__":
    main()
