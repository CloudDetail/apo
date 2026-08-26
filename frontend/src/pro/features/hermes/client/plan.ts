import type { HermesPlan } from './types.ts'

const PLAN_TOOL_NAME = 'task_flow_orchestration_nodes_upsert'

export function isPlanToolPayload(payload: Record<string, unknown>): boolean {
  const name = String(payload.name ?? payload.tool ?? payload.tool_name ?? '').trim()
  return name.split('__').at(-1) === PLAN_TOOL_NAME
}

export function planFromToolPayload(payload: Record<string, unknown>): HermesPlan | null {
  if (!isPlanToolPayload(payload)) return null
  const args = objectValue(payload.args ?? payload.arguments)
  const data = objectValue(args?.data)
  const core = objectValue(data?.core ?? args?.core)
  const rawNodes = core?.nodes ?? args?.nodes
  if (!Array.isArray(rawNodes)) return null

  const steps = rawNodes.flatMap((value, index) => {
    const node = objectValue(value)
    const title = String(node?.title ?? node?.label ?? '').trim()
    if (!title) return []
    return [{ id: String(node?.id ?? `step-${index + 1}`), title }]
  })
  return steps.length ? { steps } : null
}

export function planFingerprint(plan: HermesPlan): string {
  return JSON.stringify(plan.steps.map((step) => [step.id, step.title]))
}

function objectValue(value: unknown): Record<string, any> | null {
  return value && typeof value === 'object' && !Array.isArray(value)
    ? (value as Record<string, any>)
    : null
}
