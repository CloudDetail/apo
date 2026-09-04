export const dataIntegrationSaveMode = (formInitValues?: { id?: unknown } | null) =>
  formInitValues?.id ? 'update' : 'create'
