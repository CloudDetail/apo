import { Image, theme } from 'antd'
import i18next, { t } from 'i18next'
import mock1Png from 'src/pro/assets/1.png'
import mock2Png from 'src/pro/assets/2.png'
import mock5Png from 'src/pro/assets/5.png'
import mock61Png from 'src/pro/assets/6-1.png'
import mock62Png from 'src/pro/assets/6-2.png'
import mock7Png from 'src/pro/assets/7.png'
import mock8Png from 'src/pro/assets/8.png'
import mock1PngEn from 'src/pro/assets/en/1.png'
import mock2PngEn from 'src/pro/assets/en/2.png'
import mock5PngEn from 'src/pro/assets/en/5.png'
import mock61PngEn from 'src/pro/assets/en/6-1.png'
import mock62PngEn from 'src/pro/assets/en/6-2.png'
import mock7PngEn from 'src/pro/assets/en/7.png'
import mock8PngEn from 'src/pro/assets/en/8.png'
import {
  ServiceAnalyzeProvider,
  useServiceAnalyzeContext,
} from 'src/pro/contexts/ServiceAnalyzeContext'
import TimeLineChart from '../../serviceAnalyze/DescendantTopology/TimeLineChart'
import AnormalTypeFilter from '../../serviceAnalyze/DescendantTopology/AnormalTypeFilter'
import TopologyCom from '../../serviceAnalyze/DescendantTopology/Topology'
import AnormalList from '../../serviceAnalyze/DescendantTopology/NodeInfo/AnormalList'
import PlayDelta from '../../serviceAnalyze/DescendantTopology/PlayDelta'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next'
const i18nKey = 'pro/chatAI:explanations'
const pngLagnguageMap = {
  topologyExample: {
    en: mock1PngEn,
    zh: mock1Png,
  },
  alertAnalysisTimeRange: {
    en: mock2PngEn,
    zh: mock2Png,
  },
  upstream: {
    en: mock5PngEn,
    zh: mock5Png,
  },
  alertsFirst: {
    en: mock61PngEn,
    zh: mock61Png,
  },
  alertsSecond: {
    en: mock62PngEn,
    zh: mock62Png,
  },
  rootCause: {
    en: mock7PngEn,
    zh: mock7Png,
  },
  summary: {
    en: mock8PngEn,
    zh: mock8Png,
  },
}
const language = i18next.language
const wrapperClassName =
  'w-full overflow-hidden border p-2 rounded  mb-3 text-xs bg-[var(--ant-color-bg-layout)]'
const ServiceAnalyze = () => {
  const { loading, selectedNodeInfo } = useServiceAnalyzeContext()
  const { t } = useTranslation('pro/serviceAnalyze')
  const { useToken } = theme
  const { token } = useToken()
  return (
    <div className="relative">
      <LoadingSpinner loading={loading} />
      <div className={wrapperClassName} style={{ background: token.colorBgLayout }}>
        <div className="flex flex-row h-[200px]">
          <div className="flex-grow flex-shrink w-2/3">
            <TimeLineChart />
          </div>
          <div className="h-full flex flex-col justify-between w-1/3 text-xs">
            <AnormalTypeFilter />
          </div>
        </div>

        <div
          style={{
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'center',
            justifyContent: 'center',
          }}
        >
          <PlayDelta />
        </div>
      </div>
      <div className={`h-[300px] ${wrapperClassName}`}>
        <TopologyCom />
      </div>
      <div className={`h-[400px] ${wrapperClassName}`}>
        <div className="text-sm mb-1">
          {t(`node.alertList`)} (
          {selectedNodeInfo?.endpoint ? (
            <span>{t(`node.selected`)}</span>
          ) : (
            <span>{t(`node.all`)}</span>
          )}
          {t(`node.nodeTimeData`)})
        </div>
        <AnormalList scrollHeight={400} />
      </div>
    </div>
  )
}
export const Explanations = [
  <ServiceAnalyzeProvider>
    <ServiceAnalyze />
  </ServiceAnalyzeProvider>,
  // <>
  //   <div>{t(`${i18nKey}.alertAnalysisTimeRange`)}</div>
  //   <Image src={pngLagnguageMap.alertAnalysisTimeRange[language]} width={'90%'} />
  // </>,
  <>{t(`${i18nKey}.identifySuspectedRootCause`)}</>,
  <>
    <div>{t(`${i18nKey}.alertDetails`)}</div>
    <ol style={{ listStyleType: 'disc', paddingLeft: 20 }}>
      <li>{t(`${i18nKey}.applicationAlert`)}</li>
      <li>{t(`${i18nKey}.containerAlert`)}</li>
      <li>{t(`${i18nKey}.infrastructureAlert`)}</li>
      <li>{t(`${i18nKey}.networkAlert`)}</li>
      <li>{t(`${i18nKey}.errorAlert`)}</li>
    </ol>
    <div>{t(`${i18nKey}.expertInsight`)}</div>
  </>,
  <>
    <div>{t(`${i18nKey}.alertPropagationExample`)}</div>
    <div>{t(`${i18nKey}.upstreamNotRootCause`)}</div>
    <div>{t(`${i18nKey}.alertPropagationProcess`)}</div>
    <Image src={pngLagnguageMap.upstream[language]} width={'100%'} />
  </>,
  <>
    <div>{t(`${i18nKey}.intermittentAlerts`)}</div>
    <div>{t(`${i18nKey}.intermittentAlertDefinition`)}</div>
    <div>{t(`${i18nKey}.alertOccurredAt`)}</div>
    <Image src={pngLagnguageMap.alertsFirst[language]} width={'100%'} />
    <div>{t(`${i18nKey}.alertResolvedAt`)}</div>
    <Image src={pngLagnguageMap.alertsSecond[language]} width={'100%'} />
  </>,
  <>
    <div>{t(`${i18nKey}.rootCauseAlertPropagation`)}</div>
    <div>{t(`${i18nKey}.alertPropagationBreak`)}</div>
    <Image src={pngLagnguageMap.rootCause[language]} width={'100%'} />
  </>,
  <>
    <div>{t(`${i18nKey}.summary`)}</div>
    <div>{t(`${i18nKey}.aiModelAccuracy`)}</div>
    <div>{t(`${i18nKey}.configurableAiModel`)}</div>
    <Image src={pngLagnguageMap.summary[language]} width={'100%'} />
  </>,
]
