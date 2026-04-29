import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Alert,
  Button,
  Empty,
  Input,
  Modal,
  Select,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
  Upload,
  message,
} from 'antd'
import type { UploadFile } from 'antd/es/upload/interface'
import { Download, RefreshCw, Trash2, UploadCloud } from 'lucide-react'
import {
  DEFAULT_KNOWLEDGE_BASE_ID,
  KNOWLEDGE_BASE_FILE_ACCEPT,
  extractKnowledgeBaseErrorMessage,
  filterWritableKnowledgeBases,
  getIndexStatusDisplay,
  isIndexStatusTerminal,
  isKnowledgeBaseWritable,
  resolveKnowledgeBaseDownloadFilename,
  validateKnowledgeBaseFiles,
  type KnowledgeBaseDocument,
  type KnowledgeBaseSummary,
} from 'src/pro/api/knowledgeBase'
import {
  deleteKnowledgeBaseDocumentApi,
  downloadKnowledgeBaseDocumentApi,
  getKnowledgeBaseDocumentIndexStatusApi,
  getKnowledgeBaseDocumentsApi,
  getKnowledgeBasesApi,
  previewKnowledgeBaseDocumentApi,
  uploadKnowledgeBaseDocumentApi,
} from 'src/pro/api/knowledgeBaseClient'
import styles from './index.module.scss'

const { Title } = Typography

const INDEX_POLL_INTERVAL_MS = 3000
const INDEX_POLL_MAX_ATTEMPTS = 40

const formatFileSize = (value?: number) => {
  if (!value) {
    return '-'
  }

  if (value < 1024) {
    return `${value} B`
  }

  if (value < 1024 * 1024) {
    return `${(value / 1024).toFixed(1)} KB`
  }

  return `${(value / 1024 / 1024).toFixed(1)} MB`
}

const formatDateTime = (value?: string) => {
  if (!value) {
    return '-'
  }

  const date = new Date(value)
  if (Number.isNaN(date.getTime())) {
    return value
  }

  return date.toLocaleString()
}

const getKnowledgeBaseLabel = (item: KnowledgeBaseSummary) => item.name || '-'

const getDocumentRowKey = (record: KnowledgeBaseDocument) => record.id

const KnowledgeBasesPage = () => {
  const [knowledgeBases, setKnowledgeBases] = useState<KnowledgeBaseSummary[]>([])
  const [selectedKbId, setSelectedKbId] = useState('')
  const [documents, setDocuments] = useState<KnowledgeBaseDocument[]>([])
  const [knowledgeBaseLoading, setKnowledgeBaseLoading] = useState(false)
  const [documentLoading, setDocumentLoading] = useState(false)
  const [knowledgeBaseError, setKnowledgeBaseError] = useState<string>()
  const [documentError, setDocumentError] = useState<string>()
  const [keyword, setKeyword] = useState('')
  const [status, setStatus] = useState<string>()
  const [pagination, setPagination] = useState({ current: 1, pageSize: 20, total: 0 })
  const [uploadOpen, setUploadOpen] = useState(false)
  const [uploadFiles, setUploadFiles] = useState<UploadFile[]>([])
  const [uploading, setUploading] = useState(false)
  const [previewing, setPreviewing] = useState(false)
  const [pollingDocIds, setPollingDocIds] = useState<Set<string>>(() => new Set())
  const [downloadingDocIds, setDownloadingDocIds] = useState<Set<string>>(() => new Set())

  const selectedKnowledgeBase = useMemo(
    () => knowledgeBases.find((item) => item.id === selectedKbId),
    [knowledgeBases, selectedKbId],
  )
  const knowledgeBaseNameById = useMemo(
    () => new Map(knowledgeBases.map((item) => [item.id, getKnowledgeBaseLabel(item)])),
    [knowledgeBases],
  )
  const writable = isKnowledgeBaseWritable(selectedKnowledgeBase)

  const updateDocumentIndexStatus = useCallback(
    (docId: string, indexStatus?: string, errorMessage?: string) => {
      setDocuments((current) =>
        current.map((document) =>
          document.id === docId
            ? {
                ...document,
                indexStatus: indexStatus ?? document.indexStatus,
                errorMessage: errorMessage ?? document.errorMessage,
              }
            : document,
        ),
      )
    },
    [],
  )

  const pollIndexStatus = useCallback(
    async (docId: string) => {
      setPollingDocIds((current) => new Set(current).add(docId))

      try {
        for (let attempt = 0; attempt < INDEX_POLL_MAX_ATTEMPTS; attempt += 1) {
          const result = await getKnowledgeBaseDocumentIndexStatusApi(docId)
          updateDocumentIndexStatus(docId, result.indexStatus, result.errorMessage)

          if (isIndexStatusTerminal(result.indexStatus)) {
            return
          }

          await new Promise((resolve) => setTimeout(resolve, INDEX_POLL_INTERVAL_MS))
        }

        message.warning('导入状态查询超时，请稍后手动刷新')
      } catch (error) {
        message.error(extractKnowledgeBaseErrorMessage(error, '查询导入状态失败'))
      } finally {
        setPollingDocIds((current) => {
          const next = new Set(current)
          next.delete(docId)
          return next
        })
      }
    },
    [updateDocumentIndexStatus],
  )

  const loadKnowledgeBases = useCallback(async () => {
    setKnowledgeBaseLoading(true)
    setKnowledgeBaseError(undefined)

    try {
      const result = await getKnowledgeBasesApi({ current_page: 1, page_size: 50, status: 'active' })
      const writableItems = filterWritableKnowledgeBases(result.items)
      setKnowledgeBases(writableItems)

      if (!writableItems.some((item) => item.id === selectedKbId)) {
        setSelectedKbId(writableItems.find((item) => item.id === DEFAULT_KNOWLEDGE_BASE_ID)?.id ?? writableItems[0]?.id ?? '')
      }
    } catch (error) {
      setKnowledgeBaseError(extractKnowledgeBaseErrorMessage(error, '加载知识库失败'))
    } finally {
      setKnowledgeBaseLoading(false)
    }
  }, [selectedKbId])

  const loadDocuments = useCallback(
    async (nextPage = pagination.current, nextPageSize = pagination.pageSize) => {
      if (!selectedKbId) {
        setDocuments([])
        setPagination({ current: nextPage, pageSize: nextPageSize, total: 0 })
        return
      }

      setDocumentLoading(true)
      setDocumentError(undefined)

      try {
        const result = await getKnowledgeBaseDocumentsApi(selectedKbId, {
          page: nextPage,
          pageSize: nextPageSize,
          keyword,
          status,
        })
        setDocuments(result.items)
        setPagination({
          current: Math.floor(result.offset / result.limit) + 1,
          pageSize: result.limit,
          total: result.total,
        })
      } catch (error) {
        setDocumentError(extractKnowledgeBaseErrorMessage(error, '加载文档失败'))
      } finally {
        setDocumentLoading(false)
      }
    },
    [keyword, pagination.current, pagination.pageSize, selectedKbId, status],
  )

  useEffect(() => {
    loadKnowledgeBases()
  }, [loadKnowledgeBases])

  useEffect(() => {
    loadDocuments(1, pagination.pageSize)
  }, [selectedKbId, keyword, status])

  const handleRefresh = () => {
    loadKnowledgeBases()
    loadDocuments()
  }

  const handleSearch = (value: string) => {
    setKeyword(value.trim())
    setPagination((current) => ({ ...current, current: 1 }))
  }

  const handleUpload = async () => {
    const files = uploadFiles
      .map((file) => file.originFileObj)
      .filter((file): file is File => Boolean(file))
    const errors = validateKnowledgeBaseFiles(files)

    if (errors.length > 0) {
      message.error(errors[0])
      return
    }

    if (files.length === 0) {
      message.warning('请选择要上传的文档')
      return
    }

    setUploading(true)
    try {
      for (const file of files) {
        const result = await uploadKnowledgeBaseDocumentApi(selectedKbId, file)
        if (result.docId && result.indexStatus === 'indexing') {
          pollIndexStatus(result.docId)
        }
      }
      message.success('文档上传成功')
      setUploadFiles([])
      setUploadOpen(false)
      loadDocuments(1, pagination.pageSize)
    } catch (error) {
      message.error(extractKnowledgeBaseErrorMessage(error, '上传失败'))
    } finally {
      setUploading(false)
    }
  }

  const handlePreview = async () => {
    const file = uploadFiles[0]?.originFileObj
    if (!file) {
      message.warning('请选择要预览的文档')
      return
    }

    const errors = validateKnowledgeBaseFiles([file])
    if (errors.length > 0) {
      message.error(errors[0])
      return
    }

    setPreviewing(true)
    try {
      const result = await previewKnowledgeBaseDocumentApi(selectedKbId, file)
      const nodes = Array.isArray(result?.nodes) ? result.nodes : []
      Modal.info({
        title: '切片预览',
        width: 760,
        content:
          nodes.length > 0 ? (
            <div className={styles.previewList}>
              {nodes.slice(0, 20).map((node: any, index: number) => (
                <div key={node.id ?? index} className={styles.previewItem}>
                  {node.content ?? ''}
                </div>
              ))}
            </div>
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无切片内容" />
          ),
      })
    } catch (error) {
      message.error(extractKnowledgeBaseErrorMessage(error, '预览失败'))
    } finally {
      setPreviewing(false)
    }
  }

  const handleDelete = (record: KnowledgeBaseDocument) => {
    Modal.confirm({
      title: '删除文档',
      content: `确认删除 ${record.filename}？`,
      okText: '删除',
      okButtonProps: { danger: true },
      cancelText: '取消',
      async onOk() {
        try {
          await deleteKnowledgeBaseDocumentApi(selectedKbId, record.id)
          message.success('文档删除成功')
          loadDocuments()
        } catch (error) {
          message.error(extractKnowledgeBaseErrorMessage(error, '删除失败'))
        }
      },
    })
  }

  const handleDownload = async (record: KnowledgeBaseDocument) => {
    setDownloadingDocIds((current) => new Set(current).add(record.id))

    try {
      const response = await downloadKnowledgeBaseDocumentApi(selectedKbId, record.id)
      const filename = resolveKnowledgeBaseDownloadFilename(response.headers?.['content-disposition'], record.filename)
      const blob = response.data instanceof Blob
        ? response.data
        : new Blob([response.data], { type: response.headers?.['content-type'] || 'application/octet-stream' })
      const url = window.URL.createObjectURL(blob)
      const link = document.createElement('a')
      link.href = url
      link.download = filename
      link.style.display = 'none'
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
      window.setTimeout(() => window.URL.revokeObjectURL(url), 0)
    } catch (error) {
      message.error(extractKnowledgeBaseErrorMessage(error, '下载失败'))
    } finally {
      setDownloadingDocIds((current) => {
        const next = new Set(current)
        next.delete(record.id)
        return next
      })
    }
  }

  const columns = [
    {
      title: '文档名称',
      dataIndex: 'filename',
      key: 'filename',
      ellipsis: true,
    },
    {
      title: '知识库',
      dataIndex: 'kbId',
      key: 'kbId',
      width: 180,
      render: (value?: string) => value ? knowledgeBaseNameById.get(value) ?? '-' : '-',
    },
    {
      title: '文件类型',
      key: 'type',
      width: 120,
      render: (_: unknown, record: KnowledgeBaseDocument) => record.extension || record.mimeType || '-',
    },
    {
      title: '文件大小',
      dataIndex: 'sizeBytes',
      key: 'sizeBytes',
      width: 120,
      render: (value?: number) => formatFileSize(value),
    },
    {
      title: '文档状态',
      dataIndex: 'status',
      key: 'status',
      width: 120,
      render: (value?: string) => value || '-',
    },
    {
      title: '导入状态',
      dataIndex: 'indexStatus',
      key: 'indexStatus',
      width: 140,
      render: (value?: string, record?: KnowledgeBaseDocument) => {
        const display = getIndexStatusDisplay(value)
        const polling = record?.id ? pollingDocIds.has(record.id) : false
        return (
          <Space size={6}>
            <Tag color={display.color}>{display.label}</Tag>
            <Tooltip title="刷新导入状态">
              <Button
                type="text"
                size="small"
                icon={<RefreshCw size={14} />}
                loading={polling}
                onClick={() => record?.id && pollIndexStatus(record.id)}
              />
            </Tooltip>
          </Space>
        )
      },
    },
    {
      title: '上传时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      width: 190,
      render: (value?: string) => formatDateTime(value),
    },
    {
      title: '失败原因',
      dataIndex: 'errorMessage',
      key: 'errorMessage',
      ellipsis: true,
      render: (value?: string) => value || '-',
    },
    {
      title: '操作',
      key: 'actions',
      width: 230,
      fixed: 'right' as const,
      render: (_: unknown, record: KnowledgeBaseDocument) => (
        <Space>
          <Button
            icon={<Download size={15} />}
            size="small"
            onClick={() => handleDownload(record)}
            loading={downloadingDocIds.has(record.id)}
          >
            下载
          </Button>
          {writable && (
            <Tooltip title="删除">
              <Button
                danger
                icon={<Trash2 size={15} />}
                size="small"
                onClick={() => handleDelete(record)}
              />
            </Tooltip>
          )}
        </Space>
      ),
    },
  ]

  return (
    <div className={styles.knowledgeBasesPage}>
      <div className={styles.header}>
        <Title level={4}>知识库管理</Title>
      </div>

      {knowledgeBaseError && (
        <Alert type="error" message={knowledgeBaseError} action={<Button onClick={loadKnowledgeBases}>重试</Button>} />
      )}

      <div className={styles.toolbar}>
        <div className={styles.filters}>
          <Select
            value={selectedKbId}
            loading={knowledgeBaseLoading}
            style={{ minWidth: 280 }}
            options={knowledgeBases.map((item) => ({
              value: item.id,
              label: getKnowledgeBaseLabel(item),
            }))}
            onChange={(value) => {
              setSelectedKbId(value)
              setPagination((current) => ({ ...current, current: 1 }))
            }}
          />
          <Input.Search
            allowClear
            placeholder="搜索文档名称"
            style={{ width: 260 }}
            onSearch={handleSearch}
          />
          <Select
            allowClear
            placeholder="文档状态"
            style={{ width: 160 }}
            value={status}
            options={[
              { value: 'active', label: 'active' },
              { value: 'deleted', label: 'deleted' },
            ]}
            onChange={(value) => {
              setStatus(value)
              setPagination((current) => ({ ...current, current: 1 }))
            }}
          />
        </div>
        <Space wrap className={styles.actions}>
          <Button icon={<RefreshCw size={16} />} onClick={handleRefresh}>
            刷新
          </Button>
          {writable && (
            <Button type="primary" icon={<UploadCloud size={16} />} onClick={() => setUploadOpen(true)}>
              上传文档
            </Button>
          )}
        </Space>
      </div>

      {documentError && (
        <Alert type="error" message={documentError} action={<Button onClick={() => loadDocuments()}>重试</Button>} />
      )}

      <div className={styles.tableWrap}>
        <Table
          rowKey={getDocumentRowKey}
          columns={columns}
          dataSource={documents}
          loading={documentLoading}
          scroll={{ x: 1280 }}
          locale={{
            emptyText: (
              <Empty
                image={Empty.PRESENTED_IMAGE_SIMPLE}
                description={writable ? '暂无文档' : '当前知识库暂无文档'}
              />
            ),
          }}
          pagination={{
            current: pagination.current,
            pageSize: pagination.pageSize,
            total: pagination.total,
            showSizeChanger: true,
          }}
          onChange={(nextPagination) => {
            loadDocuments(nextPagination.current ?? 1, nextPagination.pageSize ?? pagination.pageSize)
          }}
        />
      </div>

      <Modal
        title="上传文档"
        open={uploadOpen}
        onCancel={() => {
          if (!uploading) {
            setUploadOpen(false)
          }
        }}
        footer={null}
        destroyOnClose
      >
        <Upload.Dragger
          multiple
          accept={KNOWLEDGE_BASE_FILE_ACCEPT}
          maxCount={5}
          fileList={uploadFiles}
          beforeUpload={() => false}
          onChange={({ fileList }) => {
            setUploadFiles(fileList)
          }}
        >
          <p className="ant-upload-drag-icon">
            <UploadCloud size={32} />
          </p>
          <p className="ant-upload-text">选择或拖入文档</p>
        </Upload.Dragger>
        <div className={styles.uploadActions}>
          <Button onClick={handlePreview} loading={previewing} disabled={uploadFiles.length === 0}>
            预览切片
          </Button>
          <Button type="primary" onClick={handleUpload} loading={uploading} disabled={uploadFiles.length === 0}>
            上传
          </Button>
        </div>
      </Modal>
    </div>
  )
}

export default KnowledgeBasesPage
