import { Tag } from 'antd'
import React from 'react'

const ServiceEndpointTag = ({ service, endpoint }) => {
  return (
    <Tag
      key={service + endpoint}
      className="w-full flex-shrink-0 inline-block  overflow-hidden whitespace-nowrap text-ellipsis "
    >
      <div>
        <div className="overflow-hidden whitespace-break-spaces">
          <span className="text-gray-400">应用名：</span>
          {service}
        </div>
        {endpoint && (
          <div className="overflow-hidden whitespace-break-spaces">
            <span className="text-gray-400">端点：</span>
            {endpoint}
          </div>
        )}
      </div>
    </Tag>
  )
}
export default ServiceEndpointTag
