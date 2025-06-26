import { marked } from 'marked'
import '../github-markdown.css'

interface AnalysisProcessProps {
  analyzeProcess?: string
}

const AnalysisProcess = ({ analyzeProcess }: AnalysisProcessProps) => {
  const renderMarkdown = () => {
    if (!analyzeProcess) {
      return { __html: '' }
    }
    return {
      __html: marked(analyzeProcess),
    }
  }

  if (!analyzeProcess) {
    return null
  }

  return (
    <>
      <div className="p-3">
        <div
          dangerouslySetInnerHTML={renderMarkdown()}
          className={`markdown-body`}
          style={{ fontSize: 14, background: 'bg-[var(--ant-color-bg-layout)]', minHeight: 0 }}
        />
      </div>
    </>
  )
}

export default AnalysisProcess
