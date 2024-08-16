# OneAgent
OneAgent 用于收集被监控环境中的各类可观测性数据，包括链路数据、日志数据和指标数据等。OneAgent 能够部署在各类环境中，包括传统服务器、容器、Kubernetes环境中。

**OneAgent** 集成了以下内容：

- 基于 Linux Preload 机制和 Kubernetes Webhook 机制自动安装 [OpenTelemetry](https://github.com/open-telemetry) 探针，应用重启生效
- 基于 eBPF 技术，采集北极星因果指标，并实现链路数据的[回溯采样算法](https://www.usenix.org/conference/nsdi23/presentation/zhang-lei)
- 通过 [ilogtail](https://github.com/alibaba/ilogtail) 采集故障现场日志，依据回溯算法采样结果保留故障现场日志
- 通过 [Grafana Alloy](https://grafana.com/docs/alloy/latest/) 采集指标，在容器环境中会采集全量日志，全量日志功能不支持虚拟机
- 集成 [node-agent](https://github.com/CloudDetail/node-agent) 探针，获取网络延时指标和进程状态指标

在 OneAgent 中，链路数据、日志数据直接发送至 OpenTelemetry Collector，指标数据经 Alloy 采集之后发送至 OpenTelemetry Collector。
