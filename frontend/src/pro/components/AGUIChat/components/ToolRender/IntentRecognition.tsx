import { Button, Input } from 'antd'
import { useEffect, useRef, useState } from 'react'

const IntentRecognition = ({ args }: { args: any }) => {
  const [countdown, setCountdown] = useState<number>(5)
  const [executed, setExecuted] = useState(false)
  const timerRef = useRef<NodeJS.Timeout | null>(null)

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
      <Input.TextArea value={args} autoSize readOnly={executed} />
      {!executed && (
        <div className="text-right mt-2">
          <Button type="primary">{countdown}s 后自动执行</Button>
        </div>
      )}
    </>
  )
}

export default IntentRecognition
