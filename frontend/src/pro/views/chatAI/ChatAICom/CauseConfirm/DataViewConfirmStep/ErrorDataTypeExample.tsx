import LogsPng from 'src/pro/assets/dataViewTypeExample/logs.png'
import { Image } from 'antd'
const ErrorDataTypeExample = () => {
  return (
    <div className="text-center">
      <div>错故障推荐数据将展示节点的日志列表，以下是示例数据：</div>
      <Image src={LogsPng} width={400} />
    </div>
  )
}
export default ErrorDataTypeExample
