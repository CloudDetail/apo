import { Alert, Button, Card, ConfigProvider, Image, Tabs, TabsProps, Tag } from 'antd'
import { useTranslation } from 'react-i18next'
import { SlowErrorType } from 'src/constants'
import Empty from 'src/core/components/Empty/Empty'
interface ImageData {
  inferCause: string
  inferMetricsPng: string
}
interface UrlData {
  type: string
  urls: string[]
}
interface SlowTypeProps {
  image: ImageData
  otherUrlList: UrlData[]
  reportType: string
  reportUrlList: string[]
  content: string
}
const UrlListCom = ({ urls, type }: UrlData) => {
  const { t } = useTranslation('pro/dataViewCom')

  return (
    <>
      <Alert message={SlowErrorType[type]} type="info" showIcon className="text-xs" />
      {urls?.length > 0 ? (
        <div className="w-[200px] flex-shrink-0 p-2">
          {urls?.map((reportUrl, index) => (
            <div>
              <Button type="link" target="_blank" href={reportUrl}>
                {t('report', { type, index: index + 1 })}
              </Button>
            </div>
          ))}
        </div>
      ) : (
        <Empty context={t('noReport', { type })} />
      )}
    </>
  )
}
const SlowType = ({ image, reportUrlList, content, reportType, otherUrlList }: SlowTypeProps) => {
  const { t } = useTranslation('pro/dataViewCom')

  const getTabsItems = () => {
    return Object.entries(SlowErrorType).map(([key, value]) => {
      let urls = []
      if (reportType === key) {
        urls = reportUrlList
      } else {
        urls = otherUrlList.find((item) => item.type === key)?.urls || []
      }
      return {
        label: (
          <>
            {key}
            {reportType === key && (
              <Tag color="success" className="ml-1">
                {t('recommend')}
              </Tag>
            )}
          </>
        ),
        key: key,
        children: <UrlListCom urls={urls} type={key} />,
      }
    })
  }
  return (
    <div className="h-full flex overflow-hidden ">
      <div
        className="w-1/3 flex-grow-0 flex-shrink-0 bg-white"
        style={{ display: 'flex', justifyContent: 'center', alignItems: 'flex-start' }}
      >
        {image?.inferMetricsPng ? (
          <Image
            src={'data:image/png;base64,' + image.inferMetricsPng}
            alt={image.inferCause}
            height={'400px'}
            width={'100%'}
          />
        ) : (
          <Empty context={t('noData')} />
        )}
      </div>
      <div className="flex-1 px-4">
        {content && <div className=" py-3">{content}</div>}
        <Card classNames={{ body: 'p-1' }}>
          <ConfigProvider
            theme={{
              components: {
                Tabs: {
                  verticalItemPadding: '8px',
                },
              },
              token: {
                paddingLG: 4,
              },
            }}
          >
            <Tabs tabPosition="left" items={getTabsItems()} size="small" className="pl-0" />
          </ConfigProvider>
        </Card>
      </div>
    </div>
  )
}
export default SlowType
