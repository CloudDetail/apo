import { Button, Image } from 'antd'
import Empty from 'src/core/components/Empty/Empty'

const SlowType = ({ image, reportUrlList, content }) => {
  return (
    <div className="h-full flex overflow-hidden ">
      <div style={{ display: 'flex', justifyContent: 'center', alignItems: 'center' }}>
        {image?.inferMetricsPng ? (
          <Image
            src={'data:image/png;base64,' + image.inferMetricsPng}
            alt={image.inferCause}
            height={'400px'}
            width={'100%'}
          />
        ) : (
          <Empty context="未查询到北极星指标" />
        )}
      </div>
      <div className="flex-1">
        {reportUrlList?.length > 0 ? (
          <div className="w-[200px] flex-shrink-0 p-2">
            {reportUrlList?.map((reportUrl, index) => (
              <div>
                <Button type="link" target="_blank" href={reportUrl}>
                  延时报告_{index + 1}
                </Button>
              </div>
            ))}
          </div>
        ) : (
          <Empty context="该节点无延时报告" />
        )}
      </div>
    </div>
  )
}
export default SlowType
