import React, { useState, useEffect } from 'react'
import { Modal, Table, Button, Typography } from 'antd'
import { useTranslation } from 'react-i18next'
import { getCapabilitiesApi } from 'src/pro/api/dataplane'
import { dataplaneIconMap } from 'src/pro/constants'

const { Text, Paragraph } = Typography

interface DataSource {
  type: string
  name: string
}

interface CapabilityItem {
  type: string
  enabled: boolean
  sources: DataSource[]
  description: string
}

interface DataCoverageOverviewProps {
  visible: boolean
  onClose: () => void
  onQuickSetup?: (type: string) => void
}

const DataCoverageOverview: React.FC<DataCoverageOverviewProps> = ({
  visible,
  onClose,
  onQuickSetup,
}) => {
  const { t } = useTranslation('pro/agent')
  const [loading, setLoading] = useState(false)
  const [capabilities, setCapabilities] = useState<CapabilityItem[]>([])

  useEffect(() => {
    if (visible) {
      fetchCapabilities()
    }
  }, [visible])

  const fetchCapabilities = async () => {
    setLoading(true)
    try {
      const res: any = await getCapabilitiesApi()
      const sortedData = (res?.totalSupported || []).sort((a: CapabilityItem, b: CapabilityItem) =>
        a.type.localeCompare(b.type),
      )
      setCapabilities(sortedData)
    } catch (error) {
      console.error('Failed to fetch capabilities:', error)
    } finally {
      setLoading(false)
    }
  }

  const handleQuickSetup = (type: string) => {
    // 关闭弹窗
    onClose()
    // 调用父组件的回调函数，传递要筛选的 type
    if (onQuickSetup) {
      onQuickSetup(type)
    }
  }

  const handleDoItLater = () => {
    onClose()
  }

  const columns = [
    {
      title: t('dataCoverage.dataType'),
      dataIndex: 'type',
      key: 'type',
      width: 150,
      render: (type: string, record: CapabilityItem) => (
        <Text
          className={`text-[${record.enabled ? 'var(--ant-color-success-text)' : 'var(--ant-color-error-text)'}`}
        >
          {type}
        </Text>
      ),
    },
    {
      title: t('dataCoverage.dataTypeDescription'),
      dataIndex: 'type',
      key: 'typeDescription',
      width: 320,
      render: (type: string) => (
        <Text className="text-[var(--ant-color-text-secondary)]">
          {t(`capabilityDescriptions.${type}`)}
        </Text>
      ),
    },
    {
      title: t('dataCoverage.providersSupplyingData'),
      dataIndex: 'sources',
      key: 'sources',
      width: 220,
      render: (sources: DataSource[], record: CapabilityItem) => {
        if (record.enabled && sources && sources.length > 0) {
          // 按照 type 去重

          return (
            <>
              {sources.map((source, index) => (
                <div key={index} className="inline-flex items-center gap-1 mr-3">
                  <img
                    src={dataplaneIconMap[source.type]}
                    className="w-4 h-4 object-contain"
                    alt={source.name}
                  />
                  <Text className="text-[var(--ant-color-success-text)]">{source.name}</Text>
                </div>
              ))}
            </>
          )
        }
        return <Text type="secondary">{t('dataCoverage.none')}</Text>
      },
    },
    {
      title: t('dataCoverage.impactIfMissing'),
      dataIndex: 'description',
      key: 'description',
      render: (description: string, record: CapabilityItem) => {
        if (record.enabled) {
          return (
            <Text className="text-[var(--ant-color-success-text)]">
              {t('dataCoverage.covered')}
            </Text>
          )
        }
        return (
          <Text className="text-[var(--ant-color-text-tertiary)]">
            {description || t('dataCoverage.missingDataImpact')}
          </Text>
        )
      },
    },
    {
      title: t('dataCoverage.action'),
      key: 'action',
      width: 120,
      render: (record: CapabilityItem) => {
        if (record.enabled) {
          return null
        }
        return (
          <Button size="small" type="primary" onClick={() => handleQuickSetup(record.type)}>
            {t('dataCoverage.quickSetup')}
          </Button>
        )
      },
    },
  ]

  return (
    <Modal
      title={t('dataCoverage.title')}
      open={visible}
      onCancel={onClose}
      footer={
        <div style={{ textAlign: 'right' }}>
          <Button
            className="bg-[var(--ant-color-fill-quaternary)] border-[var(--ant-color-fill-quaternary)] text-[var(--ant-color-text-secondary)]"
            onClick={handleDoItLater}
          >
            {t('dataCoverage.doItLater')}
          </Button>
        </div>
      }
      width={'70%'}
      centered
      destroyOnClose
    >
      <div style={{ padding: '16px 0' }}>
        <Paragraph className="text-[var(--ant-color-text-tertiary)] mb-6">
          {t('dataCoverage.description')}
        </Paragraph>

        <Table
          columns={columns}
          dataSource={capabilities}
          loading={loading}
          rowKey="type"
          pagination={false}
          size="middle"
          rowClassName={(record) => (record.enabled ? 'covered-row' : '')}
          components={{
            body: {
              row: (props: any) => {
                const { className: _className, ...restProps } = props
                const isCovered = capabilities.find(
                  (c) => c.type === props['data-row-key'],
                )?.enabled
                return (
                  <tr
                    {...restProps}
                    style={{
                      backgroundColor: isCovered ? 'var(--ant-color-success-bg)' : '',
                    }}
                  />
                )
              },
            },
          }}
        />
      </div>
    </Modal>
  )
}

export default DataCoverageOverview
