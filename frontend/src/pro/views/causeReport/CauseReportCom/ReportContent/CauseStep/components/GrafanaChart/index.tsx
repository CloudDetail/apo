import React, { forwardRef, memo, useImperativeHandle, useRef } from 'react'
import ReactEcharts from 'echarts-for-react'
import * as echarts from 'echarts'
import { isEqual } from 'lodash'
// @ts-ignore
import grafana from './grafana'
import grafanaLight from './grafanaLight'
import { Empty } from 'antd'
import { useSelector } from 'react-redux'

echarts.registerTheme('myDark', grafana)
echarts.registerTheme('myLight', grafanaLight)
// echarts.registerTheme('myMacarons', myMacarons);

interface Props {
  option: any
  height?: string | number
  events?: any
  className?: string
}
function GrafanaChart(props: Props, ref: any) {
  const { option, height = '100%', events = null, className = '' } = props
  const echartRef: any = useRef(null)

  // 将echarts实例抛出
  useImperativeHandle(ref, () => {
    return {
      echartInstance: echartRef?.current?.getEchartsInstance() || null,
    }
  })
  const { theme } = useSelector((state) => state.settingReducer)

  return option.series?.length > 0 && option.series[0]?.data?.length > 0 ? (
    <ReactEcharts
      theme={theme === 'dark' ? 'myDark' : 'myLight'}
      ref={echartRef}
      className={className}
      option={option}
      style={{ height, width: '100%' }}
      onEvents={events}
      notMerge
    ></ReactEcharts>
  ) : (
    <Empty />
  )
}

const isOptEqual = (prevProps: Props, nextProps: Props) => {
  if (isEqual(prevProps.option, nextProps.option)) {
    return true
  } else {
    return false
  }
}
export default memo(forwardRef(GrafanaChart), isOptEqual)
