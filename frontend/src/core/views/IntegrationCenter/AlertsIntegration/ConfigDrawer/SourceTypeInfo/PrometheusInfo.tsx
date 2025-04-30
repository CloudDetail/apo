/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { Typography } from 'antd'
import { useTranslation } from 'react-i18next'
import Paragraph from 'antd/es/typography/Paragraph'
import Title from 'antd/es/typography/Title'
import Text from 'antd/es/typography/Text'
import CopyPre from 'src/core/components/CopyPre'

const PrometheusInfo = () => {
  const { t } = useTranslation('core/alertsIntegration')

  const code1 = `global:
    ...
route:
    ...
receivers:
    ...
    - name: apo-collector
      webhook_configs:
        - send_resolved: true
          url: '<${t('prometheusDoc.webhookUrl')}>'`

  const code2 = `global:
    ...
route:
    receiver: xxx
    continue: false
    routes:
        - receiver: apo-collector
          continue: true`

  const code3 = `global:
    ...
route:
    receiver: apo-collector
    continue: false`

  return (
    <Typography>
      <Title level={4}>{t('prometheusDoc.title')}</Title>
      <Text>{t('prometheusDoc.description')}</Text>
      <Typography>{t('prometheusDoc.configInstructions')}</Typography>

      <Title level={5}>{t('prometheusDoc.step1.title')}</Title>
      <Text>{t('prometheusDoc.step1.description')}</Text>
      <CopyPre code={code1} />

      <Title level={5}>{t('prometheusDoc.step2.title')}</Title>
      <Text>{t('prometheusDoc.step2.description')}</Text>
      <CopyPre code={code2} />

      <Text>{t('prometheusDoc.step2.alternative')}</Text>
      <CopyPre code={code3} />

      <Text>{t('prometheusDoc.step2.note')}</Text>
      <ol>
        <li>{t('prometheusDoc.step2.order1')}</li>
        <li>{t('prometheusDoc.step2.order2')}</li>
      </ol>
      <Text>{t('prometheusDoc.step2.warning')}</Text>

      <Title level={5}>{t('prometheusDoc.step3.title')}</Title>
      <Title level={5}>{t('prometheusDoc.step4.title')}</Title>
      <Text>{t('prometheusDoc.step4.description')}</Text>

      <Title level={5}>{t('prometheusDoc.step5.title')}</Title>
    </Typography>
  )
}

export default PrometheusInfo
