export const PROBE_MANAGEMENT_PATH = '/probe-management'

export const isProbeManagementPreviewRoute = () =>
  (import.meta.env.DEV || import.meta.env.VITE_PROBE_MANAGEMENT_PREVIEW === 'true') &&
  typeof window !== 'undefined' &&
  window.location.hash.startsWith(`#${PROBE_MANAGEMENT_PATH}`)
