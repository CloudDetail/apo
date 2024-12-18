import { Button, Card, Collapse, Input, Tooltip } from 'antd'
import React, { useEffect } from 'react'
import IndexCollapse from './component/IndexCollapse'
import './index.css'
import FullTextSearch from '../SerarchBar/RawLogQuery/FullTextSearch'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import { IoEye } from "react-icons/io5";
import { IoMdEyeOff } from "react-icons/io";
const IndexList = () => {
  const {
    defaultFields,
    hiddenFields,
    displayFields,
    resetDisplayFields
  } = useLogsContext()

  const IndexType = {
    base: '基础字段',
    log: '日志字段',
  }

  const showHiddenAll = (event, type, field) => {
    event.stopPropagation()
    if (field === 'tags') {
      if (type === 'show') {
        resetDisplayFields([...new Set([...displayFields, ...defaultFields])])
      } else {
        resetDisplayFields(displayFields.filter((item) => !defaultFields.includes(item)))
      }
    } else {
      if (type === 'show') {
        resetDisplayFields([...new Set([...displayFields, ...hiddenFields])])
      } else {
        resetDisplayFields(displayFields.filter((item) => !hiddenFields.includes(item)))
      }
    }
  }

  const items = [
    {
      key: 'base',
      label: (<div className='flex items-center justify-between'>
        <span className='select-none'>标签</span>
        <div className='flex items-center'>
          {
            defaultFields.every(item => displayFields.includes(item)) ?
              <Tooltip title='取消展示' mouseEnterDelay={0.5}>
                <Button type='link' className='p-0 m-0 h-auto' icon={<IoMdEyeOff size={18} />} onClick={(e) => showHiddenAll(e, 'hidden', 'tags')}></Button>
              </Tooltip>
              :
              <Tooltip title='全部展示' mouseEnterDelay={0.5}>
                <Button type='link' className='p-0 m-0 h-auto' icon={<IoEye size={18} />} onClick={(e) => showHiddenAll(e, 'show', 'tags')}></Button>
              </Tooltip>
          }
        </div>
      </div>
      ),
      children: <IndexCollapse type="base" />,
    },
    {
      key: 'log',
      label: (<div className='flex justify-between'>
        <span className='select-none'>日志字段</span>
        <div className='flex items-center'>
          {
            hiddenFields.every(item => displayFields.includes(item)) ?
              <Tooltip title='全不展示' mouseEnterDelay={0.5}>
                <Button type='link' className='p-0 m-0 h-auto' icon={<IoMdEyeOff size={18} />} onClick={(e) => showHiddenAll(e, 'hidden', 'logs')}></Button>
              </Tooltip>
              :
              <Tooltip title='全部展示' mouseEnterDelay={0.5}>
                <Button type='link' className='p-0 m-0 h-auto' icon={<IoEye size={18} />} onClick={(e) => showHiddenAll(e, 'show', 'logs')}></Button>
              </Tooltip>
          }
        </div>
      </div>
      ),
      children: <IndexCollapse type="log" />,
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
        className=" flex-1 indexList h-full overflow-auto mb-2 mt-2"
      />
    </div>
  )
}
export default IndexList
