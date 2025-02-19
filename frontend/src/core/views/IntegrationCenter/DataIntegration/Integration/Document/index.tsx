/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Card } from 'antd'
import { useTranslation } from 'react-i18next'

const IntegrationDoc = () => {
  const { t } = useTranslation('core/dataIntegration')
  return (
    <Card
      title={t('guide')}
      className="h-full overflow-hidden"
      classNames={{ body: 'p-0 overflow-auto flex flex-col' }}
      styles={{ body: { height: 'calc(100% - 60px)' } }}
    ></Card>
  )
}
export default IntegrationDoc
