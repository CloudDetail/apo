import { Tag, Tooltip } from 'antd'
import React from 'react'
import { useTranslation } from 'react-i18next'

const TagInfo = ({ service, endpoint }) => {
  const { t } = useTranslation('common')
  return (
    <div className="flex flex-col">
      <div className="">
        <span className="text-gray-400">{t('service')}：</span>
        {service}
      </div>
      {endpoint && (
        <div className=" break-all">
          <span className="text-gray-400">{t('endpoint')}：</span>
          {endpoint}
        </div>
      )}
    </div>
  )
}
const ServiceEndpointTag = (props) => {
  const { service, endpoint } = props

  return (
    <Tooltip title={<TagInfo {...props} />} placement="topLeft">
      <div>
        <Tag
          key={service + endpoint}
          className="w-full flex-shrink-0 inline-block overflow-hidden whitespace-nowrap text-ellipsis"
        >
          <TagInfo {...props} />
        </Tag>
      </div>
    </Tooltip>
  )
}
export default ServiceEndpointTag
