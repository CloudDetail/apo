// DataplaneCard.tsx
import { dataplaneIconMap } from 'src/pro/constants'
import { DataplaneTypeItem } from 'src/pro/types/datasource'
import Capability from './Capability'

const DataplaneCard = ({ info }: { info: DataplaneTypeItem }) => {
  return (
    <div
      className="
        group relative rounded-xl border border-[var(--ant-color-border-secondary)]  bg-[var(--ant-color-fill-quaternary)]
        p-2 cursor-pointer hover:shadow-xl hover:border-white/30
        transition-transform duration-200 ease-[cubic-bezier(0.22,1,0.36,1)]
        hover:-translate-y-0.5
        backdrop-blur-md
        isolation-auto
        hover:z-10
      "
    >
      <div className="transform-gpu will-change-transform transition-transform duration-200 ease-[cubic-bezier(0.22,1,0.36,1)] group-hover:scale-[1.02]">
        <div className="flex items-center mb-2 gap-2">
          <img
            src={dataplaneIconMap[info.typeName]}
            className="w-8 h-8 object-contain rounded-md"
          />
          <h3 className="text-lg font-semibold text-[var(--ant-color-text-secondary)] m-0">
            {info.typeDisplayName}
          </h3>
        </div>

        <div className="flex flex-col gap-3">
          <Capability capability={info.capabilitySpec} />
        </div>
      </div>
    </div>
  )
}

export default DataplaneCard
