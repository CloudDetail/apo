import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import {
  DEFAULT_KNOWLEDGE_BASE_ID,
  buildKnowledgeBaseApiPath,
  buildKnowledgeBaseDocumentDownloadUrl,
  buildDocumentListRequestParams,
  extractKnowledgeBaseErrorMessage,
  filterWritableKnowledgeBases,
  getIndexStatusDisplay,
  isIndexStatusTerminal,
  isKnowledgeBaseWritable,
  normalizeDocumentListResponse,
  normalizeKnowledgeBaseListResponse,
  resolveKnowledgeBaseDownloadFilename,
  validateKnowledgeBaseFiles,
} from './knowledgeBase'

const makeFile = (name: string, size: number) => ({ name, size }) as File

test('normalizes knowledge base list responses and honors explicit writable flag only', () => {
  const normalized = normalizeKnowledgeBaseListResponse({
    total: 3,
    current_page: 1,
    page_size: 10,
    data: [
      {
        id: 'support',
        name: 'Support',
        description: 'Product docs',
        status: 'active',
        writable: false,
      },
      {
        id: DEFAULT_KNOWLEDGE_BASE_ID,
        name: 'Ops',
        description: 'Runbooks',
        status: 'active',
      },
    ],
  })

  assert.equal(normalized.total, 3)
  assert.deepEqual(normalized.items.map((item) => item.id), ['support', DEFAULT_KNOWLEDGE_BASE_ID])
  assert.equal(isKnowledgeBaseWritable(normalized.items[0]), false)
  assert.equal(isKnowledgeBaseWritable(normalized.items[1]), false)
})

test('filters knowledge base options to writable true only', () => {
  const visible = filterWritableKnowledgeBases([
    { id: DEFAULT_KNOWLEDGE_BASE_ID, name: 'Ops' },
    { id: 'support', name: 'Support', writable: false },
    { id: 'ops_team', name: 'Ops Team', writable: true },
  ])

  assert.deepEqual(visible.map((item) => item.id), ['ops_team'])
})

test('normalizes document list fields and preserves server pagination metadata', () => {
  const normalized = normalizeDocumentListResponse({
    items: [
      {
        id: 'doc_001',
        kb_id: DEFAULT_KNOWLEDGE_BASE_ID,
        filename: 'CPU高排查Runbook.md',
        mime_type: 'text/markdown',
        extension: '.md',
        size_bytes: 10240,
        status: 'active',
        index_status: 'indexed',
        created_at: '2026-04-28T10:00:00',
        error_message: null,
      },
    ],
    total: 1,
    limit: 20,
    offset: 0,
  })

  assert.deepEqual(normalized, {
    items: [
      {
        id: 'doc_001',
        kbId: DEFAULT_KNOWLEDGE_BASE_ID,
        filename: 'CPU高排查Runbook.md',
        mimeType: 'text/markdown',
        extension: '.md',
        sizeBytes: 10240,
        status: 'active',
        indexStatus: 'indexed',
        createdAt: '2026-04-28T10:00:00',
        errorMessage: undefined,
      },
    ],
    total: 1,
    limit: 20,
    offset: 0,
  })
})

test('builds document list query params for server-side filtering', () => {
  const params = buildDocumentListRequestParams({
    page: 3,
    pageSize: 20,
    keyword: ' cpu ',
    status: 'active',
  })

  assert.deepEqual(params, {
    limit: 20,
    offset: 40,
    filename: 'cpu',
    status: 'active',
  })
})

test('builds original document download url through rag proxy', () => {
  assert.equal(
    buildKnowledgeBaseDocumentDownloadUrl(DEFAULT_KNOWLEDGE_BASE_ID, 'doc_001'),
    '/rag/api/kb/knowledge_bases/syncause_default/documents/doc_001/download',
  )

  assert.equal(
    buildKnowledgeBaseDocumentDownloadUrl('ops kb', 'doc/001'),
    '/rag/api/kb/knowledge_bases/ops%20kb/documents/doc%2F001/download',
  )
})

test('builds encoded rag api paths from path segments', () => {
  assert.equal(
    buildKnowledgeBaseApiPath('knowledge_bases', 'ops kb', 'documents', 'doc/001'),
    '/rag/api/kb/knowledge_bases/ops%20kb/documents/doc%2F001',
  )
})

test('resolves original document download filename from content disposition', () => {
  assert.equal(
    resolveKnowledgeBaseDownloadFilename("attachment; filename*=UTF-8''CPU%E9%AB%98.md", 'fallback.md'),
    'CPU高.md',
  )
  assert.equal(resolveKnowledgeBaseDownloadFilename('attachment; filename="runbook.md"', 'fallback.md'), 'runbook.md')
  assert.equal(resolveKnowledgeBaseDownloadFilename(undefined, 'fallback.md'), 'fallback.md')
})

test('maps index status to user-facing import status display and terminal states', () => {
  assert.deepEqual(getIndexStatusDisplay('pending'), { label: '待处理', color: 'default' })
  assert.deepEqual(getIndexStatusDisplay('indexing'), { label: '处理中', color: 'processing' })
  assert.deepEqual(getIndexStatusDisplay('indexed'), { label: '可用', color: 'success' })
  assert.deepEqual(getIndexStatusDisplay('failed'), { label: '处理失败', color: 'error' })
  assert.equal(isIndexStatusTerminal('pending'), false)
  assert.equal(isIndexStatusTerminal('indexing'), false)
  assert.equal(isIndexStatusTerminal('indexed'), true)
  assert.equal(isIndexStatusTerminal('failed'), true)
})

test('validates supported upload formats, size, and count', () => {
  assert.deepEqual(validateKnowledgeBaseFiles([makeFile('runbook.md', 1024)]), [])

  assert.deepEqual(validateKnowledgeBaseFiles([makeFile('runbook.pdf', 1024)]), [
    'runbook.pdf 格式不支持',
  ])

  assert.deepEqual(validateKnowledgeBaseFiles([makeFile('large.md', 21 * 1024 * 1024)]), [
    'large.md 超过 20MB',
  ])

  assert.deepEqual(
    validateKnowledgeBaseFiles([
      makeFile('1.md', 1),
      makeFile('2.md', 1),
      makeFile('3.md', 1),
      makeFile('4.md', 1),
      makeFile('5.md', 1),
      makeFile('6.md', 1),
    ]),
    ['单次最多上传 5 个文件'],
  )
})

test('extracts rag error messages with failed step detail', () => {
  assert.equal(
    extractKnowledgeBaseErrorMessage(
      {
        response: {
          data: {
            error: '删除文档失败',
            failed_step: 'vector_index',
          },
        },
      },
      '删除失败',
    ),
    '删除文档失败，失败步骤：vector_index',
  )
})

test('knowledge base page keeps filters in one toolbar without title description', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')
  const toolbarStart = source.indexOf('<div className={styles.toolbar}>')
  const documentErrorStart = source.indexOf('{documentError &&', toolbarStart)
  const toolbarSource = source.slice(toolbarStart, documentErrorStart)
  const headerSource = source.slice(source.indexOf('<div className={styles.header}>'), toolbarStart)

  assert.equal(source.includes('<Text type="secondary">'), false)
  assert.equal(headerSource.includes('selectedKnowledgeBase?.description'), false)
  assert.match(toolbarSource, /<Select[\s\S]*value={selectedKbId}/)
  assert.match(toolbarSource, /<Input\.Search/)
  assert.match(toolbarSource, /placeholder="文档状态"/)
  assert.equal(toolbarSource.includes('<Button icon={<RefreshCw size={16} />} onClick={handleRefresh}>'), true)
  assert.match(toolbarSource, /上传文档/)
})

test('knowledge base upload dragger centers the icon and prompt text', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')
  const stylesSource = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.module.scss'), 'utf8')
  const uploadDraggerStart = source.indexOf('<Upload.Dragger')
  const uploadActionsStart = source.indexOf('<div className={styles.uploadActions}>', uploadDraggerStart)
  const uploadDraggerSource = source.slice(uploadDraggerStart, uploadActionsStart)

  assert.match(uploadDraggerSource, /className={styles\.uploadDragger}/)
  assert.match(stylesSource, /\.uploadDragger\s*{[\s\S]*text-align:\s*center;/)
  assert.match(stylesSource, /:global\(\.ant-upload-drag-icon\)\s*{[\s\S]*display:\s*flex;/)
  assert.match(stylesSource, /:global\(\.ant-upload-drag-icon\)\s*{[\s\S]*justify-content:\s*center;/)
  assert.match(stylesSource, /:global\(\.ant-upload-text\)\s*{[\s\S]*text-align:\s*center;/)
})

test('knowledge base page only shows writable knowledge bases', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')

  assert.match(source, /useState\(''\)/)
  assert.match(source, /filterWritableKnowledgeBases\(result\.items\)/)
  assert.match(source, /setKnowledgeBases\(writableItems\)/)
  assert.equal(source.includes('setKnowledgeBases(result.items)'), false)
})

test('knowledge base page displays knowledge base names without ids', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')
  const knowledgeBaseColumnStart = source.indexOf("title: '知识库'")
  const nextColumnStart = source.indexOf("title: '文件类型'", knowledgeBaseColumnStart)
  const knowledgeBaseColumnSource = source.slice(knowledgeBaseColumnStart, nextColumnStart)

  assert.match(source, /const getKnowledgeBaseLabel = \(item: KnowledgeBaseSummary\) => item\.name \|\| '-'/)
  assert.match(source, /knowledgeBaseNameById\.get\(value\) \?\? '-'/)
  assert.equal(source.includes('`${item.name} (${item.id})`'), false)
  assert.equal(knowledgeBaseColumnSource.includes("dataIndex: 'kbId'"), true)
  assert.equal(knowledgeBaseColumnSource.includes('render: (value?: string)'), true)
})

test('knowledge base page refreshes import status from status column only', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')
  const statusColumnStart = source.indexOf("title: '导入状态'")
  const nextColumnStart = source.indexOf("title: '上传时间'", statusColumnStart)
  const statusColumnSource = source.slice(statusColumnStart, nextColumnStart)
  const actionStart = source.indexOf("title: '操作'")
  const actionEnd = source.indexOf(']', actionStart)
  const actionSource = source.slice(actionStart, actionEnd)

  assert.match(statusColumnSource, /icon=\{<RefreshCw size=\{14\} \/>\}/)
  assert.match(statusColumnSource, /onClick=\{\(\) => record\?\.id && pollIndexStatus\(record\.id\)\}/)
  assert.equal(actionSource.includes('状态'), false)
  assert.equal(actionSource.includes('pollIndexStatus'), false)
})

test('knowledge base page downloads original document through handled request action', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')
  const actionStart = source.indexOf("title: '操作'")
  const actionEnd = source.indexOf(']', actionStart)
  const actionSource = source.slice(actionStart, actionEnd)

  assert.match(source, /downloadKnowledgeBaseDocumentApi\(selectedKbId, record\.id\)/)
  assert.match(source, /window\.setTimeout\(\(\) => window\.URL\.revokeObjectURL\(url\)/)
  assert.match(actionSource, /onClick=\{\(\) => handleDownload\(record\)\}/)
  assert.equal(actionSource.includes('download={record.filename}'), false)
  assert.equal(actionSource.includes('href={buildKnowledgeBaseDocumentDownloadUrl'), false)
  assert.match(actionSource, /下载/)
})
