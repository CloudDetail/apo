/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Descriptions, DescriptionsProps } from 'antd'
interface BaseInfoDescriptionsProps {
  sourceName?: string
  clusters?: any[]
}
const BaseInfoDescriptions = ({ sourceName, clusters = [] }: BaseInfoDescriptionsProps) => {
  const items: DescriptionsProps['items'] = [
    {
      key: '1',
      label: '告警接入名',
      children: sourceName,
      span: 'filled',
    },
    {
      key: '2',
      label: '集群',
      children: (
        <>
          {clusters?.length > 0 ? (
            clusters.map((cluster) => cluster.name).join('、')
          ) : (
            <span className="text-gray-500">无</span>
          )}
        </>
      ),
      span: 'filled',
    },
  ]
  return <Descriptions title="基础信息" items={items} />
}
export default BaseInfoDescriptions
