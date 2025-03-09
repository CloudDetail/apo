import { Button } from 'antd'
import LogsTabs from './LogsTabs'
import { useSelector } from 'react-redux'
import { selectSecondsTimeRange } from 'src/core/store/reducers/timeRangeReducer'
import { TimestampToISO } from 'src/core/utils/time'
import { useTranslation } from 'react-i18next'

const FaultSiteLogsWrapper = ({ pagination, list, selectedNodeInfo }: any) => {
  const { startTime, endTime } = useSelector(selectSecondsTimeRange)
  const {t} = useTranslation('pro/dataViewCom')
  return (
    <div>
      <div className="max-h-[460px] overflow-hidden flex">
        <div className="flex-1">
          <LogsTabs logsPageList={list} showFaultSite={false} />
        </div>
      </div>
      {pagination?.total > 10 && (
        <div className="flex-shrink-0 text-right">
          <Button
            type="primary"
            size="small"
            className="m-2 "
            target="_blank"
            href={`/#/logs/fault-site?service=${selectedNodeInfo.service}&endpoint=${selectedNodeInfo.endpoint}&logs-from=${TimestampToISO(startTime)}&logs-to=${TimestampToISO(endTime)}`}
          >
            {t('moreLog')}
          </Button>
        </div>
      )}
    </div>
  )
}
export default FaultSiteLogsWrapper
