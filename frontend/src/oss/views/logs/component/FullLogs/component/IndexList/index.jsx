import { Card, Collapse } from 'antd'
import React from 'react'
import IndexCollapse from './component/IndexCollapse'
import './index.css'
import FullTextSearch from '../SerarchBar/RawLogQuery/FullTextSearch'
import { useTranslation } from 'react-i18next' // 引入i18n

const IndexList = () => {
  const { t } = useTranslation('oss/fullLogs') // 使用i18n
  const items = [
    {
      key: 'base',
      label: t('indexList.basicFieldLabel'),
      children: <IndexCollapse type="base" />,
    },
    {
      key: 'log',
      label: t('indexList.logFieldLabel'),
      children: <IndexCollapse type="log" />,

      style: {
        maxHeight: '50%',
        overflow: 'hidden',
      },
    },
  ]
  return (
    <div className="flex flex-col h-full">
      <div className="flex-shrink-0 py-2">
        <FullTextSearch />
      </div>
      <Collapse
        items={items}
        defaultActiveKey={['base', 'log']}
        size="small"
        className=" flex-1 indexList h-full overflow-hidden mb-2"
      />
    </div>
  )
}
export default IndexList
