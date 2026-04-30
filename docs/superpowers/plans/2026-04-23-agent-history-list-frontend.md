# Agent History List Frontend Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Build the `/agent` history list frontend on top of the new creator-scoped conversation history contract without reviving the legacy local session model.

**Architecture:** Keep the work incremental. First add a stable conversation history contract layer and pure time-formatting helpers, then refactor `AGUIContext` so URL `threadId` is the only thread source, and only after that wire the `/agent` layout and history interactions. Keep history list behavior in commercial-only `pro` code paths.

**Tech Stack:** React 18, Vite, TypeScript, Ant Design, `node:test`, existing APO `get()` request wrapper

---

### Task 1: Conversation History Contract Foundation

**Files:**
- Create: `frontend/src/pro/api/conversationHistory.ts`
- Create: `frontend/src/pro/api/conversationHistory.test.ts`
- Modify: `frontend/src/pro/api/conversation.ts`

- [ ] **Step 1: Write failing contract tests**

Add tests for:
- response normalization from `items`, `data.items`, and raw array shapes
- fallback title insertion for blank titles
- descending sort by `timeValue`
- time formatting thresholds for minutes, hours, yesterday, days, and absolute date

- [ ] **Step 2: Run targeted test command and verify RED**

Run:

```bash
rm -rf /tmp/apo-agent-history-tests
frontend/node_modules/.bin/tsc --target ES2020 --module commonjs --moduleResolution node --lib ES2020,DOM --types node --esModuleInterop --outDir /tmp/apo-agent-history-tests --rootDir frontend/src/pro/api frontend/src/pro/api/conversationHistory.test.ts frontend/src/pro/api/conversationHistory.ts
node --test /tmp/apo-agent-history-tests/conversationHistory.test.js
```

Expected before implementation: compile or runtime failure because `conversationHistory.ts` does not exist yet.

- [ ] **Step 3: Implement minimal contract helpers**

Implement:
- `ConversationHistoryItem`
- `ConversationHistoryListResponse`
- `normalizeConversationHistoryListResponse(response, fallbackTitle)`
- `formatConversationHistoryTime(timeValue, messages, now)`

- [ ] **Step 4: Wire API entrypoint**

Add `getConversationHistoryListApi({ cursor, limit, fallbackTitle })` in `frontend/src/pro/api/conversation.ts` and return normalized data.

- [ ] **Step 5: Run targeted test command and verify GREEN**

Run the same command from Step 2 and expect the `conversationHistory.test.js` suite to pass.

### Task 2: AGUI Thread State Refactor

**Files:**
- Modify: `frontend/src/pro/contexts/AGUIContext.tsx`
- Modify: `frontend/src/pro/components/AGUIChat/index.tsx`
- Modify: `frontend/src/pro/services/ag-ui-client.ts`

- [ ] **Step 1: Add thread-oriented helpers before UI work**

Introduce explicit helpers for:
- loading a history thread
- resetting to new conversation state
- cancelling or isolating a running stream before thread switches
- clearing findings / evidence / detail state together with thread switches

- [ ] **Step 2: Remove local session influence from active thread behavior**

Keep legacy local session code disabled and ensure:
- history load does not create a new `threadId`
- new conversation clears `threadId`
- first message in a new thread creates one stable `threadId`

- [ ] **Step 3: Verify thread state behavior**

Run:

```bash
cd frontend && npm run build
```

Then manually verify `/agent`:
- refresh with `threadId`
- clear to new conversation
- send first message and confirm URL write-back

### Task 3: `/agent` History List UI

**Files:**
- Modify: `frontend/src/pro/views/agent/index.tsx`
- Create: `frontend/src/pro/views/agent/components/HistoryList.tsx`
- Create: `frontend/src/pro/views/agent/components/HistoryListItem.tsx`
- Modify: `frontend/src/pro/components/AGUIChat/AGUIChat.module.scss`
- Modify: `frontend/public/locales/zh/pro/agent.json`
- Modify: `frontend/public/locales/en/pro/agent.json`

- [ ] **Step 1: Add history list container and fetch state**

Render desktop two-column layout first. Keep mobile fallback simple but accessible.

- [ ] **Step 2: Connect list interactions**

Implement:
- list fetch on mount
- selected item highlight
- URL `threadId` sync on click
- empty and error states
- “load more” if backend returns pagination metadata

- [ ] **Step 3: Verify layout and interaction**

Run:

```bash
cd frontend && npm run build
```

Then manually verify:
- desktop fixed history list
- mobile fallback
- click history entry
- invalid `threadId` recovery

### Task 4: Change Docs And Verification

**Files:**
- Modify: `openspec/changes/add-agent-history-list/tasks.md`
- Create or Modify: `openspec/changes/add-agent-history-list/implementation-summary.md`

- [ ] **Step 1: Mark completed OpenSpec tasks only when fully done**

- [ ] **Step 2: Update implementation summary with actual frontend files changed, current deviations, and remaining follow-ups**

- [ ] **Step 3: Run final verification**

Run:

```bash
cd frontend && npm run build
```

And re-run the targeted `conversationHistory` test command from Task 1.
