## 0. Completion Gate

- [x] 0.1 Do not treat documentation freeze or backend-only implementation as full completion; AgentSRE runtime enforcement, commercial frontend adaptation, and end-to-end validation must all be finished before this change is complete.

## 1. Freeze The Canonical Contract

- [x] 1.1 Define the canonical findings node fields, supported evidence sections, and node classification vocabulary in the change docs and implementation notes.
- [x] 1.2 Define the final merge precedence for investigations reconstruction, dataplane enrichment, and template defaults so one resolved node projection owns status, summary, detail text, and evidence.
- [x] 1.3 Define the explicit unsupported or unavailable behavior for incomplete checks, including whether unsupported nodes remain clickable or degrade to summary-only nodes.

## 2. Normalize AgentSRE Findings Assembly

- [x] 2.1 Refactor AgentSRE findings assembly so every emitted node is finalized from one resolved projection instead of partial field backfills.
- [x] 2.2 Normalize node status and classification across investigations-first and dataplane fallback paths so the final payload uses one consistent vocabulary.
- [x] 2.3 Update node builders that currently return placeholder or weak outputs to emit real evidence payloads or explicit unsupported states.
- [x] 2.4 Ensure each evidence-capable node publishes canonical detail fields and only the supported evidence sections (`metrics`, `table`, `logs`, `traces`) needed by the frontend.

## 3. Align Commercial Frontend Consumption

- [x] 3.1 Update the commercial findings API types and adapters to consume the canonical node contract without undocumented aliases.
- [x] 3.2 Fix FindingsTree node transformation and click selection payloads so NodeDetail receives canonical summary, content, status, and evidence fields.
- [x] 3.3 Update NodeDetail rendering to show canonical fields in the defined order, hide missing sections cleanly, and present unsupported nodes explicitly.
- [x] 3.4 Align frontend status handling and visual states with the normalized backend status vocabulary used by the canonical contract.

## 4. Add Coverage And Regression Checks

- [x] 4.1 Add AgentSRE payload tests or fixtures for investigations-only nodes, dataplane-enriched nodes, unsupported nodes, and empty-evidence regressions.
- [x] 4.2 Add commercial frontend tests for tree-to-detail mapping covering canonical content, metrics, table, logs, traces, and omitted sections.
- [x] 4.3 Validate representative conversation threads end-to-end to confirm node status, summary, and evidence remain consistent after merge and render.

## 5. Synchronize Docs And Rollout

- [x] 5.1 Update commercial repo and AgentSRE docs to record the canonical findings contract, cross-repo ownership boundaries, and unsupported-node policy.
- [x] 5.2 Document the rollout sequence for contract-first deployment, compatibility handling during transition, and the follow-up coverage backlog for remaining unsupported checks.
