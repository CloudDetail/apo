import { Card } from 'antd'
import { useTranslation } from 'react-i18next'
import DelayLineChart from 'src/core/components/Chart/DelayLineChart'
// import DelayLineChart from 'src/core/components/Chart/DelayLineChart'

const REDMetrics = ({ data, startTime, endTime }) => {
  const { t } = useTranslation('common')
  return (
    <div className="flex">
      {data?.map((item) => (
        <Card
          key={item.type}
          title={t(`delayLineChartTitleMap.${item.type}`)}
          type="inner"
          className=" w-0 flex-1 mx-1"
        >
          <div className="h-52">
            <DelayLineChart
              data={item.chartData}
              timeRange={{ startTime, endTime }}
              type={item.type}
              needFillZero={false}
              allowTimeBrush={false}
            />
          </div>
        </Card>
      ))}
    </div>
  )
}
export default REDMetrics
