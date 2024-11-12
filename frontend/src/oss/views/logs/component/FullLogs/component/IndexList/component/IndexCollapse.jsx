import { Collapse, ConfigProvider, Empty, List } from 'antd'
import React, { useEffect, useState } from 'react'
import { useLogsContext } from 'src/core/contexts/LogsContext'
import IndexCollapseItem from './IndexCollapseItem'
import { AiFillCaretDown, AiFillCaretRight, AiFillCaretUp } from 'react-icons/ai'

const IndexCollapse = (props) => {
  const { type } = props
  const [items, setItems] = useState([])
  const { defaultFields, hiddenFields, query = '', updateQuery } = useLogsContext()

  useEffect(() => {
    const fields = type === 'base' ? defaultFields : hiddenFields
    setItems(
      (fields ?? []).map((item) => {
        return { key: item, label: item, children: <IndexCollapseItem field={item} /> }
      }),
    )
  }, [type, defaultFields, hiddenFields])

  return (
    <>
      <Collapse
        items={items}
        bordered={false}
        expandIconPosition="end"
        ghost
        className="h-full overflow-y-auto overflow-x-hidden"
        style={{ maxHeight: 'calc((100vh - 265px) / 2)', overflowY: 'hidden' }}
        expandIcon={({ isActive }) => (isActive ? <AiFillCaretUp /> : <AiFillCaretDown />)}
      />
      {items.length === 0 && <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />}
      {/* <List
        dataSource={items}
        bordered={false}
        renderItem={(item) => (
          <List.Item key={item} onClick={() => clickIndex(item)} className=" cursor-pointer">
            {item}
          </List.Item>
        )}
      /> */}
    </>
  )
}
export default IndexCollapse
