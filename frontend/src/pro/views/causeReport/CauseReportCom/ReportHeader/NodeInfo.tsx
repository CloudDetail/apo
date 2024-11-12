import { Card } from 'antd'
import TimeProgress from './TimeProgress'
import { convertTime, formatUnit } from 'src/core/utils/time'

const NodeInfo = (props) => {
  const { title, serviceName, id, startTime, url, p90 = 0, totalTime = 0 } = props
  return (
    <Card type="inner" classNames={{ body: 'px-3 py-2' }} className="mb-2">
      <div className="font-bold text-base mb-1">{title}</div>
      <div className="w-full-with-children">
        <div className="grid">
          <div>服务名: {serviceName}</div>
          <div>实例名: {id}</div>
          <div>
            请求时间: {startTime && convertTime(startTime / 1000, 'yyyy-mm-dd hh:mm:ss.SSS')}
          </div>
          <div style={{ gridColumn: 'span 3' }}>请求端点: {url}</div>
        </div>
      </div>
      <TimeProgress p90={p90} error={totalTime} />
    </Card>
  )
}
export default NodeInfo
