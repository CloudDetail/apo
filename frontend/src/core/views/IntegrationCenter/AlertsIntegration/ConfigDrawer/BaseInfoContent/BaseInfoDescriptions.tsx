/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Descriptions, DescriptionsProps, Form } from 'antd'
import Text from 'antd/es/typography/Text'
import { t } from 'i18next'
import { useTranslation } from 'react-i18next'

interface BaseInfoDescriptionsProps {
  sourceName?: string
  sourceId?: string
  clusters?: any[]
}
const BaseInfoDescriptions = ({
  sourceName,
  sourceId,
  clusters = [],
}: BaseInfoDescriptionsProps) => {
  const { t } = useTranslation('core/alertsIntegration')
  const form = Form.useFormInstance()
  const getPublishUrl = () => {
    const baseUrl = window.location.origin + '/api/alertinput/event/source?sourceId='
    if (sourceId) {
      return baseUrl + sourceId
    } else {
      return baseUrl + form.getFieldValue('sourceId')
    }
  }
  const items: DescriptionsProps['items'] = [
    {
      key: '1',
      label: t('sourceName'),
      children: sourceName,
      span: 'filled',
    },
    // {
    //   key: '2',
    //   label: '集群',
    //   children: (
    //     <>
    //       {clusters?.length > 0 ? (
    //         clusters.map((cluster) => cluster.name).join('、')
    //       ) : (
    //         <span className="text-gray-500">无</span>
    //       )}
    //     </>
    //   ),
    //   span: 'filled',
    // },
    {
      key: '2',
      label: t('pushUrl'),
      children: (
        // <div className="flex">
        //   <span className="mr-3">{getPublishUrl()}</span>
        //   <CopyButton value={getPublishUrl()}></CopyButton>
        // </div>
        <Text copyable={{ text: getPublishUrl }}>{getPublishUrl()}</Text>
      ),
      span: 'filled',
    },
  ]
  return <Descriptions title={t('basicInfo')} items={items} />
}
export default BaseInfoDescriptions
