# Cluster ownership management

## Purpose

Cluster identity and node ownership are separate operations:

- A cluster keeps an immutable environment type (`k8s` or `vm`). Its `cluster_id` is the stable identity; operators may only change the display name.
- A node keeps the environment reported at registration. Operators may change the node's effective `cluster_id` only through the cluster management page.
- Node ownership changes are deliberately separated from routine probe configuration because they change the scope of later configuration delivery.

Creating a cluster from the cluster management page only persists its `cluster_id`, display name, and immutable environment type. It does not create collector, metric, log, trace, or any other integration configuration.

Changing node ownership only updates the managed `CLUSTER_ID` override. It never creates, removes, or changes `CLUSTER_IP` or any other variable, and it does not trigger configuration delivery. `cluster_ip` is read only for mixed-cluster risk detection.

## Assignment rules

The backend is the authority for assignment validation. A target cluster must already exist, and the normalized node environment must equal the target cluster type. Kubernetes nodes cannot join VM clusters, and VM nodes cannot join Kubernetes clusters.

The UI filters target options to matching types for clarity, but this filtering is not a security or integrity boundary.

## Mixed cluster rule

APO treats a `cluster_ip` as mixed when registered nodes with that effective value contain both Kubernetes and VM environments. The effective value includes a persisted `CLUSTER_IP` override when one exists and otherwise uses the agent-reported value.

Mixed ranges are visible in cluster management and are blocked from explicit page-driven delivery, including node and batch configuration, injection, log collection, and circuit policy updates. Agent settings continue to return defaults according to the individual node's immutable environment. This preserves safe startup behavior while preventing an operator from applying one cluster policy across incompatible nodes.

The remediation path is to create type-specific `cluster_id` values and move each node to a matching cluster from `/probe-management/clusters`.
