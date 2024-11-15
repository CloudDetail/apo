import { marked } from 'marked'
import { RiRobot2Line } from 'react-icons/ri'
import 'github-markdown-css'

interface AIMessageProps {
  message: string
}

const AIMessage: React.FC<AIMessageProps> = ({ message }) => {
  const renderMarkdown = () => {
    return { __html: marked(message) }
  }

  return (
    <div
      className="w-4/5 overflow-hidden flex mb-4 flex-shrink-0"
      style={{
        // whiteSpace: 'pre-line',
        wordBreak: 'break-all',
        overflow: 'hidden',
      }}
    >
      <div className="m-2 flex-shrink-0 flex-grow-0">
        <RiRobot2Line size={24} />
      </div>
      <div className="overflow-hidden">
        <div
          dangerouslySetInnerHTML={renderMarkdown()}
          className="p-3 markdown-body rounded-2xl "
          style={{ fontSize: 14, background: '#3e3f4c' }}
        />
      </div>
    </div>
  )
}

export default AIMessage
