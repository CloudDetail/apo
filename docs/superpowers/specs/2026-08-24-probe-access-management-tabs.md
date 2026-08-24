# Probe Access And Management Tabs

## Boundary

The existing probe access workflow and the new remote probe management workflow share the APO application shell but not the same content view.

- `探针接入` owns manual cluster registration, edit, delete, install commands, and Bots Management.
- `探针管理` owns automatically discovered clusters, total probe resource usage, and the remote configuration entry.
- The selected tab is stored in the `tab` query parameter so refresh and back navigation preserve context.

## Configuration scope

The `/probe-management/config` workflow has three explicit modes. `scope=global` manages reusable templates and the global default circuit policy. `nodeId` edits one node. With neither parameter, the page selects one or more clusters and performs a template-based batch delivery.

There is no global, cluster, and node precedence chain. A probe runs the last configuration release delivered to its node and probe type. A single-node edit reads every probe's current generation, combines those current files with the user's edits, and submits the complete node change through one action. Any generation conflict rejects the whole action. The UI may switch between files for editing, but it must not expose a separate save action for each probe.

A multi-cluster edit requires a template. It renders the template first, then atomically creates a new release for every target node. The rendered template replaces the selected probe's previous content; it is never merged with node-specific content and cannot partially succeed.

Configuration files and circuit policies are unrelated records and delivery actions. Saving configuration files never changes circuit settings or their version. The global circuit policy is only the default returned to a node without a node policy. Once a node policy is saved, every probe on that node receives that latest node policy. Multi-cluster circuit delivery writes the same policy as a full node-policy replacement for every target node.

Log configuration is a third independent setting. It defines which files or managed-probe stdout on the target node or Kubernetes cluster may be used by the real-time log viewer. A node save replaces the complete source list for one probe, and a multi-cluster save replaces it for every target node. It does not create a configuration generation or change circuit policy. Without a saved log configuration, daemon-go uses its deployment-time sources; an explicitly saved empty list disables log access for that node and probe. The log configuration tab and the real-time log drawer remain separate workflows.

The global configuration page edits the overall default policy. A node policy can set one overall policy and optional per-probe policies. When a node currently uses the global default, starting a per-probe edit copies the effective overall values into that probe's draft; saving writes a complete node policy and does not create a runtime precedence chain. The node's reported probe states, not the presence of a configuration file, determine which probes are available for circuit editing.

Circuit enforcement is local to daemon-go and probe-guard. An open circuit blocks starting or restarting only the affected managed probe. It must not stop daemon-go registration, desired-state polling, configuration download and staging, state reporting, command polling, or command-result reporting to the server. A staged configuration remains pending until reset and is used by the recovered probe. Automatic recovery retries probe startup locally with exponential backoff capped at one hour; it does not ask the server whether the probe may start on every attempt.

Before the first server assignment, probe-guard reads the default files from its own config-link mappings and includes them in its daemon UDS registration. daemon-go forwards the snapshot through the agent API. The backend stores only the first snapshot for a node and probe, and ignores it when a release already exists. Collection or upload failure is logged but does not block probe startup or normal desired-config synchronization.

## Data fallback

Probe management merges configured clusters with cluster IDs observed by the Dataplane. Missing resource totals render as unavailable and are logged. Development-only sample cluster, resource, node, version, and delivery task data is allowed when sources are empty; production must not invent these values or create a local-only delivery task.

## Development preview

The `/probe-management` route family may bypass user, permission, license, deployment-mode, and data-group requests only in the Vite development environment. This preview still renders the normal APO header and sidebar, using the existing APO default deploy mode when the backend is not authenticated. Production routes remain wrapped by `AuthRouter` and must never use the preview identity or menu.

## Backend ownership

Remote probe management is part of the APO backend. The APO process owns both `/api/agent-config/*` for authenticated console operations and `/api/agent/v1/*` for daemon registration, state reporting, initial-config capture, desired-config polling, and commands. Development and deployment must not run `agentconfig-smoke` as a parallel backend or proxy the APO frontend to it.

Agent requests use `AGENT_CONFIG_API_KEY` plus `X-Agent-Node-ID`; console requests keep APO's existing user and license middleware. Both API families persist through the APO database so login, probe state, templates, releases, commands, and initial baselines share one runtime and one migration lifecycle.
