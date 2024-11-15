import { ChartAIProvider } from 'src/pro/contexts/ChartAIContext'
import ChartAICom from './ChartAICom'

function ChartAIPage() {
  return (
    <ChartAIProvider>
      <div className="w-full overflow-hidden flex" style={{ height: 'calc(100vh - 120px)' }}>
        <div className="w-full h-full overflow-hidden">
          <ChartAICom />
        </div>
      </div>
    </ChartAIProvider>
  )
}
export default ChartAIPage
