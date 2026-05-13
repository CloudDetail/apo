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
  traceId: string
  spanId?: string
  url: string
}

const normalizeReportLinkType = (value: unknown): FindingsReportLinkType | undefined => {
  const normalized = readString(value)?.toLowerCase()

  if (normalized === 'slow' || normalized === 'error') {
    return normalized
  }

  return undefined
}

const buildFindingsReportUrl = ({
  type,
  traceId,
  spanId,
  url,
}: {
  type: FindingsReportLinkType
  traceId: string
  spanId?: string
  url?: string
}) => {
  const normalizedUrl = readString(url)
  if (normalizedUrl) {
    return normalizedUrl
  }

  if (type === 'slow' && spanId) {
    return `/#/cause/report/${encodeURIComponent(traceId)}/${encodeURIComponent(spanId)}`
  }

  if (type === 'error') {
    return `/#/cause/error-report/${encodeURIComponent(traceId)}`
  }

  return undefined
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
    const url = buildFindingsReportUrl({
      type,
      traceId,
      spanId,
      url: readString(item.url),
    })

    if (!url) {
      return
    }

    const dedupeKey = `${type}:${url}`
    if (seen.has(dedupeKey)) {
      return
    }
    seen.add(dedupeKey)

    links.push({
      type,
      traceId,
      url,
      ...(spanId ? { spanId } : {}),
    })
  })

  return links.length > 0 ? links : undefined
}
