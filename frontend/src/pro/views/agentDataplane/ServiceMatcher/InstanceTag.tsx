import { useSelector } from 'react-redux'
const baseClasses = `mr-1 mb-1 inline-flex items-center gap-1.5 font-medium rounded-full transition-all duration-200  cursor-pointer shadow-sm px-2 py-1 text-xs`

const InstanceTag = ({
  instanceName,
  source,
}: {
  instanceName: string
  source: 'source' | 'rule'
}) => {
  const theme = useSelector((state: any) => state.settingReducer.theme)
  const getTagColor = () => {
    if (source === 'source') {
      if (theme === 'dark') {
        return 'bg-blue-900/50 text-blue-300  hover:bg-blue-800/50'
      } else {
        return 'bg-blue-100 text-blue-800 hover:bg-blue-200'
      }
    } else {
      if (theme === 'dark') {
        return 'bg-sky-900/50 text-sky-300 hover:bg-sky-800/50'
      } else {
        return 'bg-sky-100 text-sky-800 hover:bg-sky-200'
      }
    }
  }
  return <span className={`${baseClasses} ${getTagColor()}`}>{instanceName}</span>
}

export default InstanceTag
