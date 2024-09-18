import { Button } from 'antd'
import React from 'react'
import { LuCopy } from 'react-icons/lu'
import { useMessageContext } from 'src/contexts/MessageContext'
function CopyButton(props) {
  const { value, iconText } = props
  const messageApi = useMessageContext()
  const handleCopy = async () => {
    try {
      if (navigator.clipboard && window.isSecureContext) {
        await navigator.clipboard.writeText(value)
        messageApi.success('内容已复制到剪贴板')
      } else {
        const textArea = document.createElement('textarea')
        textArea.value = value

        textArea.style.position = 'absolute'
        textArea.style.left = '-999999px'

        document.body.prepend(textArea)
        textArea.select()

        try {
          document.execCommand('copy')
          messageApi.success('内容已复制到剪贴板')
        } catch (error) {
          messageApi.error('复制失败')
          console.error(error)
        } finally {
          textArea.remove()
        }
      }
    } catch (err) {
      console.log(err)
      messageApi.error('复制失败')
    }
  }

  return (
    // <Button type="text" icon={<LuCopy />}>
    <div className="cursor-pointer text-blue-500">
      <LuCopy onClick={handleCopy} />
    </div>
    // </Button>
  )
}

export default CopyButton
