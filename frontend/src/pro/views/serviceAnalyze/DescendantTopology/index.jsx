import React, { useEffect, useState } from 'react'
import TopologyCom from './Topology'
import AnormalTypeFilter from './AnormalTypeFilter'
import TimeLineChart from './TimeLineChart'
import { Card, Divider, Modal, Splitter } from 'antd'
import { useAlertAnalyzeContext } from 'src/pro/contexts/AlertAnalyzeContext'
import NodeInfo from './NodeInfo'
import PlayDelta from './PlayDelta'
import LoadingSpinner from 'src/core/components/Spinner'
import DateTimeRangePicker from 'src/core/components/DateTime/DateTimeRangePicker'
import { useSearchParams } from 'react-router-dom'

const DescendantTopology = () => {
  const { selectedEndpointInfo, setSelectedEndpointInfo, setSelectedNodeInfo, loading } =
    useAlertAnalyzeContext()
  const [searchParams] = useSearchParams()

  const [rightWidth, setRightWidth] = useState(300)
  const onResize = (sizes) => {
    setRightWidth(sizes[1])
  }
  useEffect(() => {
    const service = searchParams.get('service')
    const endpoint = searchParams.get('endpoint')
    setSelectedEndpointInfo({ service: service, endpoint: endpoint })
  }, [searchParams])
  return (
    <>
      <Card
        style={{ height: 'calc(100vh - 120px)' }}
        classNames={{
          body: 'p-0 ',
        }}
      >
        <LoadingSpinner loading={loading} />
        <Splitter
          style={{ height: 'calc(100vh - 120px)', boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)' }}
          onResize={onResize}
        >
          <Splitter.Panel defaultSize="60%">
            <TopologyCom />
          </Splitter.Panel>
          <Splitter.Panel className="h-full overflow-auto flex flex-col">
            <div className="flex-grow-0 flex-shrink-0 overflow-hidden">
              <div className="bg-[#2a3042] rounded p-2 m-2">
                <div className=" text-base font-bold flex items-center pl-1">告警和异常总数</div>
                <Divider className="m-2" />
                <div className=" flex flex-row overflow-hidden ">
                  <div className="flex-grow flex-shrink w-[400px]">
                    <TimeLineChart rightWidth={rightWidth} />
                  </div>
                  <div className="h-full flex flex-col justify-between w-[250px]">
                    <AnormalTypeFilter />
                  </div>
                </div>
              </div>

              <div
                style={{
                  display: 'flex',
                  flexDirection: 'column',
                  alignItems: 'center',
                  justifyContent: 'center',
                }}
              >
                <PlayDelta />
              </div>
            </div>
            <div className="flex-1">
              <NodeInfo />
            </div>
          </Splitter.Panel>
        </Splitter>
      </Card>
    </>
  )
}
export default DescendantTopology
