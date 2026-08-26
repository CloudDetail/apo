# Probe Access And Management Tabs

## Boundary

The existing probe access workflow and the new remote probe management workflow share the APO application shell but not the same content view.

- `探针接入` owns manual cluster registration, edit, delete, install commands, and Bots Management.
- `探针管理` owns automatically discovered clusters, total probe resource usage, and the remote configuration entry.
- The selected tab is stored in the `tab` query parameter so refresh and back navigation preserve context.

## Configuration scope

The `/probe-management/config` workflow has three explicit modes. `scope=global` manages reusable templates and the global default circuit policy. `nodeId` edits one node. With neither parameter, the page selects one or more clusters and performs a template-based batch delivery.

There is no global, cluster, and node precedence chain. For every ordinary component, a node runs the last release delivered to that component. A single-node `配置文件` tab submits all non-preload components as one release group. `日志采集`, `注入控制`, and `熔断策略` are special controls with dedicated latest-state contracts and different execution paths. They do not expose configuration history or rollback. Unsaved drafts in another tab are neither submitted nor discarded. Both the legacy complete-node update API and the ordinary grouped update API reject preload. Batch template delivery likewise ignores legacy preload entries and delivers only the complete ordinary configuration set; the VM batch page exposes preload separately in `注入控制` and submits it through the dedicated batch injection-control API.

The single-node configuration-file editor presents non-preload components and their files in one tree navigator. The preload component and `libapoinstrument.conf` are absent from that tree and can only be changed from the injection-control tab. Selecting an ordinary file changes the active editor, history, and template context; saving submits the complete non-preload group.

The registered component and file-path tree is deployment-owned and read-only in the server UI. Template editing, node editing, and template delivery may only replace file content; adding, removing, or renaming a registered configuration file is rejected by the backend.

For a non-Kubernetes node with a registered `preload` component and `libapoinstrument.conf`, an independent `注入控制` tab exposes Java, Python, Node.js, and .NET as full-width rows. Each row shows the runtime, injection timing, retained language section, and switch. The complete generated INI is not continuously displayed; a `配置预览` button opens it on demand. The preload runtime treats the presence of each language section as enabled, so disabling a language must remove that live section. APO wraps the original section in a standard INI comment block instead of discarding it, allowing a later enable action to restore the exact language parameters. Enabling a language that has no retained section inserts the package-compatible default section before the blacklist sections. Unknown sections and their ordering are preserved. Duplicate language sections or damaged comment boundaries make the visual controls read-only and report an explicit error; they must never cause silent content loss. Saving this tab uses the specialized node injection-control API. It keeps a monotonic generation for safe daemon application but does not expose Release history or rollback.

The cluster and multi-cluster configuration modes allow file-content edits against a local copy of the selected template. They never update that source template in place. Any changed ordinary-configuration draft must be saved as a new template before delivery; while unsaved changes exist, the ordinary batch delivery action remains disabled. After creation, the page selects the new template and may deliver it only when its complete ordinary component and file-path set matches every target node registration. VM batch mode also has an independent `注入控制` tab. Preload has no template concept: the tab starts from the target nodes' current content, using the first target as the explicit draft when target contents differ and warning that delivery will unify them. Saving validates every target's preload registration and generation before one atomic write; it creates no user-facing history or rollback entry.

Configuration templates are categorized as `kubernetes` or `vm` because the two deployment types register different components and files. The default-configuration page exposes an explicit Kubernetes/VM segmented switch and lists only templates in that category; importing a template switches to its declared category. Cluster and node pages likewise list only templates matching their target category. Selecting Kubernetes and VM clusters together clears template selection and disables configuration delivery; the operator must deliver each category separately. The backend repeats this check for both batch and direct assignment before writing release history. Legacy templates without a category default to Kubernetes. A template category is immutable, and a cluster-page draft saved as a new template inherits the source category.

Template-based multi-cluster delivery does not accept per-node variable input. A single-node editor may display node-level variables reported by the probe, but these reported values are read-only and are not part of any configuration file. Server-side variable overrides and their protected-key policy, including non-editable identity values such as `cluster_ip`, are deferred until the override delivery contract is implemented.

A multi-cluster edit requires one complete template containing all managed components and exactly their registered files. It renders the template first, then atomically creates a new release for every component on every target node. The rendered template replaces the content of each target node's complete previous configuration; it is never merged with node-specific content and cannot partially succeed.

Configuration files and circuit policies are unrelated records and delivery actions. Saving configuration files never changes circuit settings or their version. The global circuit policy is only the default returned to a node without a node policy. Once a node policy is saved, daemon-go receives that complete latest policy. Multi-cluster circuit delivery writes the same policy as a full node-policy replacement for every target node.

Probe diagnostic log collection and user application log collection remain separate data models. Probe diagnostic self-collection is enabled by deployment default and is not edited in this workflow.

User application log collection is a child function of `/probe-management/config`, independent from complete managed-probe configuration generation. Every node inherits the container stdout rules saved for its cluster and may append node-scoped host file inputs containing an absolute base path, recursive directory depth, and file-name pattern. The node page renders inherited cluster rules read-only and links directly to that cluster's log collection tab for editing. Its final preview combines the inherited cluster files with the node additions without copying cluster records into node storage.

The cluster and batch configuration pages expose cluster-scoped container stdout inputs. Filters are opt-in: an unused filter is absent from the generated configuration rather than represented by an empty field. Namespace, pod, container, Pod-label, environment-variable, and container-label filters share one flat filter list. Non-Kubernetes clusters do not expose Kubernetes-only filters. Stream, parsing, multiline, tag, and buffering options live in a collapsed advanced section and start from iLogtail-compatible defaults. The fixed input type and generated internal name are never user-editable.

The log collection preview and save paths use the same backend normalizer and iLogtail renderer. Preview returns the exact `config/local/apo-*.yaml` input files produced from the draft; disabled sources produce no file. The generated input files do not own the log receiver destination, so changing an input never silently changes the cluster's log destination.

The current implementation persists these independent records and renders their exact input files, but it does not create an iLogtail release. Production delivery must first define how the dynamic `config/local/apo-*.yaml` paths and deployment-owned processors/flushers become one daemon-go iLogtail generation. The existing daemon-go deployment only accepts its declared fixed file set, so saving a management record must not be reported as remotely applied until that receiver contract is implemented and verified.

Ordinary ilogtail releases intentionally contain only `conf/plugin_logger.xml`.
The `config/local` directory is reserved for the log collection delivery path
and must not be added back to the probe configuration template.

The global configuration page edits the default node policy. The basic policy limits the sum of CPU Core and RSS used by every running managed component on the node. An optional advanced section may define an independent quota for a specific component. Components without an explicit advanced quota are evaluated only by the node total. Overall and component rules have no priority relationship: either rule may open a circuit, and saving always replaces the complete policy snapshot. The node's reported component states, not the presence of a configuration file, determine which components are available in advanced settings.

Circuit enforcement is local to daemon-go and probe-guard. daemon-go converts each component's reported CPU percentage to Core by dividing by 100, then sums Core and RSS for the node rule. A node-total breach opens every running managed component; a component breach opens only that component. An open circuit blocks starting or restarting affected components, but it must not stop daemon-go registration, desired-state polling, configuration download and staging, state reporting, command polling, or command-result reporting to the server. A staged configuration remains pending until reset and is used by the recovered component. Automatic recovery retries component startup locally with exponential backoff capped at one hour; it does not ask the server whether the component may start on every attempt.

The probe management table exposes node and cluster start/stop controls. Stop sends `OPEN_CIRCUIT` with reason `manual_node_stop` and each component's latest reported circuit epoch; start sends `RESET_CIRCUIT` for fused components and `RESTART` for stopped components whose circuit is normal. Preload configuration controls are excluded from process lifecycle operations. Cluster actions always target the cluster's complete registered node set even when the table is filtered. Partial command failure is reported as failure and must not be presented as a successful whole-cluster operation. These controls never stop daemon-go or its server communication.

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

Agent authentication requires only a non-empty `X-Agent-Cluster-ID` header. The Agent-specific middleware does not query cluster storage or validate that the cluster is registered. `X-Agent-Node-ID` remains an optional request identity used to scope node operations, but it is not an authentication requirement. Console requests keep APO's existing user and license middleware. Both API families persist through the APO database so login, probe state, templates, releases, commands, and initial baselines share one runtime and one migration lifecycle.
