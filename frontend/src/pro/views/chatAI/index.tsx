import { ChatAIProvider } from 'src/pro/contexts/ChatAIContext'
import ChatAICom from './ChatAICom'

function ChatAIPage() {
  return (
    <ChatAIProvider>
      <ChatAICom />
    </ChatAIProvider>
  )
}
export default ChatAIPage
