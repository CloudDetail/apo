import { marked } from 'marked'
import { RiRobot2Line } from 'react-icons/ri'
import 'github-markdown-css/github-markdown-dark.css'
interface HumenMessageProps {
  message: string
}

const HumenMessage: React.FC<HumenMessageProps> = ({ message }) => {
  const renderMarkdown = () => {
    return { __html: marked(message) }
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
      <div className="p-3 " style={{ maxWidth: '80%' }}>
        <div
          dangerouslySetInnerHTML={renderMarkdown()}
          className="p-3  rounded-2xl markdown-body bg-[#2f2f2f]"
          style={{ background: '#2f2f2f', fontSize: 14 }}
        />
      </div>
    </div>
  )
}

export default HumenMessage
