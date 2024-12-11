import { ChatAIProvider } from 'src/pro/contexts/ChatAIContext'
import ChatAICom from './ChatAICom'

function ChatAIPage() {
  return (
    <ChatAIProvider>
      <div className="w-full overflow-hidden flex" style={{ height: 'calc(100vh - 120px)' }}>
        <div className="w-full h-full overflow-hidden">
          <ChatAICom />
        </div>
      </div>
    </ChatAIProvider>
  )
}
export default ChatAIPage
