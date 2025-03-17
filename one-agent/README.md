# OneAgent
OneAgent is designed to collect various observability data from monitored environments, including trace data, log data, and metrics. OneAgent can be deployed in various environments, including traditional servers, containers, and Kubernetes environments.

**OneAgent** integrates the following components:

- Automatic installation of [OpenTelemetry](https://github.com/open-telemetry) agents based on Linux Preload mechanism and Kubernetes Webhook mechanism, effective after application restart
- Grafana Beyla: Non-intrusive trace data collection for Go applications using [Grafana Beyla](https://github.com/grafana/beyla)
- Collection of Polaris causal metrics using eBPF technology, implementing trace data [retrospective sampling algorithm](https://www.usenix.org/conference/nsdi23/presentation/zhang-lei)
- Collection of fault scene logs through [ilogtail](https://github.com/alibaba/ilogtail), preserving fault scene logs based on retrospective sampling results
- Metric collection through [Grafana Alloy](https://grafana.com/docs/alloy/latest/), supporting extensible metric collection through configuration
- Integration with [node-agent](https://github.com/CloudDetail/node-agent) probe to obtain network latency metrics and process status metrics

In OneAgent, trace data and log data are sent directly to the OpenTelemetry Collector, while metric data is collected through Alloy before being sent to the OpenTelemetry Collector.
