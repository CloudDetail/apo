export const DEFAULT_KNOWLEDGE_BASE_ID = 'syncause_default'
export const RAG_KB_API_PREFIX = '/rag/api/kb'
export const MAX_KNOWLEDGE_BASE_FILE_SIZE = 20 * 1024 * 1024
export const MAX_KNOWLEDGE_BASE_FILE_COUNT = 5
export const KNOWLEDGE_BASE_FILE_ACCEPT = '.txt,.md,.docx,.csv,.xls,.xlsx'

export type KnowledgeBaseIndexStatus = 'pending' | 'indexing' | 'indexed' | 'failed'

export interface KnowledgeBaseSummary {
  id: string
  name: string
  description?: string
  status?: string
  writable?: boolean
}

export interface KnowledgeBaseList {
  items: KnowledgeBaseSummary[]
  total: number
  currentPage: number
  pageSize: number
}

export interface KnowledgeBaseDocument {
  id: string
  kbId: string
  filename: string
  mimeType?: string
  extension?: string
  sizeBytes?: number
  status?: string
  indexStatus?: KnowledgeBaseIndexStatus | string
  createdAt?: string
  errorMessage?: string
}

export interface KnowledgeBaseDocumentList {
  items: KnowledgeBaseDocument[]
  total: number
  limit: number
  offset: number
}

export interface DocumentListRequest {
  page?: number
  pageSize?: number
  keyword?: string
  status?: string
}

export interface UploadKnowledgeBaseDocumentResult {
  success?: boolean
  message?: string
  filename?: string
  kbId?: string
  docId?: string
  documentIds: string[]
  indexStatus?: KnowledgeBaseIndexStatus | string
  errorMessage?: string
}

export interface DocumentIndexStatusResult {
  success?: boolean
  indexStatus?: KnowledgeBaseIndexStatus | string
  errorMessage?: string
}

export interface IndexStatusDisplay {
  label: string
  color: 'default' | 'processing' | 'success' | 'error'
}

const SUPPORTED_FILE_EXTENSIONS = new Set(['.txt', '.md', '.docx', '.csv', '.xls', '.xlsx'])

export const buildKnowledgeBaseApiPath = (...segments: string[]) =>
  `${RAG_KB_API_PREFIX}/${segments.map((segment) => encodeURIComponent(segment)).join('/')}`

const isRecord = (value: unknown): value is Record<string, unknown> =>
  value !== null && typeof value === 'object' && !Array.isArray(value)

const readString = (value: unknown): string | undefined => {
  if (typeof value !== 'string') {
    return undefined
  }

  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : undefined
}

const readNumber = (value: unknown): number | undefined =>
  typeof value === 'number' && Number.isFinite(value) ? value : undefined

const readBoolean = (value: unknown): boolean | undefined =>
  typeof value === 'boolean' ? value : undefined

const normalizeIndexStatus = (value: unknown) => readString(value) as KnowledgeBaseIndexStatus | string | undefined

export const isKnowledgeBaseWritable = (knowledgeBase?: Pick<KnowledgeBaseSummary, 'id' | 'writable'>) => {
  if (!knowledgeBase) {
    return false
  }

  return knowledgeBase.writable === true
}

export const filterWritableKnowledgeBases = <T extends Pick<KnowledgeBaseSummary, 'writable'>>(items: T[]) =>
  items.filter((item) => item.writable === true)

const normalizeKnowledgeBase = (value: unknown): KnowledgeBaseSummary | null => {
  if (!isRecord(value)) {
    return null
  }

  const id = readString(value.id)
  if (!id) {
    return null
  }

  return {
    id,
    name: readString(value.name) ?? id,
    description: readString(value.description),
    status: readString(value.status),
    writable: readBoolean(value.writable),
  }
}

export const normalizeKnowledgeBaseListResponse = (value: unknown): KnowledgeBaseList => {
  const envelope = isRecord(value) ? value : {}
  const rawItems = Array.isArray(envelope.data)
    ? envelope.data
    : Array.isArray(envelope.items)
      ? envelope.items
      : []

  const items = rawItems
    .map((item) => normalizeKnowledgeBase(item))
    .filter((item): item is KnowledgeBaseSummary => item !== null)

  return {
    items,
    total: readNumber(envelope.total) ?? items.length,
    currentPage: readNumber(envelope.current_page) ?? readNumber(envelope.currentPage) ?? 1,
    pageSize: readNumber(envelope.page_size) ?? readNumber(envelope.pageSize) ?? items.length,
  }
}

const normalizeDocument = (value: unknown): KnowledgeBaseDocument | null => {
  if (!isRecord(value)) {
    return null
  }

  const id = readString(value.id)
  const filename = readString(value.filename)
  if (!id || !filename) {
    return null
  }

  return {
    id,
    kbId: readString(value.kb_id) ?? readString(value.kbId) ?? '',
    filename,
    mimeType: readString(value.mime_type) ?? readString(value.mimeType),
    extension: readString(value.extension),
    sizeBytes: readNumber(value.size_bytes) ?? readNumber(value.sizeBytes),
    status: readString(value.status),
    indexStatus: normalizeIndexStatus(value.index_status) ?? normalizeIndexStatus(value.indexStatus),
    createdAt: readString(value.created_at) ?? readString(value.createdAt),
    errorMessage: readString(value.error_message) ?? readString(value.errorMessage),
  }
}

export const normalizeDocumentListResponse = (value: unknown): KnowledgeBaseDocumentList => {
  const envelope = isRecord(value) ? value : {}
  const rawItems = Array.isArray(envelope.items) ? envelope.items : []
  const items = rawItems
    .map((item) => normalizeDocument(item))
    .filter((item): item is KnowledgeBaseDocument => item !== null)

  return {
    items,
    total: readNumber(envelope.total) ?? items.length,
    limit: readNumber(envelope.limit) ?? items.length,
    offset: readNumber(envelope.offset) ?? 0,
  }
}

export const buildDocumentListRequestParams = ({
  page = 1,
  pageSize = 20,
  keyword,
  status,
}: DocumentListRequest = {}) => {
  const params: Record<string, string | number> = {
    limit: pageSize,
    offset: Math.max(page - 1, 0) * pageSize,
  }

  const normalizedKeyword = keyword?.trim()
  if (normalizedKeyword) {
    params.filename = normalizedKeyword
  }

  const normalizedStatus = status?.trim()
  if (normalizedStatus) {
    params.status = normalizedStatus
  }

  return params
}

export const buildKnowledgeBaseDocumentDownloadUrl = (kbId: string, docId: string) =>
  buildKnowledgeBaseApiPath('knowledge_bases', kbId, 'documents', docId, 'download')

export const resolveKnowledgeBaseDownloadFilename = (contentDisposition: unknown, fallback: string) => {
  const disposition = readString(contentDisposition)
  if (!disposition) {
    return fallback
  }

  const encodedFilename = disposition.match(/filename\*=UTF-8''([^;]+)/i)?.[1]
  if (encodedFilename) {
    try {
      return decodeURIComponent(encodedFilename)
    } catch (_error) {
      return encodedFilename
    }
  }

  return disposition.match(/filename="?([^";]+)"?/i)?.[1] ?? fallback
}

export const getIndexStatusDisplay = (status?: string): IndexStatusDisplay => {
  switch (status) {
    case 'pending':
      return { label: '待处理', color: 'default' }
    case 'indexing':
      return { label: '处理中', color: 'processing' }
    case 'indexed':
      return { label: '可用', color: 'success' }
    case 'failed':
      return { label: '处理失败', color: 'error' }
    default:
      return { label: '未知', color: 'default' }
  }
}

export const isIndexStatusTerminal = (status?: string) => status === 'indexed' || status === 'failed'

const getFileExtension = (fileName: string) => {
  const dotIndex = fileName.lastIndexOf('.')
  return dotIndex >= 0 ? fileName.slice(dotIndex).toLowerCase() : ''
}

export const validateKnowledgeBaseFiles = (
  files: Array<Pick<File, 'name' | 'size'>>,
  {
    maxSize = MAX_KNOWLEDGE_BASE_FILE_SIZE,
    maxCount = MAX_KNOWLEDGE_BASE_FILE_COUNT,
  }: {
    maxSize?: number
    maxCount?: number
  } = {},
) => {
  if (files.length > maxCount) {
    return [`单次最多上传 ${maxCount} 个文件`]
  }

  const errors: string[] = []
  files.forEach((file) => {
    if (!SUPPORTED_FILE_EXTENSIONS.has(getFileExtension(file.name))) {
      errors.push(`${file.name} 格式不支持`)
    }

    if (file.size > maxSize) {
      errors.push(`${file.name} 超过 ${Math.round(maxSize / 1024 / 1024)}MB`)
    }
  })

  return errors
}

export const extractKnowledgeBaseErrorMessage = (error: unknown, fallback: string) => {
  const responseData = isRecord(error) && isRecord(error.response) ? error.response.data : undefined
  const data = isRecord(responseData) ? responseData : isRecord(error) ? error : undefined
  const message =
    readString(data?.error_message) ??
    readString(data?.error) ??
    readString(data?.message) ??
    (error instanceof Error ? error.message : undefined) ??
    fallback
  const failedStep = readString(data?.failed_step)

  return failedStep ? `${message}，失败步骤：${failedStep}` : message
}

export const buildDocumentFormData = (
  file: File,
  {
    chunkMethod = 'separator',
    chunkParams = { chunk_size: 512, chunk_overlap: 20 },
  }: {
    chunkMethod?: string
    chunkParams?: Record<string, unknown>
  } = {},
) => {
  const formData = new FormData()
  formData.append('file', file)
  formData.append('chunk_method', chunkMethod)
  formData.append('chunk_params', JSON.stringify(chunkParams))

  return formData
}

export const normalizeUploadResult = (value: unknown): UploadKnowledgeBaseDocumentResult => {
  const data = isRecord(value) ? value : {}

  return {
    success: readBoolean(data.success),
    message: readString(data.message),
    filename: readString(data.filename),
    kbId: readString(data.kb_id) ?? readString(data.kbId),
    docId: readString(data.doc_id) ?? readString(data.docId),
    documentIds: Array.isArray(data.document_ids)
      ? data.document_ids.map(readString).filter((item): item is string => Boolean(item))
      : [],
    indexStatus: normalizeIndexStatus(data.index_status) ?? normalizeIndexStatus(data.indexStatus),
    errorMessage: readString(data.error_message) ?? readString(data.errorMessage),
  }
}
