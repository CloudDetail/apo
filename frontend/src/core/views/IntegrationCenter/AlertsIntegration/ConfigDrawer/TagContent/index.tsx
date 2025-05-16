/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import Title from 'antd/es/typography/Title'
import Typography from 'antd/es/typography/Typography'
import TagContactRule from './TagContactRule'
import { Alert, Card } from 'antd'
import { useTranslation } from 'react-i18next'
interface TagContentProps {
  sourceId: string
}
const TagContent = ({ sourceId }: TagContentProps) => {
  const { t } = useTranslation('core/alertsIntegration')
  return (
    <>
      <Card
        className="bg-[var(--ant-color-bg-layout)] rounded-3xl mt-4"
        classNames={{ body: 'px-4 py-3' }}
      >
        <Typography>
          <Title level={5} className="flex items-center">
            {t('rulesTitle')}
          </Title>

          <Alert message={t('rulesAlert')} type="warning" showIcon className="text-xs mb-2 mx-0" />
          <TagContactRule sourceId={sourceId} />
        </Typography>
      </Card>
    </>
  )
}
export default TagContent
