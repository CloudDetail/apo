# Probe Access And Management Tabs

## Boundary

The existing probe access workflow and the new remote probe management workflow share the APO application shell but not the same content view.

- `探针接入` owns manual cluster registration, edit, delete, install commands, and Bots Management.
- `探针管理` owns automatically discovered clusters, total probe resource usage, and the remote configuration entry.
- The selected tab is stored in the `tab` query parameter so refresh and back navigation preserve context.

## Configuration scope

The `/probe-management/config` workflow has three explicit modes. `scope=global` manages reusable templates and the global default circuit policy. `nodeId` edits one node. With neither parameter, the page selects one or more clusters and performs a template-based batch delivery.

There is no global, cluster, and node precedence chain. A node runs the last complete configuration delivered to it. A single-node edit reads the current generation of every managed component, combines those current files with the user's edits, and submits the complete node change through one action. Any generation conflict rejects the whole action. The UI may switch between component files for editing, but it must not expose a separate delivery action for each component.

The single-node editor presents every managed component and its files in one tree navigator. Selecting a file only changes the active editor, history, and template context; it does not narrow the delivery scope. The single save action always writes the complete node configuration.

Template-based multi-cluster delivery does not accept per-node variable input. A single-node editor may display node-level variables reported by the probe, but these reported values are read-only and are not part of any configuration file. Server-side variable overrides and their protected-key policy, including non-editable identity values such as `cluster_ip`, are deferred until the override delivery contract is implemented.

A multi-cluster edit requires one complete template containing all managed components and all of their files. It renders the template first, then atomically creates a new release for every component on every target node. The rendered template replaces each target node's complete previous configuration; it is never merged with node-specific content and cannot partially succeed.

Configuration files and circuit policies are unrelated records and delivery actions. Saving configuration files never changes circuit settings or their version. The global circuit policy is only the default returned to a node without a node policy. Once a node policy is saved, daemon-go receives that complete latest policy. Multi-cluster circuit delivery writes the same policy as a full node-policy replacement for every target node.

Probe diagnostic log collection and user application log collection remain separate data models. Probe diagnostic self-collection is enabled by deployment default and is not edited in this workflow.

User application log collection is a child function of `/probe-management/config`, independent from complete managed-probe configuration generation. Every node inherits the container stdout rules saved for its cluster and may append node-scoped host file inputs containing an absolute base path, recursive directory depth, and file-name pattern. The node page renders inherited cluster rules read-only and links directly to that cluster's log collection tab for editing. Its final preview combines the inherited cluster files with the node additions without copying cluster records into node storage.

The cluster and batch configuration pages expose cluster-scoped container stdout inputs. Filters are opt-in: an unused filter is absent from the generated configuration rather than represented by an empty field. Namespace, pod, container, Pod-label, environment-variable, and container-label filters share one flat filter list. Non-Kubernetes clusters do not expose Kubernetes-only filters. Stream, parsing, multiline, tag, and buffering options live in a collapsed advanced section and start from iLogtail-compatible defaults. The fixed input type and generated internal name are never user-editable.

The log collection preview and save paths use the same backend normalizer and iLogtail renderer. Preview returns the exact `config/local/apo-*.yaml` input files produced from the draft; disabled sources produce no file. The generated input files do not own the log receiver destination, so changing an input never silently changes the cluster's log destination.

The current implementation persists these independent records and renders their exact input files, but it does not create an iLogtail release. Production delivery must first define how the dynamic `config/local/apo-*.yaml` paths and deployment-owned processors/flushers become one daemon-go iLogtail generation. The existing daemon-go deployment only accepts its declared fixed file set, so saving a management record must not be reported as remotely applied until that receiver contract is implemented and verified.

The global configuration page edits the default node policy. The basic policy limits the sum of CPU Core and RSS used by every running managed component on the node. An optional advanced section may define an independent quota for a specific component. Components without an explicit advanced quota are evaluated only by the node total. Overall and component rules have no priority relationship: either rule may open a circuit, and saving always replaces the complete policy snapshot. The node's reported component states, not the presence of a configuration file, determine which components are available in advanced settings.

Circuit enforcement is local to daemon-go and probe-guard. daemon-go converts each component's reported CPU percentage to Core by dividing by 100, then sums Core and RSS for the node rule. A node-total breach opens every running managed component; a component breach opens only that component. An open circuit blocks starting or restarting affected components, but it must not stop daemon-go registration, desired-state polling, configuration download and staging, state reporting, command polling, or command-result reporting to the server. A staged configuration remains pending until reset and is used by the recovered component. Automatic recovery retries component startup locally with exponential backoff capped at one hour; it does not ask the server whether the component may start on every attempt.

Before the first server assignment, probe-guard reads the default files from its own config-link mappings and includes them in its daemon UDS registration. daemon-go forwards the snapshot through the agent API. The backend stores only the first snapshot for a node and probe, and ignores it when a release already exists. Collection or upload failure is logged but does not block probe startup or normal desired-config synchronization.

## Data fallback

Probe management merges configured clusters with cluster IDs observed by the Dataplane. Missing resource totals render as unavailable and are logged. Development-only sample cluster, resource, node, version, and delivery task data is allowed when sources are empty; production must not invent these values or create a local-only delivery task.

## Development preview

The `/probe-management` route family may bypass user, permission, license, deployment-mode, and data-group requests only in the Vite development environment. This preview still renders the normal APO header and sidebar, using the existing APO default deploy mode when the backend is not authenticated. Production routes remain wrapped by `AuthRouter` and must never use the preview identity or menu.

## Route permission

`/probe-management` and its non-menu configuration route `/probe-management/config` use the existing `探针管理` feature permission. The configuration page is not an independent feature or menu permission. Both default and `PRODUCT_MODE=syncause` router seeds must contain the configuration route, and startup must create the `探针管理` feature-to-router mappings so existing roles that already own probe management can open the configuration page after an upgrade.

`/probe-management/settings` keeps its existing `接入中心` permission mapping because it configures integration access rather than a node's managed probe configuration.

## Backend ownership

Remote probe management is part of the APO backend. The APO process owns both `/api/agent-config/*` for authenticated console operations and `/api/agent/v1/*` for daemon registration, state reporting, initial-config capture, desired-config polling, and commands. Development and deployment must not run `agentconfig-smoke` as a parallel backend or proxy the APO frontend to it.

Agent requests use `AGENT_CONFIG_API_KEY` plus `X-Agent-Node-ID`; console requests keep APO's existing user and license middleware. Both API families persist through the APO database so login, probe state, templates, releases, commands, and initial baselines share one runtime and one migration lifecycle.
