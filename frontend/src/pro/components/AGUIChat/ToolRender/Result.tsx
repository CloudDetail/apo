import { Tabs } from 'antd'
import { useSelector } from 'react-redux'
import MarkdownRender from 'src/core/components/Markdown/MarkdownRender'

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
          children: <MarkdownRender content={item.children} theme={theme} />,
        }))}
      />
    </>
  )
}

export default Result
