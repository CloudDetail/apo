import { Divider, Typography } from 'antd'
import { useState } from 'react'
import DataplaneRadioList from './DataplaneRadioList'
import { DataplaneTypeItem } from 'src/pro/types/datasource'
import DataplaneConfigForm from 'src/pro/components/Dataplane/DataplaneConfigForm'
import { useTranslation } from 'react-i18next'

const DataplaneStep = ({
  probeParams,
  onSave,
  readonly,
}: {
  probeParams: any
  onSave: (apoConfig: any) => void
  readonly: boolean
}) => {
  const { t } = useTranslation('core/dataIntegration')
  const [checkedDataplane, setCheckedDataplane] = useState<DataplaneTypeItem | null>(null)

  return (
    <div className="flex flex-col h-full">
      <Typography.Title level={5}>选择接入数据源</Typography.Title>
      <DataplaneRadioList
        readonly={readonly}
        onChange={setCheckedDataplane}
        checkedDataplane={checkedDataplane}
        apoCanCheck={['trace', 'log', 'metric'].some((key) => key in (probeParams || {}))}
      />
      {checkedDataplane && (
        <>
          <Divider></Divider>
          <Typography.Title level={5}>{checkedDataplane.typeName} 配置</Typography.Title>

          <div className="flex-1 overflow-hidden h-0">
            <DataplaneConfigForm
              readonly={readonly}
              dataplaneParams={checkedDataplane}
              onSave={onSave}
              showApoConfig={checkedDataplane.typeName === 'apo'}
            />
          </div>
        </>
      )}
    </div>
  )
}

export default DataplaneStep
