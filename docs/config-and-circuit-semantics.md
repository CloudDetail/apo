# Agent configuration and circuit semantics

## Runtime rule

There is no runtime precedence between global, cluster, and node configuration.
For each probe, the node runs the last successfully delivered immutable
configuration release. `generation` is monotonic for a node and probe type.

## Configuration files

- A single-node edit reads the latest configuration for every probe and every
  file on the node. The client submits the exact complete file set in one
  node-level request with the observed generation of each probe. Missing,
  duplicated, or unexpected file paths reject the whole request.
- The management API does not expose a single-probe assignment operation.
  Per-probe endpoints are retained only for daemon delivery, initial snapshot
  inspection, and runtime commands.
- The service rechecks every generation and saves every changed probe release
  in one transaction. One stale generation rejects the complete node update.
- Before the first release, repeated registration refreshes the Generation 0
  snapshot so a deployment upgrade may add or remove managed files. After the
  first release, registration never overwrites the last delivered file set.
- Server template editing, node editing, and batch delivery never add, remove,
  or rename registered configuration files. They may only replace the content
  of the exact file set reported by the probe registration.
- The management history presents that Generation 0 snapshot as configuration
  version 1. It is an already-active local baseline, not a pending server
  release, so recording it never downloads the same files back to the probe or
  restarts the process. Server release Generation 1 is presented as version 2,
  and later display versions remain `generation + 1`.
- Each release stores per-file SHA-256 values and a deterministic SHA-256 for
  the complete file bundle. `files` is the source of truth for every release;
  a single-file response may additionally expose its content directly for the
  editor.
- Every Generation owns a complete content snapshot for the same registered
  file set. Loading an older history version restores those file contents into
  the editor without automatically delivering them.
- A multi-cluster update requires a template. The service validates and
  renders the template before writing, then atomically creates releases for
  all target nodes. Every template component must contain exactly the paths
  registered by every target node; node-specific content is not merged.
- Every template has a `clusterType`: `kubernetes` or `vm`. `k8s` is accepted
  as a Kubernetes import alias, and daemon registration value `binary` maps to
  VM. Existing templates and imported YAML without the field remain
  Kubernetes templates. Template category cannot be changed in place because
  the registered components and files are category-specific; a different
  category requires a new template. Assignment rejects a template whose
  category differs from any target node before creating release history.

The ordinary ilogtail configuration contains only `conf/plugin_logger.xml`.
Files under `config/local` describe which target or Kubernetes logs to collect
and belong exclusively to log collection management; they are never included
in an ordinary ilogtail release or configuration template.

Environment-backed behavior is still managed as configuration when it is part
of the probe's runtime contract. ebpf-agent includes `load-env.sh` in its
three-file bundle. `apo-odiglet` is a first-class managed component rather than
a lifecycle-only helper: it controls APM probe injection and configuration on
the target node, and its `odiglet.env` controls signal exporters, OpAMP, and
SkyWalking agent logging through the ordinary Generation flow. `NODE_NAME`,
`NODE_IP`, `CURRENT_NS`, and the collector endpoints derived from `NODE_IP`
remain deployment-injected runtime values and are not replaced by the managed
environment file.

## Template variables

Server-rendered template variables use the dedicated `{% VARIABLE_NAME %}`
syntax. Shell expressions such as `${NAME}` and `${NAME:-fallback}` are runtime
content and remain unchanged. Defaults and validation rules belong to template
variable metadata rather than the placeholder text. Only variables referenced
with the server syntax participate in assignment validation; obsolete metadata
from the former Shell-style syntax is ignored when a template is read.

The template page imports YAML through the backend rather than parsing it in
the browser:

```yaml
name: ebpf-production
description: Production eBPF settings
clusterType: kubernetes
variables:
  - name: LEVEL
    required: true
configs:
  - probeType: ebpf-agent
    files:
      - path: config/log4rs.yml
        template: |
          level: {% LEVEL %}
```

Imported YAML contains logical paths and template content only. Target-machine
paths remain in the probe-guard registration file, so an imported template
cannot redirect writes to an arbitrary location.

Template variables describe stable target-node attributes. They are not
operator-supplied deployment overrides. daemon-go reports the node attributes
during registration, and the service renders a template separately for each
target node before creating its releases.

The canonical server variable used by the v1.0.3 Kubernetes agent package is:

| Variable | Source | Used by |
| --- | --- | --- |
| `CLUSTER_ID` | Node registration's cluster assignment | apo-otel-collector-agent, vmagent |

daemon-go sends its configured cluster identity as part of registration, while
the service uses the persisted node-to-cluster assignment as the rendering
source of truth. Operator-provided `CLUSTER_ID` values are ignored so one
cluster cannot be rendered with another cluster's identity. An empty reported
value must not erase a non-empty persisted cluster assignment.

Package-specific environment names such as `MY_NODE_NAME`, `MY_NODE_IP`,
`VM_HOSTNAME`, `VM_MY_NODE_IP`, `_node_name_`, and `_node_ip_` are aliases of
the canonical node attributes. They are not separate template variables. The
deployment manifest must inject the host IP into daemon-go explicitly; deriving
it from the daemon Pod network is not reliable.

The current v1.0.3 configuration bundle uses `{% CLUSTER_ID %}` only in
`apo-otel-collector-agent/otelcol-config.yaml` and `vmagent/scrape.yml`.
Node name and IP continue to use deployment-injected runtime environment
expressions. `originx-gc-agent` registers `originx-gc-agent.env` as its single
managed configuration file. probe-guard reads the active Generation through
`env_file` immediately before starting the probe, while `MY_NODE_NAME` and
`MY_NODE_IP` remain deployment-injected runtime values. Runtime-only values
such as Pod UID, resource limits, mount paths, and diagnostic log paths are
outside the template-variable contract.

## Diagnostic log sources

Diagnostic log sources are independent from configuration releases and circuit
policy. They describe logs on the target node or Kubernetes cluster that
daemon-go may list, preview, and download. Saving them does not create a
configuration generation or modify circuit policy.

- A node-and-probe save fully replaces the previous source list.
- Multi-cluster delivery fully replaces the source list for every target node.
- Without a saved source list, daemon-go keeps its deployment-time defaults.
- An explicitly saved empty list disables log access for that node and probe.

## Circuit policy

Circuit policy persistence is independent from configuration templates,
bindings, releases, and generations. Saving either side never changes the
other.

- The server has one global circuit policy bundle. The bundle contains one
  default applied to all probes and optional policies for named probe types.
  Built-in defaults are enabled, 80 percent CPU, 512 MiB RSS, a 30 second
  breach duration, and automatic recovery; server environment variables may
  replace these defaults.
- A node without a saved circuit policy receives the current global default.
- Saving a node policy replaces the complete policy bundle for that node. The
  daemon receives the named probe policy when one exists in that latest bundle,
  otherwise it receives the bundle default. This is one delivered bundle, not
  a runtime priority chain.
- Multi-cluster circuit delivery replaces the complete node policy for every
  target node in one operation. Its top-level resource thresholds must use
  host-system CPU and memory percentages; process CPU Core and RSS thresholds
  remain available only for named probe policies. The percentage is evaluated
  independently on every target node, not as a cluster-wide aggregate.
- Host-system thresholds cannot appear in a named probe policy. daemon-go reads
  the system values locally; the service does not make fuse or recovery
  decisions.

Automatic recovery uses the existing epoch-checked circuit reset command. The
first retry waits for `recoveryInitialDelaySeconds`; subsequent retries double
the wait up to `recoveryMaxDelaySeconds`, which cannot exceed one hour. The
attempt counter is cleared only after the recovered probe remains below its
resource limits for one complete breach-duration window. A circuit opened by
system pressure waits for a complete below-threshold host sample before its
first recovery attempt. After a complete system sample stays at or above a
configured threshold for the breach duration, daemon-go opens every running
managed probe on that node in the same enforcement pass.

The service only persists and distributes policy. daemon-go and probe-guard
evaluate resource usage and execute fuse/recovery locally without asking the
service for each decision. `FUSED` only blocks start or restart of the managed
probe process. It does not interrupt daemon-go registration, heartbeat, policy
or configuration pull, command handling, or status reporting. A new
configuration can be downloaded and staged while fused, but it cannot trigger
a probe start or restart; recovery starts the probe with the latest staged
configuration.

The daemon may also have a deployment-time local circuit policy. It is used
before the first valid server response and retained when settings requests
fail. Invalid remote values are logged and never replace the last valid policy.

## Runtime status freshness

daemon-go reports each managed probe status once per minute. The management
API presents a stored state as `OFFLINE` when no report has arrived for two
minutes, while preserving the last runtime payload for diagnostics. The UI
must not count an offline state as healthy or include its last resource sample
in node and cluster totals.
