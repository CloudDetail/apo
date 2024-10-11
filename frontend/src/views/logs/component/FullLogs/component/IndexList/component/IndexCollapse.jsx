import { Collapse, ConfigProvider, List } from 'antd'
import React, { useEffect, useState } from 'react'
import { useLogsContext } from 'src/contexts/LogsContext'
import IndexCollapseItem from './IndexCollapseItem'
import { AiFillCaretDown, AiFillCaretRight, AiFillCaretUp } from 'react-icons/ai'

const IndexCollapse = (props) => {
  const { type } = props
  const [items, setItems] = useState([])
  const { defaultFields, hiddenFields, query = '', updateQuery } = useLogsContext()

  useEffect(() => {
    const fields = type === 'base' ? defaultFields : hiddenFields
    setItems(
      // (fields ?? []).map((item) => {
      //   return { key: item, label: item, children: <IndexCollapseItem /> }
      // }),
      fields,
    )
  }, [type, defaultFields, hiddenFields])

  const changeCollapse = (e) => {
    console.log(e)
  }
  const clickIndex = (e) => {
    let newQuery = query
    if (query.length > 0) {
      newQuery += ' And `' + e + '` ='
    } else {
      newQuery += '`' + e + '` ='
    }
    updateQuery(newQuery)
  }
  return (
    <>
      {/* <Collapse
        items={items}
        bordered={false}
        expandIconPosition="end"
        onChange={changeCollapse}
        expandIcon={({ isActive }) => (isActive ? <AiFillCaretUp /> : <AiFillCaretDown />)}
      /> */}
      <List
        dataSource={items}
        bordered={false}
        renderItem={(item) => (
          <List.Item key={item} onClick={() => clickIndex(item)} className=" cursor-pointer">
            {item}
          </List.Item>
        )}
      />
    </>
  )
}
export default IndexCollapse
