# Findings Stream Cache Lifecycle Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Keep clicked findings streams useful across service switches without allowing unbounded SSE fan-out.

**Architecture:** `EvidencePanel` will track in-flight findings streams by `threadId:serviceName` instead of using one global abort controller. Stream events will always update that service's cache, while UI state will only sync for the currently selected service. A small set of pure lifecycle helpers in `findingsHydration.ts` will make terminal-cache and concurrency decisions testable.

**Tech Stack:** React 18, TypeScript, browser `fetch` streams, `AbortController`, Node test runner with `tsx` when available.

---

### Task 1: Add Lifecycle Decision Tests

**Files:**
- Modify: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/findingsHydration.test.ts`
- Modify: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/findingsHydration.ts`

- [ ] **Step 1: Write failing tests for cache terminal states and stream overflow**

Add these imports to `findingsHydration.test.ts`:

```ts
import {
  getFindingsStreamOverflowKey,
  isTerminalFindingsPhase,
  shouldStartFindingsStream,
} from './findingsHydration'
```

Add these tests:

```ts
test('done and failed findings cache entries are terminal unless refreshed', () => {
  assert.equal(isTerminalFindingsPhase('done'), true)
  assert.equal(isTerminalFindingsPhase('failed'), true)
  assert.equal(isTerminalFindingsPhase('streaming'), false)

  assert.equal(
    shouldStartFindingsStream({
      cachedPhase: 'done',
      isLoading: false,
      refresh: false,
    }),
    false,
  )
  assert.equal(
    shouldStartFindingsStream({
      cachedPhase: 'failed',
      isLoading: false,
      refresh: false,
    }),
    false,
  )
  assert.equal(
    shouldStartFindingsStream({
      cachedPhase: 'done',
      isLoading: false,
      refresh: true,
    }),
    true,
  )
})

test('in-flight findings streams are not duplicated', () => {
  assert.equal(
    shouldStartFindingsStream({
      cachedPhase: 'streaming',
      isLoading: true,
      refresh: false,
    }),
    false,
  )
})

test('stream overflow aborts the oldest non-active key only when at capacity', () => {
  assert.equal(getFindingsStreamOverflowKey(['thread:svc-a'], 'thread:svc-b', 2), null)
  assert.equal(
    getFindingsStreamOverflowKey(['thread:svc-a', 'thread:svc-b'], 'thread:svc-b', 2),
    'thread:svc-a',
  )
  assert.equal(
    getFindingsStreamOverflowKey(['thread:svc-a', 'thread:svc-b'], 'thread:svc-a', 2),
    'thread:svc-b',
  )
  assert.equal(getFindingsStreamOverflowKey(['thread:svc-a'], 'thread:svc-a', 1), null)
})
```

- [ ] **Step 2: Run the new tests and verify they fail because helpers are missing**

Run:

```bash
cd frontend && npx --yes tsx --test src/pro/components/AGUIChat/components/EvidencePanel/findingsHydration.test.ts
```

Expected: FAIL with missing exports from `findingsHydration.ts`.

- [ ] **Step 3: Add minimal lifecycle helpers**

Add to `findingsHydration.ts`:

```ts
export const isTerminalFindingsPhase = (phase?: FindingsHydrationPhase) =>
  phase === 'done' || phase === 'failed'

export const shouldStartFindingsStream = ({
  cachedPhase,
  isLoading,
  refresh = false,
}: {
  cachedPhase?: FindingsHydrationPhase
  isLoading: boolean
  refresh?: boolean
}) => {
  if (isLoading) {
    return false
  }
  if (!refresh && isTerminalFindingsPhase(cachedPhase)) {
    return false
  }
  return true
}

export const getFindingsStreamOverflowKey = (
  loadingKeys: Iterable<string>,
  activeKey: string | null,
  maxStreams: number,
) => {
  const keys = Array.from(loadingKeys)
  if (keys.length < maxStreams) {
    return null
  }
  return keys.find((key) => key !== activeKey) ?? null
}
```

- [ ] **Step 4: Re-run the focused tests**

Run:

```bash
cd frontend && npx --yes tsx --test src/pro/components/AGUIChat/components/EvidencePanel/findingsHydration.test.ts
```

Expected: PASS.

### Task 2: Replace Single Controller With Bounded In-Flight Streams

**Files:**
- Modify: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/index.tsx`
- Modify: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/findingsHydration.ts`
- Test: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/findingsHydration.test.ts`

- [ ] **Step 1: Import lifecycle helpers**

Update the `findingsHydration` import in `EvidencePanel/index.tsx`:

```ts
import {
  applyFindingsStreamEvent,
  buildFindingsCacheKey,
  createPendingFindingsShell,
  getFindingsStreamOverflowKey,
  isTerminalFindingsPhase,
  shouldStartFindingsStream,
  type FindingsCacheEntry,
} from './findingsHydration'
```

- [ ] **Step 2: Replace the single abort controller ref**

Replace:

```ts
const findingsAbortController = useRef<AbortController | null>(null)
```

with:

```ts
const findingsAbortControllers = useRef<Map<string, AbortController>>(new Map())
const maxConcurrentFindingsStreams = 2
```

- [ ] **Step 3: Add bounded abort helper inside `EvidencePanel`**

Add near the cache refs:

```ts
const abortFindingsStream = useCallback((cacheKey: string, phase: FindingsCacheEntry['phase'] = 'idle') => {
  findingsAbortControllers.current.get(cacheKey)?.abort()
  findingsAbortControllers.current.delete(cacheKey)
  loadingServices.current.delete(cacheKey)
  const cached = findingsCache.current.get(cacheKey)
  if (cached && cached.phase === 'streaming') {
    findingsCache.current.set(cacheKey, {
      ...cached,
      phase,
    })
  }
}, [])
```

- [ ] **Step 4: Let inactive stream events update cache**

Remove the early return from `updateFindingsCacheFromEvent`:

```ts
if (activeFindingsKey.current !== cacheKey) {
  return
}
```

Guard loading state changes with:

```ts
if (activeFindingsKey.current === cacheKey) {
  setLoadingFindings(false)
}
```

- [ ] **Step 5: Update stream start logic**

In `startServiceFindingsStream`, after setting `activeFindingsKey.current`, use:

```ts
const cached = findingsCache.current.get(cacheKey)
if (cached?.nodes?.length) {
  syncSelectedServiceFindings(cacheKey)
} else if (lastFindingsShellNodes.current.length) {
  findingsCache.current.set(cacheKey, {
    nodes: createPendingFindingsShell(lastFindingsShellNodes.current),
    phase: 'streaming',
  })
  syncSelectedServiceFindings(cacheKey)
}

const isLoading = loadingServices.current.has(cacheKey)
if (!shouldStartFindingsStream({ cachedPhase: cached?.phase, isLoading })) {
  setLoadingFindings(isLoading)
  return
}

if (isTerminalFindingsPhase(cached?.phase)) {
  setLoadingFindings(false)
  return
}

const overflowKey = getFindingsStreamOverflowKey(
  loadingServices.current,
  cacheKey,
  maxConcurrentFindingsStreams,
)
if (overflowKey) {
  abortFindingsStream(overflowKey)
}

const abortController = new AbortController()
findingsAbortControllers.current.set(cacheKey, abortController)
```

- [ ] **Step 6: Keep controller maps consistent**

In the `finally` block for the stream request:

```ts
loadingServices.current.delete(cacheKey)
findingsAbortControllers.current.delete(cacheKey)
if (activeFindingsKey.current === cacheKey) {
  setLoadingFindings(false)
}
```

- [ ] **Step 7: Abort all streams on cache clear**

Replace the single-controller clear logic with:

```ts
findingsAbortControllers.current.forEach((controller) => controller.abort())
findingsAbortControllers.current.clear()
findingsCache.current.clear()
loadingServices.current.clear()
activeFindingsKey.current = null
lastFindingsShellNodes.current = []
```

- [ ] **Step 8: Run focused tests**

Run:

```bash
cd frontend && npx --yes tsx --test src/pro/components/AGUIChat/components/EvidencePanel/findingsHydration.test.ts src/pro/api/conversationStream.test.ts src/pro/components/AGUIChat/components/FindingsTree/findingsNodePresentation.test.ts
```

Expected: PASS.

### Task 3: Verify Build Surface

**Files:**
- Verify: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/index.tsx`
- Verify: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/findingsHydration.ts`

- [ ] **Step 1: Run TypeScript build**

Run:

```bash
cd frontend && npm run build
```

Expected: build completes successfully.

- [ ] **Step 2: Review diff for unrelated edits**

Run:

```bash
git diff -- frontend/src/pro/components/AGUIChat/components/EvidencePanel/index.tsx frontend/src/pro/components/AGUIChat/components/EvidencePanel/findingsHydration.ts frontend/src/pro/components/AGUIChat/components/EvidencePanel/findingsHydration.test.ts docs/superpowers/plans/2026-05-11-findings-stream-cache-lifecycle.md
```

Expected: diff only covers findings stream lifecycle, cache helpers, tests, and this plan.

### Self-Review

- Spec coverage: the plan covers continued cache hydration after switching, no duplicate stream per service, bounded stream concurrency, terminal cache reuse, and cleanup on component unmount.
- Placeholder scan: no incomplete requirements remain.
- Type consistency: helper names and `FindingsHydrationPhase` values match the existing `findingsHydration.ts` contract.
