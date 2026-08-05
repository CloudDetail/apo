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
MAX_WIDGETS = 120
LEGACY_NODE_EXPORTER_OUTPUTS = (
    "node-exporter-overview.json",
    "node-exporter-system.json",
    "node-exporter-storage-network.json",
)

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
    "Memory Slab": "内核缓存（Slab）内存",
    "Memory Vmalloc": "虚拟内存分配",
    "Memory Bounce": "块设备回弹缓冲区内存",
    "Memory Anonymous": "匿名内存",
    "Memory Kernel / CPU": "内核与 CPU 内存",
    "Memory HugePages Counter": "大页数量",
    "Memory HugePages Size": "大页容量",
    "Memory DirectMap": "直接映射内存",
    "Memory Unevictable and MLocked": "不可回收与锁定内存",
    "Memory NFS": "网络文件系统（NFS）内存",
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
    "Softnet Packets": "网络软中断数据包",
    "Softnet Out of Quota": "网络软中断处理超限",
    "Network Operational Status": "网络运行状态",
    "Sockstat TCP": "TCP 套接字统计",
    "Sockstat UDP": "UDP 套接字统计",
    "Sockstat FRAG / RAW": "IP 分片与原始套接字统计",
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
    "Node Exporter Scrape Time": "节点指标采集耗时",
    "Node Exporter Scrape": "节点指标采集状态",
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

NODE_EXPORTER_LEGEND_TRANSLATIONS = {
    "CPU": "CPU 指标",
    "Mem": "内存指标",
    "I/O": "I/O 指标",
    "Busy System": "系统态繁忙",
    "Busy User": "用户态繁忙",
    "Busy Iowait": "I/O 等待",
    "Busy IRQs": "中断处理繁忙",
    "Busy Other": "其他繁忙",
    "Idle": "空闲",
    "RAM Total": "内存总容量",
    "RAM Used": "内存使用率",
    "RAM Cache + Buffer": "缓存与缓冲区",
    "RAM Free": "可用内存",
    "SWAP Used": "交换空间使用率",
    "recv {{device}}": "接收 {{device}}",
    "trans {{device}} ": "发送 {{device}}",
    "System - Processes executing in kernel mode": "系统态 - 内核模式进程",
    "User - Normal processes executing in user mode": "用户态 - 普通用户进程",
    "Nice - Niced processes executing in user mode": "低优先级 - 调整过优先级的用户进程",
    "Iowait - Waiting for I/O to complete": "I/O 等待 - 等待输入输出完成",
    "Irq - Servicing interrupts": "硬中断 - 处理硬件中断",
    "Softirq - Servicing softirqs": "软中断 - 处理软件中断",
    "Steal - Time spent in other operating systems when running in a virtualized environment": "虚拟化窃取 - 被其他虚拟机占用的 CPU 时间",
    "Idle - Waiting for something to happen": "空闲 - 等待任务",
    "Apps - Memory used by user-space applications": "应用程序 - 用户空间应用占用内存",
    "PageTables - Memory used to map between virtual and physical memory addresses": "页表 - 虚拟地址与物理地址映射内存",
    "SwapCache - Memory that keeps track of pages that have been fetched from swap but not yet been modified": "交换缓存 - 已从交换空间读回且尚未修改的页面",
    "Slab - Memory used by the kernel to cache data structures for its own use (caches like inode, dentry, etc)": "内核缓存 - 内核数据结构缓存（如 inode、dentry）",
    "Cache - Parked file data (file content) cache": "文件缓存 - 缓存的文件内容",
    "Buffers - Block device (e.g. harddisk) cache": "块设备缓冲区 - 磁盘等块设备缓存",
    "Unused - Free memory unassigned": "未使用 - 尚未分配的可用内存",
    "Swap - Swap space used": "交换空间 - 已使用容量",
    "Hardware Corrupted - Amount of RAM that the kernel identified as corrupted / not working": "硬件损坏 - 内核识别为损坏或不可用的内存",
    "{{device}} - Receive": "{{device}} - 接收",
    "{{device}} - Transmit": "{{device}} - 发送",
    "{{device}} - Reads completed": "{{device}} - 已完成读取",
    "{{device}} - Writes completed": "{{device}} - 已完成写入",
    "{{device}} - Successfully read bytes": "{{device}} - 成功读取字节数",
    "{{device}} - Successfully written bytes": "{{device}} - 成功写入字节数",
    "Guest - Time spent running a virtual CPU for a guest operating system": "虚拟机 - 客户操作系统虚拟 CPU 运行时间",
    "GuestNice - Time spent running a niced guest  (virtual CPU for guest operating system)": "低优先级虚拟机 - 调整过优先级的虚拟 CPU 运行时间",
    "Inactive - Memory which has been less recently used.  It is more eligible to be reclaimed for other purposes": "非活跃内存 - 最近较少使用且更容易被回收",
    "Active - Memory that has been used more recently and usually not reclaimed unless absolutely necessary": "活跃内存 - 最近使用且通常不会被回收",
    "Committed_AS - Amount of memory presently allocated on the system": "已提交内存 - 系统当前承诺分配的内存",
    "CommitLimit - Amount of  memory currently available to be allocated on the system": "提交上限 - 系统当前可承诺分配的内存上限",
    "Inactive_file - File-backed memory on inactive LRU list": "非活跃文件页 - 非活跃 LRU 列表中的文件页内存",
    "Inactive_anon - Anonymous and swap cache on inactive LRU list, including tmpfs (shmem)": "非活跃匿名页 - 非活跃 LRU 中的匿名页与交换缓存（含 tmpfs）",
    "Active_file - File-backed memory on active LRU list": "活跃文件页 - 活跃 LRU 列表中的文件页内存",
    "Active_anon - Anonymous and swap cache on active least-recently-used (LRU) list, including tmpfs": "活跃匿名页 - 活跃 LRU 中的匿名页与交换缓存（含 tmpfs）",
    "Writeback - Memory which is actively being written back to disk": "正在回写 - 正在写回磁盘的内存",
    "WritebackTmp - Memory used by FUSE for temporary writeback buffers": "临时回写 - FUSE 临时回写缓冲区",
    "Dirty - Memory which is waiting to get written back to the disk": "脏页 - 等待写回磁盘的内存",
    "Mapped - Used memory in mapped pages files which have been mapped, such as libraries": "映射内存 - 动态库等文件映射页占用内存",
    "Shmem - Used shared memory (shared between several processes, thus including RAM disks)": "共享内存 - 多进程共享内存（含内存盘）",
    "ShmemHugePages - Memory used by shared memory (shmem) and tmpfs allocated  with huge pages": "共享大页 - 共享内存与 tmpfs 使用的大页内存",
    "ShmemPmdMapped - Amount of shared (shmem/tmpfs) memory backed by huge pages": "共享大页映射 - 由大页支持的共享内存与 tmpfs",
    "SUnreclaim - Part of Slab, that cannot be reclaimed on memory pressure": "不可回收内核缓存 - 内存压力下无法回收的部分",
    "SReclaimable - Part of Slab, that might be reclaimed, such as caches": "可回收内核缓存 - 缓存等可回收部分",
    "VmallocChunk - Largest contiguous block of vmalloc area which is free": "虚拟连续内存最大空闲块",
    "VmallocTotal - Total size of vmalloc memory area": "虚拟连续内存总容量",
    "VmallocUsed - Amount of vmalloc area which is used": "虚拟连续内存已使用容量",
    "Bounce - Memory used for block device bounce buffers": "回弹缓冲区 - 块设备回弹缓冲区内存",
    "AnonHugePages - Memory in anonymous huge pages": "匿名大页内存",
    "AnonPages - Memory in user pages not backed by files": "匿名页 - 未由文件支持的用户页内存",
    "KernelStack - Kernel memory stack. This is not reclaimable": "内核栈 - 不可回收的内核栈内存",
    "PerCPU - Per CPU memory allocated dynamically by loadable modules": "每 CPU 内存 - 内核模块动态分配的 CPU 独占内存",
    "HugePages_Free - Huge pages in the pool that are not yet allocated": "空闲大页 - 大页池中尚未分配的大页",
    "HugePages_Rsvd - Huge pages for which a commitment to allocate from the pool has been made, but no allocation has yet been made": "预留大页 - 已承诺但尚未实际分配的大页",
    "HugePages_Surp - Huge pages in the pool above the value in /proc/sys/vm/nr_hugepages": "超额大页 - 超出系统配置数量的大页",
    "HugePages - Total size of the pool of huge pages": "大页池总容量",
    "Hugepagesize - Huge Page size": "单个大页容量",
    "DirectMap1G - Amount of pages mapped as this size": "1 GiB 直接映射内存",
    "DirectMap2M - Amount of pages mapped as this size": "2 MiB 直接映射内存",
    "DirectMap4K - Amount of pages mapped as this size": "4 KiB 直接映射内存",
    "Unevictable - Amount of unevictable memory that can't be swapped out for a variety of reasons": "不可回收内存 - 无法换出的内存",
    "MLocked - Size of pages locked to memory using the mlock() system call": "锁定内存 - 通过 mlock() 锁定的页面",
    "NFS Unstable - Memory in NFS pages sent to the server, but not yet committed to the storage": "NFS 未稳定页 - 已发送但尚未持久化的网络文件系统页面",
    "Pagesin - Page in operations": "页面换入次数",
    "Pagesout - Page out operations": "页面换出次数",
    "Pswpin - Pages swapped in": "交换页换入次数",
    "Pswpout - Pages swapped out": "交换页换出次数",
    "Pgfault - Page major and minor fault operations": "缺页次数 - 包含主缺页与次缺页",
    "Pgmajfault - Major page fault operations": "主缺页次数",
    "Pgminfault - Minor page fault operations": "次缺页次数",
    "oom killer invocations ": "内存不足终止器调用次数",
    "Estimated error in seconds": "预计时间误差（秒）",
    "Time offset in between local system and reference clock": "本地时间与参考时钟偏差",
    "Maximum error in seconds": "最大时间误差（秒）",
    "Phase-locked loop time adjust": "锁相环时间调整",
    "Is clock synchronized to a reliable server (1 = yes, 0 = no)": "时钟是否已与可靠服务器同步（1=是，0=否）",
    "Local clock frequency adjustment": "本地时钟频率调整",
    "Seconds between clock ticks": "时钟滴答间隔（秒）",
    "International Atomic Time (TAI) offset": "国际原子时（TAI）偏移",
    "Processes blocked waiting for I/O to complete": "等待 I/O 完成的阻塞进程数",
    "Processes in runnable state": "可运行状态进程数",
    "Processes forks second": "每秒创建进程数",
    "Processes virtual memory size in bytes": "进程虚拟内存大小（字节）",
    "Maximum amount of virtual memory available in bytes": "虚拟内存最大可用容量（字节）",
    "Number of PIDs": "PID 数量",
    "PIDs limit": "PID 数量上限",
    "CPU {{ cpu }} - seconds spent running a process": "CPU {{ cpu }} - 进程运行耗时（秒）",
    "CPU {{ cpu }} - seconds spent by processing waiting for this CPU": "CPU {{ cpu }} - 进程等待调度耗时（秒）",
    "Allocated threads": "已分配线程数",
    "Threads limit": "线程数量上限",
    "Context switches": "上下文切换次数",
    "Interrupts": "中断次数",
    "Load 1m": "1 分钟负载",
    "Load 5m": "5 分钟负载",
    "Load 15m": "15 分钟负载",
    "Max": "最大值",
    "Min": "最小值",
    "CPU some": "CPU 部分任务等待",
    "Memory some": "内存部分任务等待",
    "Memory full": "内存全部任务停顿",
    "I/O some": "I/O 部分任务等待",
    "I/O full": "I/O 全部任务停顿",
    "Entropy available to random number generators": "随机数生成器可用熵",
    "Time spent": "耗时",
    "Maximum open file descriptors": "可打开文件描述符上限",
    "Open file descriptors": "已打开文件描述符数",
    "{{ chip_name }} {{ sensor }} temp": "{{ chip_name }} {{ sensor }} 温度",
    "{{ chip_name }} {{ sensor }} Critical Alarm": "{{ chip_name }} {{ sensor }} 严重告警阈值",
    "{{ chip_name }} {{ sensor }} Critical": "{{ chip_name }} {{ sensor }} 严重阈值",
    "{{ chip_name }} {{ sensor }} Critical Historical": "{{ chip_name }} {{ sensor }} 历史严重阈值",
    "{{ chip_name }} {{ sensor }} Max": "{{ chip_name }} {{ sensor }} 最大温度",
    "Current {{ name }} in {{ type }}": "{{ type }} 中 {{ name }} 的当前值",
    "Max {{ name }} in {{ type }}": "{{ type }} 中 {{ name }} 的最大值",
    "{{ power_supply }} online": "{{ power_supply }} 在线状态",
    "{{ name }} Connections": "{{ name }} 连接数",
    "Activating": "正在激活",
    "Active": "活跃",
    "Deactivating": "正在停用",
    "Failed": "失败",
    "Inactive": "未激活",
    "{{device}} - Read bytes": "{{device}} - 读取字节数",
    "{{device}} - Written bytes": "{{device}} - 写入字节数",
    "{{device}} - Read wait time avg": "{{device}} - 平均读取等待时间",
    "{{device}} - Write wait time avg": "{{device}} - 平均写入等待时间",
    "{{device}} - Read merged": "{{device}} - 合并读取次数",
    "{{device}} - Write merged": "{{device}} - 合并写入次数",
    "{{device}} - IO": "{{device}} - I/O 耗时",
    "{{device}} - discard": "{{device}} - 丢弃耗时",
    "{{device}} - IO now": "{{device}} - 当前 I/O 数量",
    "{{device}} - Discards completed": "{{device}} - 已完成丢弃次数",
    "{{device}} - Discards merged": "{{device}} - 合并丢弃次数",
    "{{mountpoint}} - Available": "{{mountpoint}} - 可用容量",
    "{{mountpoint}} - Free": "{{mountpoint}} - 空闲容量",
    "{{mountpoint}} - Size": "{{mountpoint}} - 总容量",
    "{{mountpoint}} - Free file nodes": "{{mountpoint}} - 可用文件节点数",
    "Max open files": "可打开文件数上限",
    "Open files": "已打开文件数",
    "{{mountpoint}} - File nodes total": "{{mountpoint}} - 文件节点总数",
    "{{mountpoint}} - ReadOnly": "{{mountpoint}} - 只读状态",
    "{{mountpoint}} - Device error": "{{mountpoint}} - 设备错误状态",
    "{{device}} - Receive errors": "{{device}} - 接收错误",
    "{{device}} - Transmit errors": "{{device}} - 发送错误",
    "{{device}} - Receive drop": "{{device}} - 接收丢包",
    "{{device}} - Transmit drop": "{{device}} - 发送丢包",
    "{{device}} - Receive compressed": "{{device}} - 接收压缩包",
    "{{device}} - Transmit compressed": "{{device}} - 发送压缩包",
    "{{device}} - Receive multicast": "{{device}} - 接收组播包",
    "{{device}} - Receive fifo": "{{device}} - 接收 FIFO 错误",
    "{{device}} - Transmit fifo": "{{device}} - 发送 FIFO 错误",
    "{{device}} - Receive frame": "{{device}} - 接收帧错误",
    "{{device}} - Statistic transmit_carrier": "{{device}} - 发送载波错误",
    "{{device}} - Transmit colls": "{{device}} - 发送冲突",
    "NF conntrack entries": "连接跟踪条目数",
    "NF conntrack limit": "连接跟踪条目上限",
    "{{ device }} - ARP entries": "{{ device }} - ARP 条目数",
    "{{ device }} - Bytes": "{{ device }} - MTU 字节数",
    "{{ device }} - Speed": "{{ device }} - 网卡速率",
    "{{ device }} -   Interface transmit queue length": "{{ device }} - 接口发送队列长度",
    "CPU {{cpu}} - Processed": "CPU {{cpu}} - 已处理数据包",
    "CPU {{cpu}} - Dropped": "CPU {{cpu}} - 已丢弃数据包",
    "CPU {{cpu}} - Squeezed": "CPU {{cpu}} - 处理超限次数",
    "{{interface}} - Operational state UP": "{{interface}} - 运行状态正常",
    "{{device}} - Physical link state": "{{device}} - 物理链路状态",
    "TCP_alloc - Allocated sockets": "TCP 已分配套接字数",
    "TCP_inuse - Tcp sockets currently in use": "TCP 正在使用的套接字数",
    "TCP_mem - Used memory for tcp": "TCP 套接字占用内存页数",
    "TCP_orphan - Orphan sockets": "TCP 孤儿套接字数",
    "TCP_tw - Sockets waiting close": "TCP 等待关闭套接字数",
    "UDPLITE_inuse - Udplite sockets currently in use": "UDPLite 正在使用的套接字数",
    "UDP_inuse - Udp sockets currently in use": "UDP 正在使用的套接字数",
    "UDP_mem - Used memory for udp": "UDP 套接字占用内存页数",
    "FRAG_inuse - Frag sockets currently in use": "IP 分片正在使用的套接字数",
    "RAW_inuse - Raw sockets currently in use": "原始套接字正在使用的数量",
    "mem_bytes - TCP sockets in that state": "TCP 套接字占用内存字节数",
    "mem_bytes - UDP sockets in that state": "UDP 套接字占用内存字节数",
    "FRAG_memory - Used memory for frag": "IP 分片占用内存",
    "Sockets_used - Sockets currently in use": "当前正在使用的套接字总数",
    "InOctets - Received octets": "接收字节数",
    "OutOctets - Sent octets": "发送字节数",
    "Forwarding - IP forwarding": "IP 转发状态",
    "InMsgs -  Messages which the entity received. Note that this counter includes all those counted by icmpInErrors": "接收消息数 - 包含 ICMP 错误消息",
    "OutMsgs - Messages which this entity attempted to send. Note that this counter includes all those counted by icmpOutErrors": "发送消息数 - 包含 ICMP 发送错误消息",
    "InErrors - Messages which the entity received but determined as having ICMP-specific errors (bad ICMP checksums, bad length, etc.)": "接收错误 - ICMP 校验和错误、长度错误等",
    "InDatagrams - Datagrams received": "接收数据报数",
    "OutDatagrams - Datagrams sent": "发送数据报数",
    "InErrors - UDP Datagrams that could not be delivered to an application": "接收错误 - 无法投递给应用的 UDP 数据报",
    "NoPorts - UDP Datagrams received on a port with no listener": "无监听端口 - 接收到但端口无监听者的 UDP 数据报",
    "InErrors Lite - UDPLite Datagrams that could not be delivered to an application": "UDPLite 接收错误 - 无法投递给应用的数据报",
    "RcvbufErrors - UDP buffer errors received": "接收缓冲区错误 - UDP 接收缓冲区不足",
    "SndbufErrors - UDP buffer errors send": "发送缓冲区错误 - UDP 发送缓冲区不足",
    "InSegs - Segments received, including those received in error. This count includes segments received on currently established connections": "接收报文段数 - 包含错误报文段和已建立连接报文段",
    "OutSegs - Segments sent, including those on current connections but excluding those containing only retransmitted octets": "发送报文段数 - 不含仅包含重传字节的报文段",
    "ListenOverflows - Times the listen queue of a socket overflowed": "监听队列溢出次数",
    "ListenDrops - SYNs to LISTEN sockets ignored": "监听丢弃次数 - 被忽略的 SYN 请求",
    "TCPSynRetrans - SYN-SYN/ACK retransmits to break down retransmissions in SYN, fast/timeout retransmits": "SYN 重传次数 - 包含 SYN 与 SYN/ACK 重传",
    "RetransSegs - Segments retransmitted - that is, the number of TCP segments transmitted containing one or more previously transmitted octets": "重传报文段数 - 包含已发送字节的 TCP 重传报文段",
    "InErrs - Segments received in error (e.g., bad TCP checksums)": "接收错误报文段数 - 如 TCP 校验和错误",
    "OutRsts - Segments sent with RST flag": "发送 RST 报文段数",
    "TCPRcvQDrop - Packets meant to be queued in rcv queue but dropped because socket rcvbuf limit hit": "接收队列丢包数 - 套接字接收缓冲区达到上限",
    "TCPOFOQueue - TCP layer receives an out of order packet and has enough memory to queue it": "乱序队列次数 - TCP 接收乱序包并成功入队",
    "CurrEstab - TCP connections for which the current state is either ESTABLISHED or CLOSE- WAIT": "当前已建立连接数 - ESTABLISHED 或 CLOSE_WAIT 状态",
    "MaxConn - Limit on the total number of TCP connections the entity can support (Dynamic is \"-1\")": "连接数上限 - 动态上限显示为 -1",
    "SyncookiesFailed - Invalid SYN cookies received": "无效 SYN Cookie 接收次数",
    "SyncookiesRecv - SYN cookies received": "SYN Cookie 接收次数",
    "SyncookiesSent - SYN cookies sent": "SYN Cookie 发送次数",
    "ActiveOpens - TCP connections that have made a direct transition to the SYN-SENT state from the CLOSED state": "主动打开连接数 - 从 CLOSED 转为 SYN_SENT",
    "PassiveOpens - TCP connections that have made a direct transition to the SYN-RCVD state from the LISTEN state": "被动打开连接数 - 从 LISTEN 转为 SYN_RECV",
    "established - TCP sockets in established state": "已建立连接状态",
    "fin_wait2 - TCP sockets in fin_wait2 state": "等待远端关闭状态",
    "listen - TCP sockets in listen state": "监听状态",
    "time_wait - TCP sockets in time_wait state": "等待超时关闭状态",
    "{{collector}} - Scrape duration": "{{collector}} - 采集耗时",
    "{{collector}} - Scrape success": "{{collector}} - 采集成功状态",
    "{{collector}} - Scrape textfile error (1 = true)": "{{collector}} - 文本文件采集错误（1=错误）",
}


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


def localize_legend(value: Any, source_id: str = "") -> str:
    text = str(value or "")
    if source_id == "node-exporter" and text in NODE_EXPORTER_LEGEND_TRANSLATIONS:
        return NODE_EXPORTER_LEGEND_TRANSLATIONS[text]
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


def localize_description(value: str, source_id: str = "") -> str:
    if source_id != "node-exporter":
        return value
    return (
        value.replace("Niced进程", "调整过优先级的进程")
        .replace("Node Exporter", "节点指标采集器")
    )


def localize_dashboard(
    grafana: dict[str, Any],
    name: str,
    descriptions: dict[str, str],
    source_id: str = "",
) -> dict[str, Any]:
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
            target_panel["description"] = localize_description(
                description or f"{target_panel['title']}。",
                source_id,
            )
            for target in target_panel.get("targets") or []:
                if isinstance(target, dict) and target.get("legendFormat"):
                    target["legendFormat"] = localize_legend(target["legendFormat"], source_id)

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


def remove_legacy_node_exporter_outputs(output_dir: Path) -> None:
    for filename in LEGACY_NODE_EXPORTER_OUTPUTS:
        path = output_dir / filename
        if path.is_file():
            path.unlink()


def main() -> None:
    args = parse_args()
    descriptions = json.loads((args.source_root / "script/custom-describe.json").read_text(encoding="utf-8"))
    dashboard_root = args.source_root / "dashboards"
    dashboards = []
    source_widget_count = 0
    query_count = 0
    args.output_dir.mkdir(parents=True, exist_ok=True)
    remove_legacy_node_exporter_outputs(args.output_dir)
    for source_id, relative_path, name, category in SOURCES:
        grafana = json.loads((dashboard_root / relative_path).read_text(encoding="utf-8"))
        apply_source_query_fixes(source_id, grafana)
        localized = localize_dashboard(grafana, name, descriptions, source_id)
        dashboard = dashboard_config(localized, name, source_id)
        validate_dashboard(dashboard)
        widget_count = len(dashboard["widgets"])
        dashboard_query_count = sum(
            len((widget.get("query") or {}).get("queries") or [])
            for widget in dashboard["widgets"]
        )
        source_widget_count += widget_count
        query_count += dashboard_query_count
        filename = f"{source_id}.json"
        (args.output_dir / filename).write_text(
            json.dumps(dashboard, ensure_ascii=False, indent=2) + "\n",
            encoding="utf-8",
        )
        dashboards.append(
            {
                "id": source_id,
                "file": filename,
                "name": name,
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
