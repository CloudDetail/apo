import test from 'node:test'
import assert from 'node:assert/strict'
import { readFileSync } from 'node:fs'
import { resolve } from 'node:path'

import {
  DEFAULT_KNOWLEDGE_BASE_ID,
  buildKnowledgeBaseApiPath,
  buildDocumentFormData,
  buildKnowledgeBaseDocumentDownloadUrl,
  buildKnowledgeBaseDocumentPreviewUrl,
  buildDocumentListRequestParams,
  buildDocumentPreviewRequestParams,
  buildUploadedDocumentPreviewRequestParams,
  extractKnowledgeBaseErrorMessage,
  filterWritableKnowledgeBases,
  getIndexStatusDisplay,
  isIndexStatusTerminal,
  isKnowledgeBaseWritable,
  normalizeDocumentListResponse,
  normalizeKnowledgeBaseListResponse,
  normalizeUploadedDocumentPreviewResponse,
  resolveKnowledgeBaseDownloadFilename,
  resolveKnowledgeBasePreviewNextOffset,
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

test('builds uploaded document preview url through rag proxy', () => {
  assert.equal(
    buildKnowledgeBaseDocumentPreviewUrl(DEFAULT_KNOWLEDGE_BASE_ID, 'doc_001'),
    '/rag/api/kb/knowledge_bases/syncause_default/documents/doc_001/preview',
  )

  assert.equal(
    buildKnowledgeBaseDocumentPreviewUrl('ops kb', 'doc/001'),
    '/rag/api/kb/knowledge_bases/ops%20kb/documents/doc%2F001/preview',
  )
})

test('builds encoded rag api paths from path segments', () => {
  assert.equal(
    buildKnowledgeBaseApiPath('knowledge_bases', 'ops kb', 'documents', 'doc/001'),
    '/rag/api/kb/knowledge_bases/ops%20kb/documents/doc%2F001',
  )
})

test('builds uploaded document preview query params for scroll pagination', () => {
  assert.deepEqual(buildDocumentPreviewRequestParams(), {
    limit: 50,
    offset: 0,
  })

  assert.deepEqual(buildUploadedDocumentPreviewRequestParams(), {
    limit: 50,
    offset: 0,
  })

  assert.deepEqual(buildUploadedDocumentPreviewRequestParams({ limit: 200, offset: 150 }), {
    limit: 200,
    offset: 150,
  })

  assert.deepEqual(buildUploadedDocumentPreviewRequestParams({ limit: 500, offset: -10 }), {
    limit: 200,
    offset: 0,
  })
})

test('builds document multipart form data with rag default chunk options', () => {
  const file = new File(['runbook'], 'runbook.md', { type: 'text/markdown' })
  const formData = buildDocumentFormData(file)

  assert.equal(formData.get('file'), file)
  assert.equal(formData.get('chunk_method'), 'separator')
  assert.deepEqual(JSON.parse(String(formData.get('chunk_params'))), {
    separator: ['\n\n', '\n', '.', '。', '?', '？', '!', '！', ' '],
    chunk_size: 1024,
    chunk_overlap: 100,
  })
})

test('normalizes uploaded document preview responses', () => {
  const normalized = normalizeUploadedDocumentPreviewResponse({
    success: true,
    kb_id: DEFAULT_KNOWLEDGE_BASE_ID,
    doc_id: 'doc_001',
    filename: 'CPU高排查Runbook.md',
    index_status: 'indexed',
    total: 2,
    limit: 50,
    offset: 0,
    nodes: [
      {
        id: 'chunk_001',
        content: '切片文本内容',
        metadata: {
          page: 1,
          section: '排查步骤',
        },
      },
      {
        id: 'chunk_empty',
        content: '   ',
      },
    ],
  })

  assert.deepEqual(normalized, {
    success: true,
    kbId: DEFAULT_KNOWLEDGE_BASE_ID,
    docId: 'doc_001',
    filename: 'CPU高排查Runbook.md',
    indexStatus: 'indexed',
    total: 2,
    limit: 50,
    offset: 0,
    nodes: [
      {
        id: 'chunk_001',
        content: '切片文本内容',
        metadata: {
          page: 1,
          section: '排查步骤',
        },
      },
    ],
  })
})

test('preserves non-empty preview chunk whitespace while dropping whitespace-only nodes', () => {
  const normalized = normalizeUploadedDocumentPreviewResponse({
    total: 2,
    limit: 50,
    offset: 0,
    nodes: [
      {
        id: 'chunk_code',
        content: '    code block\n  line2  ',
      },
      {
        id: 'chunk_empty',
        content: '   ',
      },
    ],
  })

  assert.equal(normalized.nodes.length, 1)
  assert.equal(normalized.nodes[0]?.id, 'chunk_code')
  assert.equal(normalized.nodes[0]?.content, '    code block\n  line2  ')
})

test('advances preview pagination by server limit even when empty chunks are filtered', () => {
  const normalized = normalizeUploadedDocumentPreviewResponse({
    total: 120,
    limit: 50,
    offset: 0,
    nodes: [
      {
        id: 'chunk_001',
        content: '切片文本内容',
      },
      {
        id: 'chunk_empty',
        content: '   ',
      },
    ],
  })

  assert.equal(normalized.nodes.length, 1)
  assert.equal(resolveKnowledgeBasePreviewNextOffset(normalized), 50)
  assert.equal(
    resolveKnowledgeBasePreviewNextOffset({
      ...normalized,
      total: 120,
      limit: 50,
      offset: 100,
      nodes: [{ content: '最后一页' }],
    }),
    120,
  )
  assert.equal(
    resolveKnowledgeBasePreviewNextOffset({
      ...normalized,
      total: 20,
      limit: 0,
      offset: 0,
      nodes: [{ content: '兜底 1' }, { content: '兜底 2' }],
    }),
    2,
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

test('knowledge base upload file list items are left aligned inside centered dragger', () => {
  const stylesSource = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.module.scss'), 'utf8')

  assert.match(stylesSource, /\.uploadFileItem\s*{[^}]*text-align:\s*left;/)
  assert.match(stylesSource, /\.uploadFileNode\s*{[^}]*text-align:\s*left;/)
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

test('knowledge base page previews uploaded document chunks from row action', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')
  const clientSource = readFileSync(resolve(process.cwd(), 'src/pro/api/knowledgeBaseClient.ts'), 'utf8')
  const actionStart = source.indexOf("title: '操作'")
  const actionEnd = source.indexOf(']', actionStart)
  const actionSource = source.slice(actionStart, actionEnd)

  assert.match(clientSource, /getUploadedKnowledgeBaseDocumentPreviewApi/)
  assert.match(clientSource, /buildKnowledgeBaseDocumentPreviewUrl\(kbId, docId\)/)
  assert.match(clientSource, /buildUploadedDocumentPreviewRequestParams\(params\)/)
  assert.match(clientSource, /normalizeUploadedDocumentPreviewResponse/)
  assert.match(source, /PREVIEW_PAGE_SIZE = 50/)
  assert.match(source, /getUploadedKnowledgeBaseDocumentPreviewApi\(selectedKbId, record\.id, \{\s*limit: PREVIEW_PAGE_SIZE,\s*offset: 0,\s*\}\)/)
  assert.match(source, /offset: previewNextOffset/)
  assert.match(source, /setPreviewNodes\(\(current\) => \[\.\.\.current, \.\.\.result\.nodes\]\)/)
  assert.match(source, /handlePreviewScroll/)
  assert.match(source, /setPreviewingDocIds/)
  assert.match(actionSource, /onClick=\{\(\) => handlePreviewUploadedDocument\(record\)\}/)
  assert.match(actionSource, /预览/)
})

test('knowledge base upload preview posts the selected file with paged multipart requests', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')
  const clientSource = readFileSync(resolve(process.cwd(), 'src/pro/api/knowledgeBaseClient.ts'), 'utf8')

  assert.match(clientSource, /previewKnowledgeBaseDocumentApi = async \(\s*kbId: string,\s*file: File,\s*params\?: UploadedDocumentPreviewRequest,/)
  assert.match(clientSource, /post\(\s*buildKnowledgeBaseApiPath\('knowledge_bases', kbId, 'documents', 'preview'\)/)
  assert.match(clientSource, /params: buildDocumentPreviewRequestParams\(params\)/)
  assert.match(clientSource, /normalizeUploadedDocumentPreviewResponse/)
  assert.match(source, /type: 'uploadFile'; kbId: string; file: File/)
  assert.match(source, /resolveKnowledgeBasePreviewNextOffset\(result\)/)
  assert.match(source, /previewKnowledgeBaseDocumentApi\(selectedKbId, originFile, \{\s*limit: PREVIEW_PAGE_SIZE,\s*offset: 0,\s*\}\)/)
  assert.match(source, /previewTarget\.type === 'uploadFile'/)
  assert.match(source, /previewKnowledgeBaseDocumentApi\(previewTarget\.kbId, previewTarget\.file, \{\s*limit: PREVIEW_PAGE_SIZE,\s*offset: previewNextOffset,\s*\}\)/)
})

test('knowledge base preview ignores stale async page results after switching previews', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')

  assert.match(source, /const previewRequestSeqRef = useRef\(0\)/)
  assert.match(source, /const startPreviewRequest = \(\) => \{\s*previewRequestSeqRef\.current \+= 1/)
  assert.match(source, /const isCurrentPreviewRequest = \(requestSeq: number\) => requestSeq === previewRequestSeqRef\.current/)
  assert.match(source, /if \(!isCurrentPreviewRequest\(requestSeq\)\) \{\s*return\s*\}/)
  assert.match(source, /onCancel=\{handlePreviewModalCancel\}/)
})

test('knowledge base preview auto-loads more when returned chunks do not fill the modal', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')

  assert.match(source, /const previewListRef = useRef<HTMLDivElement>\(null\)/)
  assert.match(source, /previewListRef\.current/)
  assert.match(source, /previewList\.scrollHeight <= previewList\.clientHeight/)
  assert.match(source, /loadNextPreviewPage\(\)/)
})

test('knowledge base preview modal renders returned chunks without fixed local truncation', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')

  assert.equal(source.includes('previewNodes.slice(0, 20)'), false)
  assert.match(source, /previewNodes\.map\(\(node, index\) =>/)
})

test('knowledge base preview uses controlled modal so theme context is preserved', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')

  assert.equal(source.includes('Modal.info('), false)
  assert.match(source, /const \[previewModalOpen, setPreviewModalOpen\] = useState\(false\)/)
  assert.match(source, /const \[previewNodes, setPreviewNodes\] = useState<KnowledgeBasePreviewNode\[\]>\(\[\]\)/)
  assert.match(source, /setPreviewModalOpen\(true\)/)
  assert.match(source, /title="切片预览"/)
  assert.match(source, /open=\{previewModalOpen\}/)
})

test('knowledge base upload preview is attached to each selected file', () => {
  const source = readFileSync(resolve(process.cwd(), 'src/pro/views/knowledgeBases/index.tsx'), 'utf8')
  const uploadDraggerStart = source.indexOf('<Upload.Dragger')
  const uploadActionsStart = source.indexOf('<div className={styles.uploadActions}>', uploadDraggerStart)
  const uploadDraggerSource = source.slice(uploadDraggerStart, uploadActionsStart)
  const uploadActionsSource = source.slice(uploadActionsStart, source.indexOf('</Modal>', uploadActionsStart))

  assert.equal(source.includes('const [previewing, setPreviewing] = useState(false)'), false)
  assert.equal(source.includes('uploadFiles[0]?.originFileObj'), false)
  assert.match(source, /handleUploadFilePreview = async \(file: UploadFile\)/)
  assert.match(source, /previewingUploadFileUids/)
  assert.match(uploadDraggerSource, /itemRender=\{\(originNode, file\) =>/)
  assert.match(uploadDraggerSource, /onClick=\{\(\) => handleUploadFilePreview\(file\)\}/)
  assert.match(uploadDraggerSource, /预览/)
  assert.equal(uploadActionsSource.includes('预览切片'), false)
})
