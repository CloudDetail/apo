import _ from 'lodash'
import { formatUnit } from 'src/core/utils/time'

export const flameDataHandle = (data: string) => {
  const stackList: any[] = []
  const list = _.flatten(data.split('|'))

  if (data.length < 2) {
    return []
  }
  const functionList = list[0].split('#')
  let stackInfo = list[1].split('#')
  stackInfo.forEach((opt) => {
    let temp = opt.split('-')
    let stackItem: any = {
      depth: parseInt(temp[0], 10),
      index: parseInt(temp[1], 10),
      width: parseInt(temp[2], 10),
      color: parseInt(temp[3], 10),
      name: functionList[parseInt(temp[4], 10)],
    }
    stackList.push(stackItem)
  })
  return stackList
}
const truncateText = (text: string, maxLength: number, font: string, ellipsis: string) => {
  if (text.length > maxLength) {
    return text.substring(0, maxLength - ellipsis.length) + ellipsis
  }
  return text
}
export const basiclineOption: any = {
  tooltip: {
    trigger: 'axis',
    order: 'seriesAsc',
    confine: true,
    extraCssText: 'white-space: normal;word-break: break-all;',
    formatter: (params: any) => {
      let text = `<div class="echarts-tooltip">
            <div class="tooltip-name">${params[0].name}</div>
            <div class="tooltip-details">`
      const prepareSeriesName = (seriesName: string) => {
        try {
          const item = JSON.parse(seriesName)
          let text = ''
          Object.keys(item).forEach((key) => {
            text += `<div class="series-item"><span class="series-key">${key}</span><span class="series-value">:${item[key]}</span></div>`
          })
          return text
        } catch (error) {
          return seriesName
        }
      }

      params.forEach((item: any) => {
        text += `<div class="tooltip-item">
                <span class="dot" style="background-color:${item.color}"></span>
                <span class="series-name">${prepareSeriesName(item.seriesName)}</span>
                <span class="value">${isNaN(item.value) ? 'Null' : item.value}</span>
            </div>`
      })

      text += `</div></div>`
      return text
    },
  },
  legend: {
    show: true,
    icon: 'roundRect',
    type: 'scroll',
    left: 10,
    top: 0,
    itemWidth: 10,
    itemHeight: 10,
    data: [],
    itemGap: 30,
    textStyle: {
      width: 800,
      overflow: 'breakAll',
    },
    formatter: function (name: string) {
      return `${truncateText(name, 300, '14px Microsoft Yahei', '…')}`
    },
    tooltip: {
      show: true,
    },
  },
  grid: {
    left: 20,
    right: 40,
    top: 70,
    bottom: 40,
    containLabel: true,
  },
  xAxis: [
    {
      type: 'category',
      boundaryGap: false,
      data: [],
      // axisLabel: {
      //     formatter: (v: any) => dayjs(parseInt(v, 10)).format('MM-DD HH:mm')
      // }
    },
  ],
  yAxis: [
    {
      type: 'value',
      axisLabel: {},
    },
  ],
  series: [],
}

export const buildLineSeriesData = (
  name: string,
  data: any,
  options: { areaStyle?: object; type?: string; connectNulls?: boolean } = {},
) => {
  const { areaStyle = { opacity: 0.12 }, type = 'line', connectNulls = true } = options
  return {
    name: name,
    type: type,
    data: data,
    areaStyle: areaStyle,
    smooth: false,
    connectNulls: connectNulls,
    showSymbol: false,
  }
}

export const stackDataHandle = (data: any[]) => {
  const result: any[] = []
  _.forEach(data, (lockEvent) => {
    let infoList = lockEvent.dataValue.split('!')
    let javalock = {
      waitTime: formatUnit(infoList[0].split('@')[1] / 1000000, 'date'),
      wakeTime: formatUnit(infoList[1] / 1000000, 'date'),
      lockObjectAddress: infoList[3],
      lockType: infoList[4],
      waitThread: formatUnit(infoList[6] / 1000000, 'ms'),
      threadTid: infoList[7],
      stack: infoList[8],
    }
    result.push(javalock)
  })
  return result
}
