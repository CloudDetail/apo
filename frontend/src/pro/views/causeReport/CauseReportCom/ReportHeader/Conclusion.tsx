import { Descriptions } from 'antd'
import { useEffect, useState } from 'react'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'

const Conclusion = () => {
  const { reportData } = useCauseReportContext()
  const [items, setItems] = useState([])
  useEffect(() => {
    let items = []
    let conclusionStruct = reportData?.conclusionStruct
    // conclusionStruct
    conclusionStruct?.metricDetails?.forEach((item) => {
      items.push({
        key: item.describe,
        label: item.describe,
        children: item.value,
      })
    })
    Object.entries(conclusionStruct?.contextDetails || {})?.forEach(([key, value]) => {
      items.push({
        key: key,
        label: key,
        children: value,
      })
    })
    conclusionStruct?.checked?.forEach((item) => {
      items.push({
        key: item.name,
        label: item.name,
        children: '正常',
      })
    })
    setItems(items)
  }, [reportData])
  return (
    <>
      <pre
        style={{
          color: '#f3485c',
          whiteSpace: 'break-spaces',
          wordBreak: 'break-all',
          overflow: 'hidden',
          marginBottom: 3,
        }}
      >
        {reportData?.conclusionStruct?.messages?.join('\n\n') || ''}
      </pre>
      <div className="flex items-center m-3">
        <Descriptions
          items={items}
          bordered
          column={1}
          size="small"
          //   labelStyle={{ width: 300 }}
          className="flex-grow-0 flex-shrink-0"
        />
      </div>
    </>
  )
}
export default Conclusion
