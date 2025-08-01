import { Button, Table } from 'antd'
import { useState } from 'react'
import { TbArrowNarrowRight, TbArrowNarrowRightDashed, TbChevronDown } from 'react-icons/tb'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import { useTranslation } from 'react-i18next'

const AnomalyCausalChainTable = ({ args }: { args: any }) => {
  const { t } = useTranslation('pro/chatAI')
  const { columns, data } = args
  const defaultColumns = [
    {
      title: t('anomalyCausalChain.title'),
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
            {t('anomalyCausalChain.traceChain')}
            <div className="flex ml-4 items-center text-xs text-[var(--ant-color-text-secondary)]">
              {'('}
              <TbArrowNarrowRight className="text-[var(--ant-color-primary)]" size={15} />
              {t('anomalyCausalChain.realTraceConnection')}
              <TbArrowNarrowRightDashed
                className="text-[var(--ant-color-text-secondary)] ml-2"
                size={15}
              />
              {t('anomalyCausalChain.nonRealConnection')}
              {')'}
            </div>
          </div>
        )
      },
      dataIndex: 'traceChain',
      width: 400,
      render: (text, record) => {
        return (
          <div className="flex items-center gap-2">
            <span
              className={`w-[150px]  ${columns?.includes('analysis') && record?.deleteService?.includes(record.from) ? 'line-through text-[var(--ant-color-text-tertiary)]' : ''}`}
            >
              {record.from}
            </span>
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
            <span
              className={`w-[150px]  ${columns?.includes('analysis') && record?.deleteService?.includes(record.to) ? 'line-through text-[var(--ant-color-text-tertiary)]' : ''}`}
            >
              {record.to}
            </span>
          </div>
        )
      },
    },
    // {
    //   title: '存在自身异常的节点',
    //   dataIndex: 'service',
    //   width: 150,
    // },
    {
      title: t('anomalyCausalChain.confidence'),
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
      // rowClassName={(record, index) => {
      //   return
      // }}
      expandable={{
        expandedRowKeys: expandedRowKeys,
        onExpandedRowsChange: (keys) => {
          setExpandedRowKeys([...keys])
        },
        expandedRowRender: (record) => <MarkdownRender content={record.analysis} theme="light" />,
        rowExpandable: () => true,
        expandRowByClick: true,
        showExpandColumn: columns?.includes('analysis'),
        expandIcon: ({ expanded, record }) => (
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
            {expanded ? t('anomalyCausalChain.collapse') : t('anomalyCausalChain.ebpfEvidence')}
          </Button>
        ),
      }}
    />
  )
}

export default AnomalyCausalChainTable
