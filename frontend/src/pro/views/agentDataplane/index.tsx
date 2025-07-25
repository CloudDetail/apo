import { Divider, Flex } from 'antd'
import { useState } from 'react'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import DataplaneList from 'src/pro/components/Dataplane/DataplaneList'
import ConfiguredSources from './components/ConfiguredSources'
import ConfigDrawer from './ConfigDrawer'
import { DataplaneItem, DataplaneTypeItem } from 'src/pro/types/datasource'

const AgentDataplanePage = () => {
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [dataplaneConfigInfo, setDatasourceConfigInfo] = useState<
    DataplaneItem | DataplaneTypeItem | null
  >(null)
  const [key, setKey] = useState(0)
  const openAddDrawer = (dataplaneConfigInfo: DataplaneTypeItem) => {
    setDrawerOpen(true)
    setDatasourceConfigInfo(dataplaneConfigInfo)
  }
  const openEditDrawer = (dataplaneConfigInfo: DataplaneItem) => {
    setDrawerOpen(true)
    setDatasourceConfigInfo(dataplaneConfigInfo)
  }
  const closeDrawer = () => {
    setKey((prevKey) => prevKey + 1)
    setDrawerOpen(false)
    setDatasourceConfigInfo(null)
  }
  return (
    <>
      <BasicCard>
        <Flex
          style={{
            // boxShadow: '0 0 10px rgba(0, 0, 0, 0.1)',
            // padding: 10,
            height: '100%',
          }}
        >
          <div className="h-full w-[310px] px-1 flex-shrink-0">
            <DataplaneList onClick={openAddDrawer} />
          </div>
          <Divider type="vertical" className="h-full text-[var(--ant-color-border)]" />
          <div className="px-4 flex-1 overflow-hidden">
            <ConfiguredSources openDrawer={openEditDrawer} key={key} />
          </div>
        </Flex>
      </BasicCard>
      <ConfigDrawer
        drawerOpen={drawerOpen}
        closeDrawer={closeDrawer}
        dataplaneConfigInfo={dataplaneConfigInfo}
      />
    </>
  )
}

export default AgentDataplanePage
