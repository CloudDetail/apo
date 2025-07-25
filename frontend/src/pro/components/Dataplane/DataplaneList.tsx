import { useEffect, useState } from 'react'
import { getDataplaneTypeListApi } from 'src/pro/api/dataplane'
import DataplaneCard from './DataplaneCard'

interface DataplaneItem {
  typeName: string
  paramSpec: any
  desc: string
  capabilitySpec: string[]
}

const DataplaneList = ({ onClick }: { onClick: (type: string) => void }) => {
  const [dataplaneList, setDataplaneList] = useState<DataplaneItem[]>([])

  useEffect(() => {
    getDataplaneTypeListApi().then((res) => {
      setDataplaneList(res.dpTypes)
    })
  }, [])

  return (
    <div className="w-full p-1">
      <div className="grid gap-2 w-full grid-cols-[repeat(auto-fill,minmax(200px,1fr))]">
        {dataplaneList.map((option) => (
          <div key={option.typeName} onClick={() => onClick(option)}>
            <DataplaneCard info={option} />
          </div>
        ))}
      </div>
    </div>
  )
}

export default DataplaneList
