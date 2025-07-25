import { Drawer } from 'antd'
import { dataplaneIconMap } from 'src/pro/constants'
import { DataplaneItem, DataplaneTypeItem } from 'src/pro/types/datasource'
import DataplaneConfigForm from 'src/pro/components/Dataplane/DataplaneConfigForm'

const ConfigDrawer = ({
  drawerOpen,
  closeDrawer,
  dataplaneConfigInfo,
}: {
  drawerOpen: boolean
  closeDrawer: () => void
  dataplaneConfigInfo: DataplaneTypeItem | DataplaneItem
}) => {
  return (
    <Drawer
      title={
        dataplaneConfigInfo && (
          <div className="flex items-center gap-2">
            <img
              src={dataplaneIconMap[dataplaneConfigInfo.typeName]}
              alt={dataplaneConfigInfo.typeName}
              className="w-6 h-6"
            />
            {dataplaneConfigInfo.typeName}
          </div>
        )
      }
      closable={{ 'aria-label': 'Close Button' }}
      onClose={() => closeDrawer()}
      open={drawerOpen}
      width={'50%'}
    >
      {dataplaneConfigInfo && (
        <DataplaneConfigForm
          dataplaneParams={dataplaneConfigInfo}
          onSave={() => {
            closeDrawer()
          }}
          readonly={false}
        />
      )}
    </Drawer>
  )
}

export default ConfigDrawer
