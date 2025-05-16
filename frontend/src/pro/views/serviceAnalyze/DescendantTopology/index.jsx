import React, { useState } from 'react'
import TopologyCom from './Topology'
import AnormalTypeFilter from './AnormalTypeFilter'
import TimeLineChart from './TimeLineChart'
import { Card, Divider, Splitter, theme } from 'antd'
import { useServiceAnalyzeContext } from 'src/pro/contexts/ServiceAnalyzeContext'
import NodeInfo from './NodeInfo'
import PlayDelta from './PlayDelta'
import LoadingSpinner from 'src/core/components/Spinner'
import { useTranslation } from 'react-i18next'

const DescendantTopology = () => {
  const { t } = useTranslation('pro/serviceAnalyze')
  const { loading } = useServiceAnalyzeContext()

  const [rightWidth, setRightWidth] = useState(300)
  const onResize = (sizes) => {
    setRightWidth(sizes[1])
  }
  const { useToken } = theme
  const { token } = useToken()
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
              <div className="rounded p-2 m-2" style={{ background: token.colorBgLayout }}>
                <div className=" text-base font-bold flex items-center pl-1">{t('alertTotal')}</div>
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
