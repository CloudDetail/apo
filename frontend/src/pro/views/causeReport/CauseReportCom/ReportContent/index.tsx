import { Card, Empty } from 'antd'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'
import CauseStep from './CauseStep'
import { useTranslation } from 'react-i18next'

const ReportContent = () => {
  const { t } = useTranslation('pro/causeReport')
  const { reportData, category, loading, isTimeOut } = useCauseReportContext()
  return (
    <>
      {reportData?.reports ? (
        <>
          {reportData.reports.map((item: any, idx: number) => (
            <CauseStep data={item} occurrenceAt={reportData?.metadata?.timestamp} key={idx} />
          ))}
        </>
      ) : (
        loading || (isTimeOut ? <Empty description={t('reportTimeout')}></Empty> : <Empty />)
      )}
    </>
  )
}
export default ReportContent
