import { Collapse, ConfigProvider } from 'antd'
import Summary from './components/Summary/index'
import SupportingEvidence from './components/SupportingEvidence'
import IconWrapper from './components/IconWrapper'
import { LuSquareCheckBig, LuWebhook, LuWorkflow } from 'react-icons/lu'
import { BiAnalyse } from 'react-icons/bi'
import { SlMagnifier, SlArrowDown } from 'react-icons/sl'
import AffectedEndpoints from './components/AffectedEndpoints'
import RecommendedAction from './components/RecommendedAction'
import RootCauseAnalysis from './components/RootCauseAnalysis'
import AnalysisProcess from './components/AnalysisProcess'
import { useCallback, useState } from 'react'
import styles from './index.module.scss'
import { getReportApi } from 'src/pro/api/report'
import Empty from 'src/core/components/Empty/Empty'
import { useDebounce } from 'react-use'
import { useSearchParams } from 'react-router-dom'
import NoTraceHint from './components/NoTraceHint'
import { useTranslation } from 'react-i18next'
import LoadingSpinner from 'src/core/components/Spinner'
const CollapseLabel = ({ icon, title }) => {
  return (
    <div className="flex items-center ">
      <IconWrapper icon={icon} size={38} />
      <div className="text-base font-bold pl-2">{title}</div>
    </div>
  )
}
function Report() {
  const { t } = useTranslation('pro/report')
  const [activeTabKey, setActiveTabKey] = useState(['rootCauseAnalysis', 'recommendedAction'])
  const [searchParams] = useSearchParams()
  const [reportData, setReportData] = useState()
  const [reportType, setReportType] = useState()
  const [loading, setLoading] = useState(true)
  const startTime = Number(searchParams.get('startTime'))
  const endTime = Number(searchParams.get('endTime'))
  const getReportInfo = () => {
    setLoading(true)
    const alertEventId = searchParams.get('alertEventId')
    const agentChatId = searchParams.get('agentChatId')
    const tags = alertEventId ? { alertEventId } : { agentChatId }
    getReportApi({ startTime: startTime, endTime: endTime, tags })
      .then((res) => {
        setReportData(res.reportData)
        setReportType(res.reportType)
      })
      .catch((error) => {
        setLoading(false)
      })
      .finally(() => {
        setLoading(false)
      })
  }

  useDebounce(
    () => {
      getReportInfo()
    },
    300,
    [searchParams],
  )
  const applicationItems = useCallback(
    () => [
      {
        key: 'affectedEndpoints',
        label: (
          <CollapseLabel
            icon={<LuWebhook className="w-full h-full" />}
            title={t('affectedEndpoints')}
          />
        ),
        children: reportData?.topology ? (
          <AffectedEndpoints topology={reportData.topology} />
        ) : (
          <NoTraceHint />
        ),
      },
      {
        key: 'rootCauseAnalysis',
        label: (
          <CollapseLabel
            icon={<BiAnalyse className="w-full h-full" />}
            title={t('rootCauseAnalysis')}
          />
        ),
        children: <RootCauseAnalysis {...reportData?.rootCauseAnalysis} />,
      },
      {
        key: 'recommendedAction',
        label: (
          <CollapseLabel
            icon={<LuSquareCheckBig className="w-full h-full" />}
            title={t('recommendedAction')}
          />
        ),
        children: <RecommendedAction suggest={reportData?.suggest} />,
      },
      {
        key: 'analysisProcess',
        label: (
          <CollapseLabel
            icon={<LuWorkflow className="w-full h-full" />}
            title={t('analysisProcess')}
          />
        ),
        children: <AnalysisProcess analyzeProcess={reportData?.analyzeProcess} />,
      },
      {
        key: 'supportingEvidence',
        label: (
          <CollapseLabel
            icon={<SlMagnifier className="w-full h-full" />}
            title={t('supportingEvidence')}
          />
        ),
        children: (
          <SupportingEvidence
            evidence={reportData?.evidence}
            startTime={startTime}
            endTime={endTime}
          />
        ),
      },
    ],
    [reportData],
  )
  return (
    <div className="w-full">
      {loading ? (
        <LoadingSpinner loading={loading} />
      ) : reportType && reportData ? (
        <ConfigProvider
          theme={{
            components: {
              Statistic: {
                contentFontSize: 16,
                titleFontSize: 12,
              },
            },
          }}
        >
          <Summary
            type={reportType}
            overview={reportData?.overview}
            cause={reportData?.rootCauseAnalysis?.cause}
          />
          <>
            {applicationItems()?.map((item) => (
              <Collapse
                key={item.key}
                items={[item]}
                defaultActiveKey={[item.key]}
                className={styles.collapse}
                bordered={false}
                expandIconPosition={'end'}
                onChange={setActiveTabKey}
                activeKey={activeTabKey}
                expandIcon={({ isActive }) => {
                  return isActive ? (
                    <SlArrowDown className={'rotate-180'} size={18} />
                  ) : (
                    <SlArrowDown className={'rotate-0'} size={18} />
                  )
                }}
              />
            ))}
          </>
        </ConfigProvider>
      ) : (
        <Empty width={'20%'} />
      )}
    </div>
  )
}
export default Report
