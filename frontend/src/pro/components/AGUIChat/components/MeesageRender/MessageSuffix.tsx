import { Button } from 'antd'
import { FaRegCopy } from 'react-icons/fa'
import { FiThumbsDown, FiThumbsUp } from 'react-icons/fi'
import { PiShareFatBold } from 'react-icons/pi'

const FinalMessageSuffix = () => {
  return (
    <div className="flex items-center justify-start ml-5 mt-3 text-[var(--ant-color-text-secondary)]">
      <Button type="text" icon={<FaRegCopy />} className="text-[var(--ant-color-text-secondary)]" />
      <Button
        type="text"
        icon={<PiShareFatBold />}
        className="text-[var(--ant-color-text-secondary)]"
      />
      <Button
        type="text"
        icon={<FiThumbsUp />}
        className="text-[var(--ant-color-text-secondary)]"
      />
      <Button
        type="text"
        icon={<FiThumbsDown />}
        className="text-[var(--ant-color-text-secondary)]"
      />
    </div>
  )
}

export default FinalMessageSuffix
