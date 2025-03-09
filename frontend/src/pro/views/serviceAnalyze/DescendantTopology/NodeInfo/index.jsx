import { Collapse } from 'antd'
import { useEffect, useState } from 'react'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import AnormalList from './AnormalList'
import ErrorInstanceInfo from 'src/pro/views/serviceAnalyze/DescendantTopology/NodeInfo/ErrorInstanceInfo'
import PolarisMetricsInfo from 'src/pro/views/serviceAnalyze/DescendantTopology/NodeInfo/PolarisMetricsInfoModal/PolarisMetricsInfo'
import { useTranslation } from 'react-i18next'

const i18nKey = 'node'
const NodeInfo = () => {
  const { t } = useTranslation('pro/serviceAnalyze')
  const [viewportHeight, setViewportHeight] = useState(window.innerHeight)
  const [activeItem, setActiveItem] = useState([])
  const { selectedNodeInfo } = useAlertAnalyzeContext()
  const [errorInstanceOpen, setErrorInstanceOpen] = useState(false)

  useEffect(() => {
    if (activeItem.includes('2')) {
      setErrorInstanceOpen(true)
    } else {
      setErrorInstanceOpen(false)
    }
  }, [activeItem])

  const items = [
    {
      key: '1',
      label: (
        <div className="flex items-center justify-between">
          <div className="flex items-center">{t(`${i18nKey}.delayReport`)}</div>
        </div>
      ),
      children: (
        <PolarisMetricsInfo
          service={selectedNodeInfo?.service}
          endpoint={selectedNodeInfo?.endpoint}
        />
      ),
      forceRender: true,
      styles: {
        body: {
          paddingRight: 0,
        },
      },
    },
    {
      key: '2',
      label: t(`${i18nKey}.errorReport`),
      children: (
        <ErrorInstanceInfo
          service={selectedNodeInfo?.service}
          endpoint={selectedNodeInfo?.endpoint}
          showLogs={false}
          errorInstanceOpen={errorInstanceOpen}
        />
      ),
      forceRender: true,
      styles: {
        body: {
          paddingRight: 0,
        },
      },
    },
    {
      key: '3',
      label: (
        <div>
          {t(`${i18nKey}.alertList`)} (
          {selectedNodeInfo?.endpoint ? (
            <span>{t(`${i18nKey}.selected`)}</span>
          ) : (
            <span>{t(`${i18nKey}.all`)}</span>
          )}
          {t(`${i18nKey}.nodeTimeData`)})
        </div>
      ),
      children: <AnormalList />,
      forceRender: true,
      styles: {
        body: {
          height: viewportHeight - 150,
          paddingRight: 0,
        },
      },
    },
  ]
  return (
    <>
      <Collapse
        className="h-full relative"
        items={items}
        defaultActiveKey={['3']}
        onChange={(value) => setActiveItem(value)}
      />
    </>
  )
}
export default NodeInfo
