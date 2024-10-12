import { Button, message } from 'antd'
import React from 'react'
import { LuCopy } from 'react-icons/lu'

export const copyValue = async (value) => {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(value)
      message.success('内容已复制到剪贴板')
    } else {
      const textArea = document.createElement('textarea')
      textArea.value = value

      textArea.style.position = 'absolute'
      textArea.style.left = '-999999px'

      document.body.prepend(textArea)
      textArea.select()

      try {
        document.execCommand('copy')
        message.success('内容已复制到剪贴板')
      } catch (error) {
        message.error('复制失败')
        console.error(error)
      } finally {
        textArea.remove()
      }
    }
  } catch (err) {
    console.log(err)
    message.error('复制失败')
  }
}
function CopyButton(props) {
  const { value, iconText } = props

  return (
    // <Button type="text" icon={<LuCopy />}>
    <div className="cursor-pointer text-blue-500">
      <LuCopy onClick={copyValue} />
    </div>
    // </Button>
  )
}

export default CopyButton
