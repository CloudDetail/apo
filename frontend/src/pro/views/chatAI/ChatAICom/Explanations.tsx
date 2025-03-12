import { Image } from 'antd'
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
export const Explanations = [
  <div>
    <div>{t(`${i18nKey}.topologyExample`)}</div>
    <Image src={pngLagnguageMap.topologyExample[language]} width={'60%'} />
  </div>,
  <>
    <div>{t(`${i18nKey}.alertAnalysisTimeRange`)}</div>
    <Image src={pngLagnguageMap.alertAnalysisTimeRange[language]} width={'90%'} />
  </>,
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
