import { GoDotFill } from 'react-icons/go'
import { FaChevronDown } from 'react-icons/fa'

const ToolCollapseIcon = ({ hovered }: { hovered: boolean }) => {
  return (
    <div className="flex flex-col items-center justify-center h-full">
      {hovered ? (
        <FaChevronDown
          style={{
            color: 'var(--ant-color-text-tertiary)',
          }}
        />
      ) : (
        <GoDotFill
          style={{
            zIndex: 1000,
            color: 'var(--ant-color-text-tertiary)',
          }}
        />
      )}
    </div>
  )
}
export default ToolCollapseIcon
