import { Alert, Popover, Tooltip } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { IoMdInformationCircleOutline } from 'react-icons/io'

const HintCom = () => {
  const [visible, setVisible] = useState(false)
  const shouldShowPopup = () => {
    const hasShown = localStorage.getItem('AIHintComShown')
    if (hasShown) {
      const parsedData = JSON.parse(hasShown)

      // 检查当前时间是否超过过期时间
      if (new Date().getTime() > parsedData.expires) {
        localStorage.removeItem('AIHintComShown')
        return true // 弹窗过期，应该重新显示
      }

      return false // 弹窗已经显示过且未过期，不需要再显示
    }

    return true // 未找到标记，应该显示弹窗
  }
  const setPopoverShown = () => {
    const expirationDate = new Date()
    expirationDate.setMonth(expirationDate.getMonth() + 1) // 设置过期时间为一个月后

    const popupData = {
      shown: true,
      expires: expirationDate.getTime(),
    }

    localStorage.setItem('AIHintComShown', JSON.stringify(popupData))
    handleManualTrigger(5000)
  }
  const handleVisibleChange = (newVisible: boolean) => {
    // 确保手动和触发行为共存
    setVisible(newVisible)
  }

  const handleManualTrigger = (time = 2000) => {
    // 手动控制触发
    setVisible(true)
    setTimeout(() => setVisible(false), time) // 2秒后自动关闭
  }
  useEffect(() => {
    const visible = shouldShowPopup()
    if (visible) setPopoverShown()
  })
  return (
    <Popover
      trigger="hover" // 保留 hover 和 click 行为
      open={visible} // 控制显示
      onOpenChange={handleVisibleChange} // 触发事件
      placement="right"
      overlayClassName="w-[600px] bg-none "
      content={
        <Alert
          message={
            <div>
              <div>大模型根因分析，主要完成了以下工作。</div>
              <div className="pl-5">
                <div>1、从故障业务提取拓扑图，发送给大模型，请大模型记住该业务的拓扑图。</div>
                <div>
                  2、将故障业务所有关联的告警信息，发送给大模型，请大模型记住故障业务的所有告警。
                </div>
                <div>
                  3、给大模型提供一系列的规则，请大模型按照规则，模仿排障专家，依据告警将疑似根因节点范围不断缩小
                </div>
                <div>4、最终为确认的疑似根因节点，提供北极星指标，请大模型帮助用户选择根因方向</div>
                <div>5、最终提供按照根因方向，提供故障报告佐证疑似根因节点的准确性</div>
              </div>
            </div>
          }
          type="info"
        />
      }
    >
      <IoMdInformationCircleOutline size={20} color="#f7c01a" className="mx-2" />
    </Popover>
  )
}
export default HintCom
