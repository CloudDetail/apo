export type PreloadRuntimeId = 'java' | 'python' | 'nodejs' | 'dotnet'

export type PreloadRuntime = {
  id: PreloadRuntimeId
  name: string
  injection: string
  timing: string
}

export type PreloadConfigState = {
  enabled: Record<PreloadRuntimeId, boolean>
  issues: string[]
}

type LineRange = {
  start: number
  end: number
  lines: string[]
}

const DISABLED_BEGIN = 'APO_INJECTION_DISABLED_BEGIN'
const DISABLED_CONTENT = 'APO_INJECTION_DISABLED_CONTENT'
const DISABLED_END = 'APO_INJECTION_DISABLED_END'

export const PRELOAD_RUNTIMES: PreloadRuntime[] = [
  { id: 'java', name: 'Java', injection: 'javaagent', timing: '新启动 JVM' },
  { id: 'python', name: 'Python', injection: 'PYTHONPATH', timing: '新启动解释器' },
  { id: 'nodejs', name: 'Node.js', injection: 'NODE_OPTIONS', timing: '新启动进程' },
  { id: 'dotnet', name: '.NET', injection: 'CLR Profiler', timing: '新启动进程' },
]

const RUNTIME_IDS = new Set<PreloadRuntimeId>(PRELOAD_RUNTIMES.map((runtime) => runtime.id))

const DEFAULT_SECTIONS: Record<PreloadRuntimeId, string> = {
  java: `[java]
OTEL_SERVICE_NAME={{APO_AUTO_SERVICE_NAME}}
OTEL_EXPORTER_OTLP_ENDPOINT={{APO_COLLECTOR_GRPC_ENDPOINT}}
JAVA_TOOL_OPTIONS=-javaagent:/etc/apo/instrumentations/java/javaagent.jar
JAVA_OPTS=-javaagent:/etc/apo/instrumentations/java/javaagent.jar
OTEL_EXPORTER_OTLP_PROTOCOL=grpc
OTEL_TRACES_SAMPLER=always_on
OTEL_TRACES_EXPORTER=otlp
OTEL_METRICS_EXPORTER=otlp
OTEL_LOGS_EXPORTER=none`,
  python: `[python]
OTEL_SERVICE_NAME = {{APO_AUTO_SERVICE_NAME}}
OTEL_EXPORTER_OTLP_ENDPOINT = {{APO_COLLECTOR_HTTP_ENDPOINT}}
PYTHONPATH = /etc/apo/instrumentations/python:/etc/apo/instrumentations/python/opentelemetry/instrumentation/auto_instrumentation
OTEL_TRACES_EXPORTER = otlp
OTEL_METRICS_EXPORTER = otlp
OTEL_EXPORTER_OTLP_TRACES_PROTOCOL = http/protobuf
OTEL_EXPORTER_OTLP_METRICS_PROTOCOL = http/protobuf
OTEL_PYTHON_LOG_CORRELATION = true`,
  nodejs: `[nodejs]
OTEL_SERVICE_NAME = {{APO_AUTO_SERVICE_NAME}}
OTEL_EXPORTER_OTEL_ENDPOINT = {{APO_COLLECTOR_GRPC_ENDPOINT}}
NODE_OPTIONS = --require /etc/apo/instrumentations/nodejs/autoinstrumentation.js`,
  dotnet: `[dotnet]
OTEL_SERVICE_NAME = {{APO_AUTO_SERVICE_NAME}}
OTEL_EXPORTER_OTLP_ENDPOINT = {{APO_COLLECTOR_HTTP_ENDPOINT}}
CORECLR_PROFILER = {918728DD-259F-4A6A-AC2B-B85E1B658318}
CORECLR_ENABLE_PROFILING = 1
CORECLR_PROFILER_PATH = /etc/apo/instrumentations/dotnet/OpenTelemetry.AutoInstrumentation.ClrProfiler.Native.so
OTEL_TRACES_EXPORTER = otlp
DOTNET_STARTUP_HOOKS = /etc/apo/instrumentations/dotnet/net/OpenTelemetry.AutoInstrumentation.StartupHook.dll
DOTNET_ADDITIONAL_DEPS = /etc/apo/instrumentations/dotnet/AdditionalDeps
DOTNET_SHARED_STORE = /etc/apo/instrumentations/dotnet/store`,
}

const runtimeId = (value: string): PreloadRuntimeId | undefined => {
  const normalized = value.trim().toLowerCase() as PreloadRuntimeId
  return RUNTIME_IDS.has(normalized) ? normalized : undefined
}

const linesAndEnding = (content: string) => ({
  lines: content.split(/\r?\n/),
  ending: content.includes('\r\n') ? '\r\n' : '\n',
})

const inspectPreloadConfig = (content: string) => {
  const { lines } = linesAndEnding(content)
  const active = new Map<PreloadRuntimeId, LineRange[]>()
  const disabled = new Map<PreloadRuntimeId, LineRange[]>()
  const issues: string[] = []
  const sectionHeaders: Array<{ name: string; start: number }> = []

  for (let index = 0; index < lines.length; index += 1) {
    const disabledMatch = lines[index].match(
      /^\s*;\s*APO_INJECTION_DISABLED_BEGIN\s+(java|python|nodejs|dotnet)\s*$/i,
    )
    if (disabledMatch) {
      const id = runtimeId(disabledMatch[1])!
      const restored: string[] = []
      let cursor = index + 1
      for (; cursor < lines.length; cursor += 1) {
        if (new RegExp(`^\\s*;\\s*${DISABLED_END}\\s+${id}\\s*$`, 'i').test(lines[cursor])) {
          break
        }
        const contentMatch = lines[cursor].match(
          /^\s*;\s*APO_INJECTION_DISABLED_CONTENT(?: (.*))?$/,
        )
        if (!contentMatch) {
          issues.push(`${id} 的停用配置块已损坏`)
          continue
        }
        restored.push(contentMatch[1] || '')
      }
      if (cursor >= lines.length) {
        issues.push(`${id} 的停用配置块缺少结束标记`)
        break
      }
      const ranges = disabled.get(id) || []
      ranges.push({ start: index, end: cursor + 1, lines: restored })
      disabled.set(id, ranges)
      index = cursor
      continue
    }

    const sectionMatch = lines[index].match(/^\s*\[([^\]]+)\]\s*$/)
    if (sectionMatch) sectionHeaders.push({ name: sectionMatch[1].trim(), start: index })
  }

  sectionHeaders.forEach((header, index) => {
    const id = runtimeId(header.name)
    if (!id) return
    const ranges = active.get(id) || []
    ranges.push({
      start: header.start,
      end: sectionHeaders[index + 1]?.start ?? lines.length,
      lines: lines.slice(header.start, sectionHeaders[index + 1]?.start ?? lines.length),
    })
    active.set(id, ranges)
  })

  PRELOAD_RUNTIMES.forEach(({ id }) => {
    const activeCount = active.get(id)?.length || 0
    const disabledCount = disabled.get(id)?.length || 0
    if (activeCount > 1) issues.push(`${id} 存在重复 section`)
    if (disabledCount > 1) issues.push(`${id} 存在重复停用配置块`)
    if (activeCount && disabledCount) issues.push(`${id} 同时存在启用和停用配置`)
  })

  return { lines, active, disabled, sectionHeaders, issues }
}

export const parsePreloadConfig = (content: string): PreloadConfigState => {
  const inspected = inspectPreloadConfig(content)
  return {
    enabled: Object.fromEntries(
      PRELOAD_RUNTIMES.map(({ id }) => [id, Boolean(inspected.active.get(id)?.length)]),
    ) as Record<PreloadRuntimeId, boolean>,
    issues: inspected.issues,
  }
}

export const getPreloadRuntimeSection = (content: string, id: PreloadRuntimeId) => {
  const inspected = inspectPreloadConfig(content)
  const range = inspected.active.get(id)?.[0] || inspected.disabled.get(id)?.[0]
  return {
    content: (range?.lines || DEFAULT_SECTIONS[id].split('\n')).join(
      linesAndEnding(content).ending,
    ),
    issues: inspected.issues,
  }
}

const disabledLines = (id: PreloadRuntimeId, lines: string[]) => [
  `; ${DISABLED_BEGIN} ${id}`,
  ...lines.map((line) => `; ${DISABLED_CONTENT}${line ? ` ${line}` : ''}`),
  `; ${DISABLED_END} ${id}`,
]

export const setPreloadRuntimeEnabled = (
  content: string,
  id: PreloadRuntimeId,
  enabled: boolean,
): { content: string; issues: string[] } => {
  const inspected = inspectPreloadConfig(content)
  if (inspected.issues.length) return { content, issues: inspected.issues }

  const { ending } = linesAndEnding(content)
  const active = inspected.active.get(id)?.[0]
  const disabled = inspected.disabled.get(id)?.[0]
  if ((enabled && active) || (!enabled && !active)) return { content, issues: [] }

  const next = [...inspected.lines]
  if (!enabled && active) {
    next.splice(active.start, active.end - active.start, ...disabledLines(id, active.lines))
  } else if (disabled) {
    next.splice(disabled.start, disabled.end - disabled.start, ...disabled.lines)
  } else {
    const blacklistIndex = inspected.sectionHeaders.find(({ name }) =>
      ['basename-blacklist', 'cmdline-blacklist', 'env-blacklist'].includes(name.toLowerCase()),
    )?.start
    const insertion = DEFAULT_SECTIONS[id].split('\n')
    const insertAt = blacklistIndex ?? Math.max(0, next.length - (next.at(-1) === '' ? 1 : 0))
    if (insertAt > 0 && next[insertAt - 1] !== '') insertion.unshift('')
    if (next[insertAt] !== '') insertion.push('')
    next.splice(insertAt, 0, ...insertion)
  }

  return { content: next.join(ending), issues: [] }
}
