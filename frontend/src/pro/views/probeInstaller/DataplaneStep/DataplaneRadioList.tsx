import { useEffect, useState } from 'react'
import { getDataplaneTypeListApi } from 'src/pro/api/dataplane'
import DataplaneCard from 'src/pro/components/Dataplane/DataplaneCard'
import { DataplaneTypeItem } from 'src/pro/types/datasource'
import { Tooltip } from 'antd'
import { useTranslation } from 'react-i18next'

const DataplaneRadioList = ({
  onChange,
  checkedDataplane,
  apoCanCheck = true,
  readonly = false,
}: {
  onChange: (type: DataplaneTypeItem) => void
  checkedDataplane: DataplaneTypeItem | null
  apoCanCheck: boolean
  readonly: boolean
}) => {
  const { t } = useTranslation('pro/agentDatasource')
  const [dataplaneList, setDataplaneList] = useState<DataplaneTypeItem[]>([])

  useEffect(() => {
    getDataplaneTypeListApi().then((res) => {
      setDataplaneList(res?.dpTypes || [])
    })
  }, [])
  return (
    <div className="w-full p-1">
      <div className="grid gap-2 w-full grid-cols-[repeat(auto-fill,minmax(200px,1fr))]">
        {dataplaneList.map((option) => {
          const isSelected = checkedDataplane?.typeName === option.typeName
          console.log(apoCanCheck, 'apoCanCheck')
          console.log(isSelected, 'isSelected')
          const cardContent = (
            <div
              key={option.typeName}
              onClick={() => {
                if (!apoCanCheck && !readonly) {
                  onChange(option)
                }
              }}
              className={`
                ${isSelected ? 'border-2 border-blue-500 bg-white/20 rounded-xl' : 'border-2 border-transparent'}
                ${apoCanCheck ? 'opacity-50 cursor-not-allowed' : 'cursor-pointer'}
                transition-all duration-300
              `}
            >
              <DataplaneCard info={option} />
            </div>
          )

          return apoCanCheck ? (
            <Tooltip key={option.typeName} title={t('tooltips.apoDataSourceRequired')}>
              {cardContent}
            </Tooltip>
          ) : (
            cardContent
          )
        })}
      </div>
    </div>
  )
}

export default DataplaneRadioList
