import { Result } from 'antd'

const NoTraceHint = () => {
  return (
    <Result
      title="未对接链路追踪数据，无法分析影响范围"
      // extra={
      //   <Button type="primary" key="console">
      //     Go Console
      //   </Button>
      // }
    />
  )
}
export default NoTraceHint
