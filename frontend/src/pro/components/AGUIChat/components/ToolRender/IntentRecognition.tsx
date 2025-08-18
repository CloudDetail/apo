import { Input } from 'antd'
import { useEffect, useRef, useState } from 'react'
import { useTranslation } from 'react-i18next'

const IntentRecognition = ({ args }: { args: any }) => {
  const { t } = useTranslation('pro/agent')
  const [countdown, setCountdown] = useState<number>(5)
  const [executed, setExecuted] = useState(false)
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    // 初始化后自动开始倒计时
    if (countdown === 0) {
      setExecuted(true)
      return
    }

    timerRef.current = setTimeout(() => {
      setCountdown((prev) => prev - 1)
    }, 1000)

    return () => {
      if (timerRef.current) clearTimeout(timerRef.current)
    }
  }, [countdown, args])

  return (
    <>
      <Input.TextArea value={args} autoSize readOnly={true} />
      {/* {!executed && (
        <div className="text-right mt-2">
          <Button type="primary">
            {countdown}s {t('autoExecute')}
          </Button>
        </div>
      )} */}
    </>
  )
}

export default IntentRecognition
