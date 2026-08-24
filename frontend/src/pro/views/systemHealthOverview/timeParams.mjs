const safeText = (value) => (value === undefined || value === null ? '' : `${value}`.trim())

const positiveNumber = (value) => {
  const number = Number(value)
  return Number.isFinite(number) && number > 0 ? number : 0
}

const isoToTimestamp = (isoString) => {
  if (!/^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}\.\d{6}Z$/.test(isoString || '')) {
    return null
  }
  const datePart = isoString.slice(0, -4) + 'Z'
  const date = new Date(datePart)
  if (Number.isNaN(date.getTime())) return null
  const microseconds = Number.parseInt(isoString.slice(-4, -1), 10)
  return date.getTime() * 1000 + microseconds
}

export const normalizeUrlTimeParams = (searchParams) => {
  const relativeTime = searchParams.get('relativeTime')
  const startTime =
    safeText(searchParams.get('startTime')) || isoToTimestamp(searchParams.get('from'))
  const endTime = safeText(searchParams.get('endTime')) || isoToTimestamp(searchParams.get('to'))

  if (positiveNumber(startTime) > 0 && positiveNumber(endTime) > positiveNumber(startTime)) {
    return {
      startTime,
      endTime,
    }
  }

  if (relativeTime) {
    return { relativeTime }
  }

  return null
}

export const buildStoreTimeParams = (timeRange = {}) => {
  const storeStartTime = positiveNumber(timeRange.startTime)
  const storeEndTime = positiveNumber(timeRange.endTime)

  if (storeStartTime > 0 && storeEndTime > storeStartTime) {
    return { startTime: storeStartTime, endTime: storeEndTime }
  }
  const rangeTypeKey = safeText(timeRange.rangeTypeKey)
  if (rangeTypeKey) return { relativeTime: rangeTypeKey }
  return null
}

export const makeTimeRangeSignature = (timeRange = {}) =>
  [
    safeText(timeRange.rangeTypeKey),
    positiveNumber(timeRange.startTime) || '',
    positiveNumber(timeRange.endTime) || '',
  ].join(':')

export const buildOverviewRequestParams = (search, timeRange = {}, options = {}) => {
  const urlTimeParams = normalizeUrlTimeParams(new URLSearchParams(search))
  const storeTimeParams = buildStoreTimeParams(timeRange)
  const requireResolvedTimeRange = Boolean(options.requireResolvedTimeRange)
  const isTimeRangeRehydrating = Boolean(
    timeRange?._persist && timeRange._persist.rehydrated === false,
  )
  const hasStoreConcreteTime = Boolean(storeTimeParams?.startTime && storeTimeParams?.endTime)
  const hasUrlConcreteTime = Boolean(urlTimeParams?.startTime && urlTimeParams?.endTime)

  if (requireResolvedTimeRange && isTimeRangeRehydrating) return null
  if (!options.preferStore && hasUrlConcreteTime) {
    return urlTimeParams
  }
  if (storeTimeParams && (!requireResolvedTimeRange || hasStoreConcreteTime)) return storeTimeParams
  if (urlTimeParams && (!requireResolvedTimeRange || hasUrlConcreteTime)) return urlTimeParams
  if (requireResolvedTimeRange) return null
  return { relativeTime: '15m' }
}

export const buildOverviewRequestKey = (params) => {
  if (!params) return ''
  return JSON.stringify(
    Object.entries(params)
      .filter(([, value]) => value !== undefined && value !== null && `${value}` !== '')
      .sort(([left], [right]) => left.localeCompare(right)),
  )
}
