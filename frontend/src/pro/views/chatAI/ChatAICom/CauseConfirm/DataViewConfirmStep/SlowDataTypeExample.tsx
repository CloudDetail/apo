import PolarisPng from 'src/pro/assets/9-1.png'
import { Image } from 'antd'
const SlowDataTypeExample = () => {
  return (
    <div className="text-center overflow-hidden">
      <div>慢故障推荐数据将展示节点的北极星数据，以下是示例数据：</div>
      <Image src={PolarisPng} width={200} />
    </div>
  )
}
export default SlowDataTypeExample
