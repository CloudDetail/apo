import { Collapse } from 'antd'
import React, { useEffect, useState } from 'react'
import { useLogsContext } from 'src/contexts/LogsContext'

const IndexCollapseItem = (props) => {
  const { type } = props
  const [items, setItems] = useState([])
  const { defaultFields, hiddenFields } = useLogsContext()
  return <>1</>
}
export default IndexCollapseItem
