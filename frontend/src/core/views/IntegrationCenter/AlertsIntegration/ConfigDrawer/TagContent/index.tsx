/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import Title from 'antd/es/typography/Title'
import Typography from 'antd/es/typography/Typography'
import TagContactRule from './TagContactRule'
import { Card } from 'antd'
interface TagContentProps {
  sourceId: string
}
const TagContent = ({ sourceId }: TagContentProps) => {
  return (
    <>
      <Card className="bg-[#202023] rounded-3xl mt-4" classNames={{ body: 'px-4 py-3' }}>
        <Typography>
          <Title level={5}>标签增强</Title>

          <TagContactRule sourceId={sourceId} />
        </Typography>
      </Card>
    </>
  )
}
export default TagContent
