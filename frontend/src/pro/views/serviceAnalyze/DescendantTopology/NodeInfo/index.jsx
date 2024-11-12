import { Collapse } from 'antd'
import { useEffect, useState } from 'react'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import AnormalList from './AnormalList'
import ErrorInstanceInfo from 'src/pro/views/serviceAnalyze/DescendantTopology/NodeInfo/ErrorInstanceInfo'
import PolarisMetricsInfo from 'src/pro/views/serviceAnalyze/DescendantTopology/NodeInfo/PolarisMetricsInfoModal/PolarisMetricsInfo'

const NodeInfo = () => {
  const [viewportHeight, setViewportHeight] = useState(window.innerHeight)
  const [activeItem, setActiveItem] = useState([])
  const { selectedNodeInfo } = useAlertAnalyzeContext()
  const [errorInstanceOpen, setErrorInstanceOpen] = useState(false)

  useEffect(() => {
    if (activeItem.includes('2')) {
      setErrorInstanceOpen(true)
    } else {
      setErrorInstanceOpen(false)
    }
  }, [activeItem])

  const items = [
    {
      key: '1',
      label: (
        <div className="flex items-center justify-between">
          <div className="flex items-center">延时报告</div>
        </div>
      ),
      children: (
        <PolarisMetricsInfo
          service={selectedNodeInfo?.service}
          endpoint={selectedNodeInfo?.endpoint}
        />
      ),
      forceRender: true,
      styles: {
        body: {
          paddingRight: 0,
        },
      },
    },
    {
      key: '2',
      label: '错误报告',
      children: (
        <ErrorInstanceInfo
          service={selectedNodeInfo?.service}
          endpoint={selectedNodeInfo?.endpoint}
          showLogs={false}
          errorInstanceOpen={errorInstanceOpen}
        />
      ),
      forceRender: true,
      styles: {
        body: {
          paddingRight: 0,
        },
      },
    },
    {
      key: '3',
      label: (
        <div>
          告警事件列表 ({selectedNodeInfo?.endpoint ? <span>当前选中</span> : <span>当前所有</span>}
          节点播放时间段数据)
        </div>
      ),
      children: <AnormalList />,
      forceRender: true,
      styles: {
        body: {
          height: viewportHeight - 150,
          paddingRight: 0,
        },
      },
    },
  ]
  return (
    <>
      <Collapse
        className="h-full relative"
        items={items}
        defaultActiveKey={['3']}
        onChange={(value) => setActiveItem(value)}
      />
    </>
  )
}
export default NodeInfo
