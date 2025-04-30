import { marked } from 'marked'
import 'github-markdown-css/github-markdown-dark.css'
import { Button, Image } from 'antd'
import { LuPaperclip } from 'react-icons/lu'
import { FaRegUser } from 'react-icons/fa'
import { downloadFileApi } from 'src/pro/api/chatAI'
interface HumanMessageProps {
  message: string
  fileList: string[]
  img?: any
}

const HumanMessage: React.FC<HumanMessageProps> = ({ message, fileList, img }) => {
  const renderMarkdown = () => {
    return { __html: marked(message) }
  }
  const downloadFile = (filename: string) => {
    downloadFileApi({ filename }).then((res) => {
      // 自动判断文件类型
      const fileExtension = filename.split('.').pop().toLowerCase()
      let mimeType
      let blob
      if (fileExtension === 'json') {
        mimeType = 'application/json'
        try {
          // 将 JSON 对象转为字符串（格式化输出）
          const jsonString = JSON.stringify(res, null, 2)
          blob = new Blob([jsonString], { type: mimeType })
        } catch (e) {
          // alert('JSON 转换错误，请检查数据！')
          return
        }
      } else if (fileExtension === 'txt') {
        mimeType = 'text/plain'
        blob = new Blob([res], { type: mimeType }) // 假设 res 是纯文本
      } else {
        mimeType = 'application/octet-stream'
        blob = new Blob([res], { type: mimeType })
      }

      // 创建下载链接
      const link = document.createElement('a')
      link.href = URL.createObjectURL(blob)
      link.download = filename
      link.style.display = 'none'

      // 触发下载
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    })
  }

  return (
    <div
      className="w-full overflow-hidden flex mb-4  justify-end"
      style={{
        // whiteSpace: 'pre-line',
        wordBreak: 'break-all',
        overflow: 'hidden',
      }}
    >
      <div className="overflow-hidden" style={{ maxWidth: '80%' }}>
        <div
          dangerouslySetInnerHTML={renderMarkdown()}
          className="p-3  rounded-2xl markdown-body bg-[#2f2f2f] "
          style={{ background: '#2f2f2f', fontSize: 14, minHeight: 0 }}
        />
        <div className="text-right ">
          {img?.inferMetricsPng && (
            <Image src={'data:image/png;base64,' + img.inferMetricsPng} alt={img.inferCause} />
          )}
          {fileList.map((file) => (
            <Button type="text" onClick={() => downloadFile(file)}>
              <LuPaperclip color="#fff" />
              <span className=" underline">{file}</span>
            </Button>
          ))}
        </div>
      </div>

      <div className="m-2 flex-shrink-0 flex-grow-0">
        <FaRegUser size={19} />
      </div>
    </div>
  )
}

export default HumanMessage
