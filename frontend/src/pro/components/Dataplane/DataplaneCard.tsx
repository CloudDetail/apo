import { dataplaneIconMap } from 'src/pro/constants'
import { DataplaneTypeItem } from 'src/pro/types/datasource'
import Capability from './Capability'

const DataplaneCard = ({ info }: { info: DataplaneTypeItem }) => {
  return (
    <div
      key={info.typeName}
      className={` h-full relative backdrop-blur-md rounded-xl p-2 transition-all duration-300 cursor-pointer  hover:shadow-lg
                     'border border-white/20 bg-white/10 hover:border-white/30 hover:bg-white/15'
              `}
    >
      <div className="flex items-center mb-2 gap-2">
        <img src={dataplaneIconMap[info.typeName]} className="w-8 h-8 object-contain rounded-md" />
        <h3 className="text-lg font-semibold text-gray-800 dark:text-gray-200 m-0">
          {info.typeName}
        </h3>
      </div>
      <div className="flex flex-col gap-3">
        <Capability capabilitySpec={info.capabilitySpec} />
      </div>
    </div>
  )
}

export default DataplaneCard
