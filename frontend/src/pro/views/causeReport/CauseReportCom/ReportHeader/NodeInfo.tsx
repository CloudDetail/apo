import { Card } from 'antd'
import TimeProgress from './TimeProgress'
import { convertTime, formatUnit } from 'src/core/utils/time'
import { useTranslation } from 'react-i18next'

const NodeInfo = (props) => {
  const { t } = useTranslation('pro/causeReport')
  const { title, serviceName, id, startTime, url, p90 = 0, totalTime = 0 } = props

  return (
    <Card type="inner" classNames={{ body: 'px-3 py-2' }} className="mb-2">
      <div className="font-bold text-base mb-1">{title}</div>
      <div className="w-full-with-children">
        <div className="grid">
          <div>
            {t('serviceName')}: {serviceName}
          </div>
          <div>
            {t('instanceName')}: {id}
          </div>
          <div>
            {t('requestTime')}:{' '}
            {startTime && convertTime(startTime / 1000, 'yyyy-mm-dd hh:mm:ss.SSS')}
          </div>
          <div style={{ gridColumn: 'span 3' }}>
            {t('requestEndpoint')}: {url}
          </div>
        </div>
      </div>
      <TimeProgress p90={p90} error={totalTime} />
    </Card>
  )
}

export default NodeInfo
