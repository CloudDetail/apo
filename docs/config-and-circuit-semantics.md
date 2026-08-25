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
- Each release stores per-file SHA-256 values and a deterministic SHA-256 for
  the complete file bundle. The single-file `content` field remains available
  for compatibility, while multi-file releases use `files` as the source of
  truth.
- A multi-cluster update requires a template. The service validates and
  renders the template before writing, then atomically creates releases for
  all target nodes. A template may contain one or more named files; its exact
  file set must match the target probe's current file set. Template output
  fully replaces every file for the selected probe, and node-specific content
  is not merged. Legacy templates without named files remain valid only for a
  single-file probe and are rejected for directory-style probes such as
  ilogtail.

## Template variables

Template variables describe stable target-node attributes. They are not
operator-supplied deployment overrides. daemon-go reports the node attributes
during registration, and the service renders a template separately for each
target node before creating its releases.

The canonical variable set required by the v1.0.3 Kubernetes agent package is:

| Variable | Source | Used by |
| --- | --- | --- |
| `node_name` | Kubernetes `spec.nodeName` | apo-otel-collector-agent, vmagent |
| `node_ip` | Kubernetes `status.hostIP` | vmagent |
| `cluster_id` | Node registration's cluster assignment | apo-otel-collector-agent, vmagent |

`cluster_id` is the exception to node self-discovery. daemon-go sends its
configured cluster identity as part of registration, while the service uses
the persisted node-to-cluster assignment as the rendering source of truth.
An empty reported value must not erase a non-empty persisted cluster
assignment.

Package-specific environment names such as `MY_NODE_NAME`, `MY_NODE_IP`,
`VM_HOSTNAME`, `VM_MY_NODE_IP`, `_node_name_`, and `_node_ip_` are aliases of
the canonical node attributes. They are not separate template variables. The
deployment manifest must inject the host IP into daemon-go explicitly; deriving
it from the daemon Pod network is not reliable.

The current v1.0.3 configuration bundle contains node or cluster-dependent
values only in `apo-otel-collector-agent/otelcol-config.yaml` and
`vmagent/scrape.yml`. The other ten managed files do not require reported
variables. Runtime-only values such as Pod UID, resource limits, mount paths,
and diagnostic log paths are outside the template-variable contract.

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
  target node in one operation.

Automatic recovery uses the existing epoch-checked circuit reset command. The
first retry waits for `recoveryInitialDelaySeconds`; subsequent retries double
the wait up to `recoveryMaxDelaySeconds`, which cannot exceed one hour. The
attempt counter is cleared only after the recovered probe remains below its
resource limits for one complete breach-duration window.

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
