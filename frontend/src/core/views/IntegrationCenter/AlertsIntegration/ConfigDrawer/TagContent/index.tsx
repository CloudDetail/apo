/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import Title from 'antd/es/typography/Title'
import Typography from 'antd/es/typography/Typography'
import TagContactRule from './TagContactRule'
import { Alert, Card } from 'antd'
interface TagContentProps {
  sourceId: string
}
const TagContent = ({ sourceId }: TagContentProps) => {
  return (
    <>
      <Card className="bg-[#202023] rounded-3xl mt-4" classNames={{ body: 'px-4 py-3' }}>
        <Typography>
          <Title level={5} className="flex items-center">
            关联应用规则
          </Title>

          <Alert
            message="
              从输入的告警事件的标签(Tag/Label)中取出关联信息,用于告警分析时关联告警事件和受监控应用
              "
            type="warning"
            showIcon
            className="text-xs mb-2 mx-0"
          />
          <TagContactRule sourceId={sourceId} />
        </Typography>
      </Card>
    </>
  )
}
export default TagContent
