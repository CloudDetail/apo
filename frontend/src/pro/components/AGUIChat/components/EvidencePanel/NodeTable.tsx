import React, { useMemo } from 'react'
import { Table, theme, Tooltip } from 'antd'
import { TableOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'
import type { ColumnsType } from 'antd/es/table'
import type { FindingsTable } from '../../../../api/conversation'
import { buildNodeTableDataSource, type TableRecord } from './nodeTableData'

const { useToken } = theme

interface NodeTableProps {
  table: FindingsTable
}

const NodeTable: React.FC<NodeTableProps> = ({ table }) => {
  const { token } = useToken()
  const { t } = useTranslation('pro/agui')
  const dataSource = useMemo<TableRecord[]>(() => buildNodeTableDataSource(table.data), [table.data])

  const columns = useMemo<ColumnsType<TableRecord>>(() => {
    return (table.headers ?? []).map((header) => {
      const titleContent = typeof header.name === 'string' ? header.name : header.name
      return {
        title: titleContent,
        dataIndex: header.key,
        key: header.key,
        width: header.width,
        ellipsis: true,
        render: (value: React.ReactNode) =>
          value === null || value === undefined || value === '' ? '-' : value,
      }
    })
  }, [table.headers])

  const tableTitle = useMemo(() => {
    if (!table.title) return t('evidencePanel.tableTitle', { defaultValue: '表格' })
    const translationKey = `evidencePanel.${table.title}`
    const translated = t(translationKey, { defaultValue: table.title })
    return translated
  }, [table.title, t])

  return (
    <div className="space-y-3">
      <div className="space-y-2">
        <div className="flex items-center gap-2">
          <TableOutlined
            style={{
              fontSize: 16,
              color: token.colorTextSecondary,
            }}
          />
          <span
            className="text-base font-medium"
            style={{
              color: token.colorTextBase,
            }}
          >
            {tableTitle}
          </span>
          {table.headers && table.headers.length > 0 && (
            <Tooltip title={t('evidencePanel.tableTooltip', { defaultValue: '数据支持依赖配置' })}>
              <span className="text-xs" style={{ color: token.colorTextTertiary }}>
                {t('evidencePanel.tableColumns', { defaultValue: `${table.headers.length} 列` })}
              </span>
            </Tooltip>
          )}
        </div>
        <div
          style={{
            height: 1,
            backgroundColor: token.colorBorder,
          }}
        />
      </div>

      <Table<TableRecord>
        columns={columns}
        dataSource={dataSource}
        rowKey="_rowKey"
        pagination={false}
        size="small"
        bordered
        scroll={{ x: 'max-content' }}
      />
    </div>
  )
}

export default NodeTable
