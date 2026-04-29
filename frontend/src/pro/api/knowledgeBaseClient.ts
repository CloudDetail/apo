import { del, get, post } from 'src/core/utils/request'
import {
  RAG_KB_API_PREFIX,
  buildKnowledgeBaseApiPath,
  buildDocumentFormData,
  buildKnowledgeBaseDocumentDownloadUrl,
  buildDocumentListRequestParams,
  normalizeDocumentListResponse,
  normalizeKnowledgeBaseListResponse,
  normalizeUploadResult,
  type DocumentIndexStatusResult,
  type DocumentListRequest,
  type KnowledgeBaseSummary,
} from './knowledgeBase'

export interface DeleteKnowledgeBaseDocumentResult {
  message?: string
}

const normalizeKnowledgeBase = (value: unknown): KnowledgeBaseSummary => {
  const normalized = normalizeKnowledgeBaseListResponse({ data: [value] }).items[0]
  if (!normalized) {
    throw new Error('知识库响应格式不正确')
  }
  return normalized
}

export const getKnowledgeBasesApi = async (params?: Record<string, unknown>) =>
  normalizeKnowledgeBaseListResponse(await get(`${RAG_KB_API_PREFIX}/knowledge_bases`, params))

export const getKnowledgeBaseApi = async (kbId: string) =>
  normalizeKnowledgeBase(await get(buildKnowledgeBaseApiPath('knowledge_bases', kbId)))

export const getKnowledgeBaseDocumentsApi = async (kbId: string, params?: DocumentListRequest) =>
  normalizeDocumentListResponse(
    await get(buildKnowledgeBaseApiPath('knowledge_bases', kbId, 'documents'), buildDocumentListRequestParams(params)),
  )

export const previewKnowledgeBaseDocumentApi = async (kbId: string, file: File) =>
  post(buildKnowledgeBaseApiPath('knowledge_bases', kbId, 'documents', 'preview'), buildDocumentFormData(file), {
    headers: { 'Content-Type': 'multipart/form-data' },
  })

export const uploadKnowledgeBaseDocumentApi = async (kbId: string, file: File) =>
  normalizeUploadResult(
    await post(buildKnowledgeBaseApiPath('knowledge_bases', kbId, 'documents'), buildDocumentFormData(file), {
      headers: { 'Content-Type': 'multipart/form-data' },
    }),
  )

export const getKnowledgeBaseDocumentIndexStatusApi = async (
  docId: string,
): Promise<DocumentIndexStatusResult> => {
  const data = await get(buildKnowledgeBaseApiPath('documents', docId, 'index_status'))
  const record = data && typeof data === 'object' && !Array.isArray(data) ? data as Record<string, unknown> : {}

  return {
    success: typeof record.success === 'boolean' ? record.success : undefined,
    indexStatus:
      typeof record.index_status === 'string'
        ? record.index_status
        : typeof record.indexStatus === 'string'
          ? record.indexStatus
          : undefined,
    errorMessage:
      typeof record.error_message === 'string'
        ? record.error_message
        : typeof record.errorMessage === 'string'
          ? record.errorMessage
          : undefined,
  }
}

export const deleteKnowledgeBaseDocumentApi = async (
  kbId: string,
  docId: string,
): Promise<DeleteKnowledgeBaseDocumentResult> =>
  del(buildKnowledgeBaseApiPath('knowledge_bases', kbId, 'documents', docId))

export const downloadKnowledgeBaseDocumentApi = async (kbId: string, docId: string) =>
  get(buildKnowledgeBaseDocumentDownloadUrl(kbId, docId), {}, { responseType: 'blob' })
