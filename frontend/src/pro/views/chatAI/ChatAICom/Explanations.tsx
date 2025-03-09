import { Image } from 'antd'
import { t } from 'i18next'
import mock1Png from 'src/pro/assets/1.png'
import mock2Png from 'src/pro/assets/2.png'
import mock5Png from 'src/pro/assets/5.png'
import mock61Png from 'src/pro/assets/6-1.png'
import mock62Png from 'src/pro/assets/6-2.png'
import mock7Png from 'src/pro/assets/7.png'
import mock8Png from 'src/pro/assets/8.png'
const i18nKey = 'pro/chatAI:explanations'
export const Explanations = [
  <div>
    <div>{t(`${i18nKey}.topologyExample`)}</div>
    <Image src={mock1Png} width={'60%'} />
  </div>,
  <>
    <div>{t(`${i18nKey}.alertAnalysisTimeRange`)}</div>
    <Image src={mock2Png} width={'90%'} />
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
    <Image src={mock5Png} width={'100%'} />
  </>,
  <>
    <div>{t(`${i18nKey}.intermittentAlerts`)}</div>
    <div>{t(`${i18nKey}.intermittentAlertDefinition`)}</div>
    <div>{t(`${i18nKey}.alertOccurredAt`)}</div>
    <Image src={mock61Png} width={'100%'} />
    <div>{t(`${i18nKey}.alertResolvedAt`)}</div>
    <Image src={mock62Png} width={'100%'} />
  </>,
  <>
    <div>{t(`${i18nKey}.rootCauseAlertPropagation`)}</div>
    <div>{t(`${i18nKey}.alertPropagationBreak`)}</div>
    <Image src={mock7Png} width={'100%'} />
  </>,
  <>
    <div>{t(`${i18nKey}.summary`)}</div>
    <div>{t(`${i18nKey}.aiModelAccuracy`)}</div>
    <div>{t(`${i18nKey}.configurableAiModel`)}</div>
    <Image src={mock8Png} width={'100%'} />
  </>,
]
