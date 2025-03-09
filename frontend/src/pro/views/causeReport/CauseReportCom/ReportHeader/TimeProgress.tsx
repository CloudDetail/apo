import React, { useEffect, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { formatUnit } from 'src/core/utils/time'
interface TimeProps {
  p90: number
  error: number
  width?: number
}
function TimeProgress({ p90, error }: TimeProps) {
  const { t } = useTranslation('pro/causeReport')
  const [p90WidthPercent, setP90WidthPercent] = useState<number>(0)
  const [errorWidthPercent, setErrorWidthPercent] = useState<number>(0)
  useEffect(() => {
    if (p90 > error) {
      setP90WidthPercent(80)
      setErrorWidthPercent((error / p90) * 80)
    } else {
      setP90WidthPercent((p90 / error) * 80)
      setErrorWidthPercent(80)
    }
  }, [p90, error])
  return (
    <div className={`duration_warp`}>
      <div className="duration_item">
        <span>{t('historicalBaseline')}</span>
        {p90 > 0 ? (
          <React.Fragment>
            <div className="health_line" style={{ width: `${p90WidthPercent}%` }}></div>
            <span>{formatUnit(p90, 'ns')}</span>
          </React.Fragment>
        ) : null}
      </div>
      <div className="duration_item">
        <span>{t('currentDuration')}</span>
        {error > 0 ? (
          <React.Fragment>
            <div className="abnormal_line" style={{ width: `${errorWidthPercent}%` }}></div>
            <span>{formatUnit(error, 'ns')}</span>
          </React.Fragment>
        ) : null}
      </div>
    </div>
  )
}

export default TimeProgress
