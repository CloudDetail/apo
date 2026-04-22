## ADDED Requirements

### Requirement: Findings Response Exposes A Canonical Node Contract
The findings data response SHALL expose a canonical node projection that is sufficient for both tree rendering and right-side detail rendering without relying on undocumented aliases. Each returned node SHALL provide a stable identifier, display title, canonical classification, normalized status, optional summary text, optional detail text, and zero or more evidence sections from the supported evidence types (`metrics`, `table`, `logs`, `traces`).

#### Scenario: Frontend renders from canonical fields only
- **WHEN** the frontend receives a findings response
- **THEN** it MUST be able to build the tree and the detail selection state from the canonical node fields alone

#### Scenario: Unsupported sections are represented explicitly
- **WHEN** a node does not have a value for a supported evidence section
- **THEN** the response MUST omit that section or mark it unavailable, rather than requiring the frontend to infer data from undocumented fallback fields

### Requirement: Findings Nodes Declare Stable Classification Semantics
Each findings node SHALL be classified as exactly one of `summary`, `evidence`, or `unsupported`, and that classification SHALL govern how the node is interpreted. Summary nodes MAY aggregate child results without carrying raw evidence. Evidence nodes MUST represent a concrete diagnostic check. Unsupported nodes MUST represent a known gap or intentionally unavailable check and MUST NOT masquerade as completed evidence.

#### Scenario: Summary node remains aggregative
- **WHEN** a node is classified as `summary`
- **THEN** the node MAY present aggregate status and summary text without requiring raw evidence sections

#### Scenario: Unsupported node is not treated as completed evidence
- **WHEN** a node is classified as `unsupported`
- **THEN** the consumer MUST preserve that classification in selection state and aggregation logic, and MUST NOT treat the node as completed normal evidence; the UI MAY de-emphasize it and reuse normal-equivalent visual styling to keep attention on abnormal data

### Requirement: Evidence Nodes Are Self-Justifying
Every clickable evidence node SHALL provide at least one self-justifying outcome: populated `metrics`, populated `table`, populated `logs`, populated `traces`, or an explicit unsupported or unavailable explanation. The system MUST NOT emit a clickable evidence node that shows a default conclusion or default normal status without real evidence or an explicit unsupported marker.

#### Scenario: Evidence node with real payload
- **WHEN** a node is classified as `evidence` and the underlying check completes
- **THEN** the node MUST include at least one supported evidence section or structured detail content explaining the conclusion

#### Scenario: Evidence node without implementation coverage
- **WHEN** a diagnostic check does not yet retrieve or evaluate real data
- **THEN** the node MUST be marked unsupported or unavailable instead of being emitted as a normal evidence node with placeholder certainty

### Requirement: Final Node Projection Keeps State And Evidence Consistent
The final findings payload SHALL be assembled from a single resolved projection per node. If investigations reconstruction, dataplane fallback, or another enrichment step contributes updated status, summary, detail text, or evidence, the final node SHALL publish the mutually consistent result from that resolved projection. The system MUST NOT publish evidence from one source while retaining contradictory status or conclusion from an earlier template.

#### Scenario: Dataplane fallback updates node state
- **WHEN** dataplane enrichment adds or replaces a node's evidence and computes a more specific conclusion or status
- **THEN** the final emitted node MUST publish the enriched status, conclusion, and evidence together

#### Scenario: Investigations result remains authoritative when complete
- **WHEN** investigations reconstruction already provides a complete and non-contradictory node projection
- **THEN** the final emitted node MUST preserve that projection without mixing in stale template fields

### Requirement: Detail Panel Consumes Canonical Node Detail Fields
The frontend selection flow for findings nodes SHALL pass canonical node detail fields to the right-side panel. The right-side panel SHALL render status and title first, followed by at most one primary canonical text block (`summaryConclusion` when present, otherwise `content`), and then any available evidence sections, and SHALL hide absent sections cleanly. The frontend MUST NOT depend on deprecated or undocumented aliases to display primary evidence.

#### Scenario: Canonical content is displayed in the detail panel
- **WHEN** a selected node includes canonical summary text, detail text, or evidence sections
- **THEN** the right-side panel MUST prefer `summaryConclusion` over `content`, and only fall back to `content` when summary text is absent

#### Scenario: Missing evidence sections do not produce phantom content
- **WHEN** a selected node lacks a given evidence section
- **THEN** the right-side panel MUST omit that section instead of showing empty placeholders or reading from unrelated fields

### Requirement: Coverage Gaps Are Surfaced Explicitly
Nodes that are retained in the tree for narrative completeness but do not yet perform real retrieval or judgment SHALL be surfaced as explicit coverage gaps in the canonical contract. Such nodes MUST be classified as `unsupported` or converted to non-clickable summaries until real evidence behavior is implemented.

#### Scenario: Placeholder diagnostic check remains visible
- **WHEN** product requirements keep a partially implemented check in the tree
- **THEN** the node MUST remain classified as `unsupported` and MUST NOT be counted as completed normal evidence, even if a consumer de-emphasizes its visual treatment

#### Scenario: Node becomes evidence-capable later
- **WHEN** a previously unsupported check gains real retrieval and evaluation logic
- **THEN** it MUST transition to the evidence-node rules and satisfy the evidence completeness requirement
