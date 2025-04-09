import { Tag } from 'antd'
import React from 'react'
import { useTranslation } from 'react-i18next'

const ServiceEndpointTag = ({ service, endpoint }) => {
  const { t } = useTranslation('common')
  return (
    <Tag
      key={service + endpoint}
      className="w-full flex-shrink-0 inline-block overflow-hidden whitespace-nowrap text-ellipsis"
    >
      <div className="flex flex-col">
        <div className="overflow-hidden">
          <span className="text-gray-400">{t('service')}：</span>
          {service}
        </div>
        {endpoint && (
          <div className="overflow-hidden whitespace-break-spaces">
            <span className="text-gray-400">{t('endpoint')}：</span>
            {endpoint}
          </div>
        )}
      </div>
    </Tag>
  )
}
export default ServiceEndpointTag
