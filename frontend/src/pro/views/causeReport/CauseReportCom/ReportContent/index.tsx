import { Card, Empty } from 'antd'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'
import CauseStep from './CauseStep'

const ReportContent = () => {
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
        loading ||
        (isTimeOut ? <Empty description="报告请求超时，请刷新重试。"></Empty> : <Empty />)
      )}
    </>
  )
}
export default ReportContent
