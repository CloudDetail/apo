import { Card } from 'antd'
// import DelayLineChart from 'src/core/components/Chart/DelayLineChart'

const Metrics = ({ data, startTime, endTime }) => {
  return (
    <>
      {data?.map((item) => (
        <Card key={item.type} title={item.name} type="inner">
          <div className="h-52 w-80">
            {/* <DelayLineChart
              data={item.chartData}
              timeRange={{ startTime, endTime }}
              type={item.type}
            /> */}
          </div>
        </Card>
      ))}
    </>
  )
}
export default Metrics
