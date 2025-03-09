import { Alert, Popover, Tooltip } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { IoMdInformationCircleOutline } from 'react-icons/io'

const i18nKey = 'hint'
const HintCom = () => {
  const { t } = useTranslation('pro/chatAI')
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
              <div>{t(`${i18nKey}.aiRootCauseAnalysis`)}</div>
              <div className="pl-5">
                <div>{t(`${i18nKey}.extractTopology`)}</div>
                <div>{t(`${i18nKey}.sendAlertInfo`)}</div>
                <div>{t(`${i18nKey}.applyRules`)}</div>
                <div>{t(`${i18nKey}.provideMetrics`)}</div>
                <div>{t(`${i18nKey}.generateReport`)}</div>
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
