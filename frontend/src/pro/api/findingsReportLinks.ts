const isObject = (value: unknown): value is Record<string, unknown> =>
  value !== null && typeof value === 'object' && !Array.isArray(value)

const readString = (value: unknown): string | undefined => {
  if (typeof value !== 'string') {
    return undefined
  }

  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : undefined
}

export type FindingsReportLinkType = 'slow' | 'error'

export interface FindingsReportLink {
  type: FindingsReportLinkType
  title?: string
  serviceName?: string
  errorGroupKey?: string
  representative?: boolean
  groupCount?: number
  traceId: string
  spanId?: string
  mutatedType?: string
  url: string
  source?: string
}

const normalizeReportLinkType = (value: unknown): FindingsReportLinkType | undefined => {
  const normalized = readString(value)?.toLowerCase()

  if (normalized === 'slow' || normalized === 'error') {
    return normalized
  }

  return undefined
}

const buildExpectedFindingsReportUrl = ({
  type,
  traceId,
  spanId,
}: {
  type: FindingsReportLinkType
  traceId: string
  spanId?: string
}) => {
  if (type === 'slow' && spanId) {
    return `/#/cause/report/${encodeURIComponent(traceId)}/${encodeURIComponent(spanId)}`
  }

  if (type === 'error') {
    return `/#/cause/error-report/${encodeURIComponent(traceId)}`
  }

  return undefined
}

const resolveSlowReportUrl = ({
  traceId,
  spanId,
  mutatedType,
  url,
}: {
  traceId: string
  spanId: string
  mutatedType?: string
  url: string
}) => {
  const expectedBaseUrl = buildExpectedFindingsReportUrl({
    type: 'slow',
    traceId,
    spanId,
  })
  const queryStart = url.indexOf('?')
  const path = queryStart >= 0 ? url.slice(0, queryStart) : url
  const query = queryStart >= 0 ? url.slice(queryStart + 1) : undefined

  if (path !== expectedBaseUrl) {
    return undefined
  }

  if (query === undefined) {
    return mutatedType ? undefined : { url }
  }

  const params = new URLSearchParams(query)
  const entries = Array.from(params.entries())
  if (entries.length !== 1 || entries[0][0] !== 'mutatedType') {
    return undefined
  }

  const mutatedTypeFromUrl = readString(entries[0][1])
  if (!mutatedTypeFromUrl || (mutatedType && mutatedType !== mutatedTypeFromUrl)) {
    return undefined
  }

  if (query !== `mutatedType=${encodeURIComponent(mutatedTypeFromUrl)}`) {
    return undefined
  }

  return {
    url,
    mutatedType: mutatedType ?? mutatedTypeFromUrl,
  }
}

const resolveFindingsReportUrl = ({
  type,
  traceId,
  spanId,
  mutatedType,
  url,
}: {
  type: FindingsReportLinkType
  traceId: string
  spanId?: string
  mutatedType?: string
  url?: string
}) => {
  const normalizedUrl = readString(url)

  if (!normalizedUrl) {
    return undefined
  }

  if (type === 'slow' && spanId) {
    return resolveSlowReportUrl({
      traceId,
      spanId,
      mutatedType,
      url: normalizedUrl,
    })
  }

  const expectedUrl = buildExpectedFindingsReportUrl({
    type,
    traceId,
    spanId,
  })

  if (!expectedUrl) {
    return undefined
  }

  return normalizedUrl === expectedUrl ? { url: normalizedUrl } : undefined
}

export const normalizeFindingsReportLinks = (value: unknown): FindingsReportLink[] | undefined => {
  if (!Array.isArray(value)) {
    return undefined
  }

  const seen = new Set<string>()
  const links: FindingsReportLink[] = []

  value.filter(isObject).forEach((item) => {
    const type = normalizeReportLinkType(item.type)
    const traceId = readString(item.traceId)

    if (!type || !traceId) {
      return
    }

    const spanId = readString(item.spanId)
    const mutatedType = readString(item.mutatedType)
    const title = readString(item.title)
    const serviceName = readString(item.serviceName)
    const errorGroupKey = readString(item.errorGroupKey)
    const representative = typeof item.representative === 'boolean' ? item.representative : undefined
    const groupCount =
      typeof item.groupCount === 'number' && Number.isFinite(item.groupCount) && item.groupCount > 0
        ? item.groupCount
        : undefined
    const source = readString(item.source)
    const resolvedUrl = resolveFindingsReportUrl({
      type,
      traceId,
      spanId,
      mutatedType,
      url: readString(item.url),
    })

    if (!resolvedUrl) {
      return
    }

    const dedupeKey = `${type}:${resolvedUrl.url}`
    if (seen.has(dedupeKey)) {
      return
    }
    seen.add(dedupeKey)

    links.push({
      type,
      ...(title ? { title } : {}),
      ...(serviceName ? { serviceName } : {}),
      ...(errorGroupKey ? { errorGroupKey } : {}),
      ...(representative !== undefined ? { representative } : {}),
      ...(groupCount ? { groupCount } : {}),
      traceId,
      ...(spanId ? { spanId } : {}),
      ...(resolvedUrl.mutatedType ? { mutatedType: resolvedUrl.mutatedType } : {}),
      url: resolvedUrl.url,
      ...(source ? { source } : {}),
    })
  })

  links.sort((left, right) => (left.type === right.type ? 0 : left.type === 'slow' ? -1 : 1))

  return links.length > 0 ? links : undefined
}
