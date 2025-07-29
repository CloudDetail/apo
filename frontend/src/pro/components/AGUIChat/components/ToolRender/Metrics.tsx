import MultiLineChart from 'src/core/components/Chart/MultiLineChart'

const convertMetricsData = (data) => {
  return data.map((item) => ({
    data: Object.entries(item.chart.chartData)
      ?.map(([key, value]) => [Number(key) / 1000, value])
      .sort((a, b) => a[0] - b[0]),
    legend: item.legend,
  }))
}
const Metrics = ({ args }: { args: any }) => {
  return (
    <div className="grid grid-cols-2 gap-2">
      {args?.map((item) => (
        <div
          className="h-[300px] flex flex-col bg-[var(--ant-color-bg-elevated)] p-2 rounded-xl "
          key={item.title}
        >
          <div className="pb-2 px-2 font-bold">{item.title}</div>
          {/* <LineChart data={data} unit={unit} /> */}
          <MultiLineChart
            chartData={convertMetricsData(item.timeseries)}
            unit={item.unit}
            allowTimeBrush={false}
            YFormatter={(value) => value.toLocaleString() + ' ' + item.unit}
          />
        </div>
      ))}
    </div>
  )
}

export default Metrics
