export const SYSTEM_HEALTH_PATH = '/healthy-service'
export const SYSTEM_HEALTH_LABEL = '系统健康评分'
export const SYSTEM_HEALTH_ABBREVIATION = '健康评分'

export const isSystemHealthPreviewRoute = () =>
  import.meta.env.DEV &&
  typeof window !== 'undefined' &&
  window.location.hash.startsWith(`#${SYSTEM_HEALTH_PATH}`)

export const normalizeSystemHealthMenu = (items = []) =>
  items.map((item) => {
    const to = item.router?.to || item.to
    const isSystemHealthItem = item.key === 'healthy' || to === SYSTEM_HEALTH_PATH
    const children = Array.isArray(item.children)
      ? normalizeSystemHealthMenu(item.children)
      : item.children

    return {
      ...item,
      ...(isSystemHealthItem
        ? {
            label: SYSTEM_HEALTH_LABEL,
            abbreviation: SYSTEM_HEALTH_ABBREVIATION,
          }
        : {}),
      ...(children ? { children } : {}),
    }
  })
