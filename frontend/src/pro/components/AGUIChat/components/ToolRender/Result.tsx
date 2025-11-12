import { Tabs } from 'antd'
import { useSelector } from 'react-redux'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'
import EvidencePanel from '../EvidencePanel'

const Result = ({ args }: { args: any }) => {
  const { fields } = args
  const { theme } = useSelector((state: any) => state.settingReducer)

  return (
    <>
      <Tabs
        type="card"
        items={fields?.map((item) => ({
          label: item.label,
          key: item.key,
          children:
            item.key === 'evidence_json' ? (
              <EvidencePanel data={item.children} />
            ) : (
              <div className="max-h-[500px] overflow-y-auto px-2">
                <MarkdownRender content={item.children} theme={theme} />
              </div>
            ),
        }))}
      />
    </>
  )
}

export default Result
