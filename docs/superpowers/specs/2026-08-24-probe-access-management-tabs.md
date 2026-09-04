# Probe Access And Management

## Boundary

`/probe-management` is the single probe operations view. The sidebar routes directly to it; the legacy `探针接入` page and `ClusterTable` implementation are removed.

- The route has two top-level tabs: `探针管理` and `Bots Management`.
- `探针管理` is the default. `Bots Management` is stored as `?tab=bots`; an absent or unknown value falls back to probe management.
- Only the active tab content is mounted. Slack status loading and polling therefore run only while `Bots Management` is visible.
- Bot management is a standalone tab panel, not a side-by-side section of cluster access.

- The header owns the `接入新集群` action and routes to `/probe-management/settings`.
- The table merges manually registered integration clusters with clusters discovered from agent node reports by exact `cluster_id`.
- A manually registered cluster remains visible before any node reports. A discovered-only cluster remains visible without a manual registration.
- Integration settings, display-name editing, deletion, and installation commands are table actions. Installation commands are shown only when a manual integration record exists.
- Deleting a manual record removes its integration metadata; reported nodes with the same `cluster_id` continue to appear as a discovered-only cluster.

Manual registration requires `cluster_id`. Its display name is optional; the backend trims both values and persists `cluster_id` as the display name when the submitted display name is empty. `cluster_id` is immutable after creation. Display-name changes use a metadata-only endpoint so they cannot rewrite trace, metric, log, or collector settings.

## Configuration scope

The `/probe-management/config` workflow has three explicit modes. `scope=global` manages reusable templates and the global default circuit policy. `nodeId` edits one node. With neither parameter, the page selects one or more clusters and performs a template-based batch delivery.

There is no global, cluster, and node precedence chain. For every ordinary component, a node runs the last release delivered to that component. A single-node `配置文件` tab submits all non-preload components as one release group. `日志采集`, `注入控制`, and `熔断策略` are special controls with dedicated latest-state contracts and different execution paths. They do not expose configuration history or rollback. Unsaved drafts in another tab are neither submitted nor discarded. Both the legacy complete-node update API and the ordinary grouped update API reject preload. Batch template delivery likewise ignores legacy preload entries and delivers only the complete ordinary configuration set; the VM batch page exposes preload separately in `注入控制` and submits it through the dedicated batch injection-control API.

The single-node configuration-file editor presents non-preload components and their files in one tree navigator. The preload component and `libapoinstrument.conf` are absent from that tree and can only be changed from the injection-control tab. Selecting an ordinary file changes the active editor, history, and template context; saving submits the complete non-preload group.

The registered component and file-path tree is deployment-owned and read-only in the server UI. Template editing, node editing, and template delivery may only replace file content; adding, removing, or renaming a registered configuration file is rejected by the backend.

For a non-Kubernetes node with a registered `preload` component and `libapoinstrument.conf`, an independent `注入控制` tab owns both the system-level Preload enable switch and the language configuration. The system switch sends `ENABLE_PRELOAD` or `DISABLE_PRELOAD`; it is not duplicated on the probe-management node row. Java, Python, Node.js, and .NET appear as full-width rows. Each row shows the runtime, injection timing, retained language section, and switch. The complete generated INI is not continuously displayed; a `配置预览` button opens it on demand. The preload runtime treats the presence of each language section as enabled, so disabling a language must remove that live section. APO wraps the original section in a standard INI comment block instead of discarding it, allowing a later enable action to restore the exact language parameters. Enabling a language that has no retained section inserts the package-compatible default section before the blacklist sections. Unknown sections and their ordering are preserved. Duplicate language sections or damaged comment boundaries make the visual controls read-only and report an explicit error; they must never cause silent content loss. Saving this tab uses the specialized node injection-control API. It keeps a monotonic generation for safe daemon application and participates in delivery auditing, but does not expose selectable configuration history or rollback.

The cluster and multi-cluster configuration modes allow file-content edits against a local copy of the selected template. They never update that source template in place. Any changed ordinary-configuration draft must be saved as a new template before delivery; while unsaved changes exist, the ordinary batch delivery action remains disabled. After creation, the page selects the new template and may deliver it only when its complete ordinary component and file-path set matches every target node registration. VM batch mode also has an independent `注入控制` tab. Preload has no template concept: the tab starts from the target nodes' current content, using the first target as the explicit draft when target contents differ and warning that delivery will unify them. Saving validates every target's preload registration and generation before one atomic write; it creates no selectable configuration-history or rollback entry, but its generation target is included in delivery auditing.

## Delivery audit

The probe-management header exposes `下发记录`. One audit item represents one successful generation-producing server action, not one component release. Node ordinary configuration, template batch delivery, node injection control, and batch injection control create audit items. Circuit policy, lifecycle commands, log preview, and log-collection management records are excluded because their current contracts do not provide the same per-generation apply acknowledgement.

The audit item and all of its node/component target generations are saved in the same database transaction as bindings and releases. A failed audit write rolls back the release write; timestamps must never be used to infer or merge delivery actions. No-op saves that create no release create no audit item.

Receipt state is derived against each immutable target generation using the latest daemon apply state. `APPLIED` requires `appliedGeneration >= targetGeneration`. `RECEIVED` requires `desiredGeneration >= targetGeneration` but not yet applied. `FAILED` additionally requires a failure/rejection/unavailable report for that target or a later desired generation. A missing or lower desired generation remains `PENDING`. A later applied generation proves that an older target was applied or superseded and therefore keeps the older audit target in `APPLIED`. The UI shows the aggregate node counts first and expands an item to the per-node and per-component generations, latest agent state, report time, and failure reason.

Configuration templates are categorized as `kubernetes` or `vm` because the two deployment types register different components and files. The default-configuration page exposes an explicit Kubernetes/VM segmented switch and lists only templates in that category; importing a template switches to its declared category. Cluster and node pages likewise list only templates matching their target category. Selecting Kubernetes and VM clusters together clears template selection and disables configuration delivery; the operator must deliver each category separately. The backend repeats this check for both batch and direct assignment before writing release history. Legacy templates without a category default to Kubernetes. A template category is immutable, and a cluster-page draft saved as a new template inherits the source category.

Template-based multi-cluster delivery does not accept per-node variable input. A single-node editor may display node-level variables reported by the probe, but these reported values are read-only and are not part of any configuration file. Server-side variable overrides and their protected-key policy, including non-editable identity values such as `cluster_ip`, are deferred until the override delivery contract is implemented.

A multi-cluster edit must be saved as a template before delivery. Delivery uses only that saved template and each target node's reported variables; it does not read, merge, or validate against the node's previous component or file set. The backend renders every ordinary component in the template for every target node, then atomically creates all releases so the batch cannot partially succeed.

Configuration files and circuit policies are unrelated records and delivery actions. Saving configuration files never changes circuit settings or their version. The global circuit policy is only the default returned to a node without a node policy. Once a node policy is saved, daemon-go receives that complete latest policy. Multi-cluster circuit delivery writes the same policy as a full node-policy replacement for every target node.

Probe diagnostic log collection and user application log collection remain separate data models. Probe diagnostic self-collection is enabled by deployment default and is not edited in this workflow.

User application log collection is a child function of `/probe-management/config`, independent from complete managed-probe configuration generation. Every node inherits the container stdout rules saved for its cluster and may append node-scoped host file inputs containing an absolute base path, recursive directory depth, and file-name pattern. The node page renders inherited cluster rules read-only and links directly to that cluster's log collection tab for editing. Its final preview combines the inherited cluster files with the node additions without copying cluster records into node storage.

The cluster and batch configuration pages expose cluster-scoped container stdout inputs. Filters are opt-in: an unused filter is absent from the generated configuration rather than represented by an empty field. Namespace, pod, container, Pod-label, environment-variable, and container-label filters share one flat filter list. Non-Kubernetes clusters do not expose Kubernetes-only filters. Stream, parsing, multiline, tag, and buffering options live in a collapsed advanced section and start from iLogtail-compatible defaults. The fixed input type and generated internal name are never user-editable.

The log collection preview and save paths use the same backend normalizer and iLogtail renderer. Preview returns the exact `config/local/apo-*.yaml` input files produced from the draft; disabled sources produce no file. The generated input files do not own the log receiver destination, so changing an input never silently changes the cluster's log destination.

The current implementation persists these independent records and renders their exact input files, but it does not create an iLogtail release. Production delivery must first define how the dynamic `config/local/apo-*.yaml` paths and deployment-owned processors/flushers become one daemon-go iLogtail generation. The existing daemon-go deployment only accepts its declared fixed file set, so saving a management record must not be reported as remotely applied until that receiver contract is implemented and verified.

Ordinary ilogtail releases intentionally contain only `conf/plugin_logger.xml`.
For both `ilogtail` and `loongcollector`, the `config/local` directory is reserved
for the log collection tab and is filtered from existing templates, node ordinary
configuration views, and ordinary releases.

The global configuration page edits the default node policy. The basic policy limits the sum of CPU Core and RSS used by every running managed component on the node. An optional advanced section may define an independent quota for a specific component. Components without an explicit advanced quota are evaluated only by the node total. Overall and component rules have no priority relationship: either rule may open a circuit, and saving always replaces the complete policy snapshot. The node's reported component states, not the presence of a configuration file, determine which components are available in advanced settings.

Circuit enforcement is local to daemon-go and probe-guard. daemon-go converts each component's reported CPU percentage to Core by dividing by 100, then sums Core and RSS for the node rule. A node-total breach opens every running managed component; a component breach opens only that component. An open circuit blocks starting or restarting affected components, but it must not stop daemon-go registration, desired-state polling, configuration download and staging, state reporting, command polling, or command-result reporting to the server. A staged configuration remains pending until reset and is used by the recovered component. Automatic recovery retries component startup locally with exponential backoff capped at one hour; it does not ask the server whether the component may start on every attempt.

The probe management table exposes node and cluster start/stop controls. Stop sends `OPEN_CIRCUIT` with reason `manual_node_stop` and each component's latest reported circuit epoch; start sends `RESET_CIRCUIT` for fused components and `RESTART` for stopped components whose circuit is normal. Preload configuration controls are excluded from process lifecycle operations. Cluster actions always target the cluster's complete registered node set even when the table is filtered. A successful command is only an acknowledgement: the UI keeps the action pending until `/api/agent-config/states` reports the requested circuit and child-process state, then replaces local state with that authoritative response. Periodic state requests started before or during the action are versioned and discarded after a lifecycle transition so their stale snapshots cannot overwrite the confirmed state. Command failure, state verification timeout, and partial completion are reported as failure and must not be presented as a successful whole-cluster operation. These controls never stop daemon-go or its server communication.

Before the first server assignment, probe-guard reads the default files from its own config-link mappings and includes them in its daemon UDS registration. daemon-go forwards the snapshot through the agent API. The backend stores only the first snapshot for a node and probe, and ignores it when a release already exists. Collection or upload failure is logged but does not block probe startup or normal desired-config synchronization.

## Data fallback

Probe management merges manually registered integration clusters with cluster IDs reported by agent nodes. Failure to load manual metadata is logged and the page continues with reported cluster IDs; the last successfully loaded manual metadata remains in memory. Missing resource totals render as unavailable and are logged. Production must not invent cluster, resource, node, version, or delivery-task values.

## Development preview

The `/probe-management` route family may bypass user, permission, license, deployment-mode, and data-group requests only in the Vite development environment. This preview still renders the normal APO header and sidebar, using the existing APO default deploy mode when the backend is not authenticated. Production routes remain wrapped by `AuthRouter` and must never use the preview identity or menu.

## Route permission

`/probe-management`, `/probe-management/settings`, and the non-menu configuration route `/probe-management/config` use the existing `探针管理` feature permission. The settings and configuration pages are not independent menu permissions. `/probe-management/settings` also keeps its legacy `接入中心` mapping so existing access-center roles do not lose access during the navigation migration. Both default and `PRODUCT_MODE=syncause` router seeds must contain these routes, and startup must create the feature-to-router mappings after an upgrade.

## Backend ownership

Remote probe management is part of the APO backend. The APO process owns both `/api/agent-config/*` for authenticated console operations and `/api/agent/v1/*` for daemon registration, state reporting, initial-config capture, desired-config polling, and commands. Development and deployment must not run `agentconfig-smoke` as a parallel backend or proxy the APO frontend to it.

Agent authentication requires only a non-empty `X-Agent-Cluster-ID` header. The Agent-specific middleware does not query cluster storage or validate that the cluster is registered. `X-Agent-Node-ID` remains an optional request identity used to scope node operations, but it is not an authentication requirement. Console requests keep APO's existing user and license middleware. Both API families persist through the APO database so login, probe state, templates, releases, commands, and initial baselines share one runtime and one migration lifecycle.
