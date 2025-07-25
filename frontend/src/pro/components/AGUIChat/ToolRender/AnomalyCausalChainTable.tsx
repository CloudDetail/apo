import { Button, Table } from 'antd'
import { useState } from 'react'
import { TbArrowNarrowRight, TbArrowNarrowRightDashed, TbChevronDown } from 'react-icons/tb'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'

const AnomalyCausalChainTable = ({ args }: { args: any }) => {
  const { columns, data, title = null } = args
  const defaultColumns = [
    {
      title: '因果关系链',
      dataIndex: 'chain',
      width: 300,
      render: (text, record) => {
        return (
          <div className="flex items-center  gap-2">
            <span className="w-[150px]">{record.from}</span>
            <div className="mx-4">
              <TbArrowNarrowRight size={15} />
            </div>
            <span>{record.to}</span>
          </div>
        )
      },
    },
    {
      title: () => {
        return (
          <div className="flex items-center">
            因果关系链
            <div className="flex items-center text-xs">
              {'('}
              <TbArrowNarrowRight className="text-[var(--ant-color-primary)]" size={15} />
              存在真实Trace连接
              <TbArrowNarrowRightDashed
                className="text-[var(--ant-color-text-secondary)] ml-2"
                size={15}
              />
              非真实连接{')'}
            </div>
          </div>
        )
      },
      dataIndex: 'traceChain',
      width: 400,
      render: (text, record) => {
        return (
          <div className="flex items-center gap-2">
            <span className="w-[150px]">{record.from}</span>
            <div className="mx-4">
              {record.isTrue ? (
                <TbArrowNarrowRight className="text-[var(--ant-color-primary)]" size={30} />
              ) : (
                <TbArrowNarrowRightDashed
                  className="text-[var(--ant-color-text-secondary)]"
                  size={30}
                />
              )}
            </div>
            <span>{record.to}</span>
          </div>
        )
      },
    },
    {
      title: '存在自身异常的节点',
      dataIndex: 'service',
      width: 150,
    },
    {
      title: '置信度',
      dataIndex: 'confidence',
      width: 100,
    },
  ]
  const [expandedRowKeys, setExpandedRowKeys] = useState<React.Key[]>([])
  return (
    <Table
      size="small"
      rowKey={(record) => record.from + record.to + record?.service}
      columns={[
        ...defaultColumns.filter((item) => columns?.includes(item.dataIndex)),
        ...(columns?.includes('analysis') ? [Table.EXPAND_COLUMN] : []),
      ]}
      dataSource={data}
      rowClassName={(record, index) => {
        return columns?.includes('analysis') && record.isDelete ? 'line-through' : ''
      }}
      expandable={{
        expandedRowKeys: expandedRowKeys,
        onExpandedRowsChange: (keys) => {
          setExpandedRowKeys(keys)
        },
        expandedRowRender: (record) => <MarkdownRender content={record.analysis} />,
        rowExpandable: () => true,
        expandRowByClick: true,
        showExpandColumn: columns?.includes('analysis'),
        expandIcon: ({ expanded, onExpand, record }) => (
          <Button
            type="link"
            onClick={(e) => {
              e.stopPropagation()
              const key = record.from + record.to + record?.service
              const isExpanded = expandedRowKeys.includes(key)
              const newKeys = isExpanded
                ? expandedRowKeys.filter((k) => k !== key)
                : [...expandedRowKeys, key]
              setExpandedRowKeys(newKeys)
            }}
            icon={<TbChevronDown size={15} />}
          >
            {expanded ? '收起' : 'eBPF证据'}
          </Button>
        ),
      }}
    />
  )
}

export default AnomalyCausalChainTable
