import { Button, Image } from 'antd'
import Empty from 'src/core/components/Empty/Empty'

const SlowType = ({ image, reportUrlList, content }) => {
  return (
    <div className="h-full flex ">
      <div>
        {image?.inferMetricsPng ? (
          <Image src={'data:image/png;base64,' + image.inferMetricsPng} alt={image.inferCause} />
        ) : (
          <Empty context="未查询到北极星指标" />
        )}
      </div>
      {reportUrlList?.length > 0 && (
        <div className="w-[200px] flex-shrink-0 p-2">
          {reportUrlList?.map((reportUrl, index) => (
            <div>
              <Button type="link" target="_blank" href={reportUrl}>
                延时报告_{index + 1}
              </Button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
export default SlowType
