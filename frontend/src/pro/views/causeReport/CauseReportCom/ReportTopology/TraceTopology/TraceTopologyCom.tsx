import React, { useEffect, useRef } from 'react'
import * as d3 from 'd3'
import _ from 'lodash'
import { ThemeColor } from 'src/constants'
import { formatUnit } from 'src/core/utils/time'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'
import { useTranslation } from 'react-i18next'
import { theme } from 'antd'
const i18nKey = 'topology'
const redRGB = {
  r: 255,
  g: 60,
  b: 60,
}
const greenRGB = {
  r: 95,
  g: 224,
  b: 97,
}
// 构造对应深浅的色系组
function rgba(rgb: any, alpha: number) {
  return `rgba(${rgb.r}, ${rgb.g}, ${rgb.b}, ${alpha})`
}
function rgbaToHex(rgba: any) {
  let [r, g, b, a] = rgba.match(/\d+(\.\d+)?/g)
  let format = (n: any) => {
    n = parseInt(n, 10)
    return n.toString(16).padStart(2, '0')
  }
  return `#${format(r)}${format(g)}${format(b)}${Math.round(a * 255).toString(16)}`
}
// 平铺树状节点数据，构造唯一ID。节点id存在重复调用的情况
function getAllNodes(
  list: any[],
  data: any[],
  parentId: any,
  timeInfo: { earliestStartTime: number; latestEndTime: number },
) {
  _.map(data, (node, idx) => {
    // 检查并更新最早的 startTime 和最晚的 endTime
    if (timeInfo.earliestStartTime === 0 || node.startTime < timeInfo.earliestStartTime) {
      timeInfo.earliestStartTime = node.startTime
    }
    if (timeInfo.latestEndTime === 0 || node.totalTime + node.startTime > timeInfo.latestEndTime) {
      timeInfo.latestEndTime = node.totalTime + node.startTime
    }

    if (!node.hasOwnProperty('key')) {
      Object.defineProperty(node, 'key', {
        value: `${parentId}_${idx}`,
        writable: true,
        enumerable: true,
        configurable: true,
      })
    }

    let newNode = { ...node }
    delete newNode.children
    list.push(newNode)

    if (node.children && node.children.length > 0) {
      getAllNodes(list, node.children, `${parentId}_${idx}`, timeInfo)
    }
  })
}

function TraceTopologyCom({ data, category }: any) {
  const { t } = useTranslation('pro/causeReport')
  const nodes = useRef<any>()
  const maxandminMutatedValue = useRef<any>()
  const chartRef = useRef<any>()
  const svgRef = useRef<any>()
  const timeScale = useRef<any>()
  const colorScale = useRef<any>()
  const { node, setNode: setSelectedErrorTreeNode } = useCauseReportContext()
  let selectedErrorTreeNode = useRef<any>()
  // const [parentWidth, setParentWidth] = useState(0);
  const { useToken } = theme
  const { token } = useToken()
  const legendItems = [
    {
      label: t(`${i18nKey}.noAnomaly`),
      color: ThemeColor.gray,
      yAxis: 0,
    },
    {
      label: t(`${i18nKey}.anomaly`),
      color: ThemeColor.red,
      yAxis: 0,
    },
  ]

  const draw = (nodeList: any[], chartWarp: any, level: number) => {
    let nodeLevelG: any
    if (d3.select(`#node_level_${level}`).node()) {
      nodeLevelG = d3.select(`#node_level_${level}`)
    } else {
      nodeLevelG = chartWarp
        .append('g')
        .attr('id', `node_level_${level}`)
        .attr('transform', `translate(0, ${level * 20})`)
    }

    if (_.some(nodeList, (node) => node.children && node.children.length > 0)) {
      level++
    }
    nodeList.forEach((node) => {
      const left = timeScale.current(node.startTime) as number
      const right = timeScale.current(node.startTime + node.totalTime) as number
      const rectWidth = right - left
      const visibleRectWidth = Math.max(rectWidth, 3)
      const rect = nodeLevelG
        .append('rect')
        .style('cursor', 'pointer')
        .attr('width', visibleRectWidth)
        .attr('data-key', node.key)
        .attr('height', 20)
        .attr('stroke', token.colorBgContainer)
        .attr('stroke-width', rectWidth < 3 ? 1 : 0)
        .attr('fill', () => {
          if (node.mutatedValue > 0) {
            let precent = node.mutatedValue / maxandminMutatedValue.current[1]
            return colorScale.current(precent * 100)
          } else {
            let precent = -node.mutatedValue / maxandminMutatedValue.current[0]
            return colorScale.current(isNaN(precent) ? 0 : precent * 100)
          }
        })
        .attr('x', left)
        .attr('y', 5)

      nodeLevelG
        .append('text')
        .text(rectWidth >= 20 ? node.id.substring(0, Math.floor(rectWidth / 10)) : '')
        .attr('class', 'node-text')
        .attr('style', 'font-size: 12px; fill:  var(--ant-color-text-base); pointer-events: none;')
        .attr('x', left + 5)
        .attr('y', 20)

      rect.on('click', function (this: any, e: any) {
        e.stopPropagation()
        const { pageX, clientY } = e
        const bodyWidth = document.body.clientWidth
        const tooltip = d3.select('#chart_tooltip')
        const tooltipWidth = tooltip.node()?.getBoundingClientRect().width || 0
        // 判断 tooltip 在页面中左右的位置
        let tooltipX = pageX
        if (pageX + tooltipWidth > bodyWidth) {
          // 如果 tooltip 超出右边界，则左对齐
          tooltipX = pageX - tooltipWidth
        }

        // 设置 tooltip 的准确位置
        tooltip.attr(
          'style',
          `z-index: 99; visibility: visible; position: fixed; left: ${tooltipX}px; top: ${clientY}px`,
        )
        const key = d3.select(this).attr('data-key')
        let node = nodes.current.find((opt: any) => opt.key === key)
        const htmlContent = `<div class='tooltip_content'>
  <div class='title'>${node.id}</div>
  <div class='content'>
    <div class='content-info'>URL: ${node.url}</div>
    <div class='content-info-row'>
      <div class='content-info'>${t(`${i18nKey}.profiled`)}: ${node.isProfiled ? t(`${i18nKey}.profiled`) : t(`${i18nKey}.notProfiled`)}</div>
      <div class='content-info'>${t(`${i18nKey}.selfTime`)}: ${formatUnit(node.selfTime, 'ns')}</div>
    </div>
    <div class='content-info-row'>
      <div class='content-info'>${t(`${i18nKey}.historicalBaseline`)}: ${formatUnit(node.p90, 'ns')}</div>
      <div class='content-info'>${t(`${i18nKey}.currentDuration`)}: ${formatUnit(node.totalTime, 'ns')}</div>
    </div>
    <div class='content-info-row'>
      <div class='content-info'>${t(`${i18nKey}.mutationRatio`)}: ${(node.mutatedValuePrecent * 100).toFixed(2) + '%'}</div>
      <div class='content-info'>${t(`${i18nKey}.mutationRank`)}: ${node.mutatedValuePrecentSort}</div>
    </div>
  </div>
</div>`
        d3.select('#chart_tooltip').html(htmlContent)
      })
      // rect.on('mouseenter', () => {
      //     console.log('mouseenter')
      // })
      // rect.on('mouseleave', () => {
      //     console.log('mouseleave')
      // })

      if (node.children && node.children.length > 0) {
        draw(node.children, chartWarp, level)
      } else {
        const height = chartRef.current.clientHeight
        let contentHeight = (level + 1) * 20 + 70
        if (contentHeight > height) {
          svgRef.current.style('height', contentHeight)
        }
      }
    })
  }

  const drawError = (nodeList: any[], chartWarp: any, level: number) => {
    let nodeLevelG: any
    if (d3.select(`#node_level_${level}`).node()) {
      nodeLevelG = d3.select(`#node_level_${level}`)
    } else {
      nodeLevelG = chartWarp
        .append('g')
        .attr('id', `node_level_${level}`)
        .attr('transform', `translate(0, ${level * 20})`)
    }

    if (_.some(nodeList, (node) => node.children && node.children.length > 0)) {
      level++
    }

    nodeList.forEach((node) => {
      const left = timeScale.current(node.startTime) as number
      const right = timeScale.current(node.startTime + node.totalTime) as number
      const rectWidth = right - left
      const rect = nodeLevelG
        .append('rect')
        .style('cursor', 'pointer')
        .attr('width', rectWidth)
        .attr('data-key', node.key)
        .attr('class', 'node-row')
        .attr('height', 20)
        .attr('fill', () => {
          let color = ThemeColor.gray
          if (node.isError) {
            color = ThemeColor.red
          }
          return color
        })
        .attr('x', left)
        .attr('y', 5)

      nodeLevelG
        .append('text')
        .text(node.id.substring(0, Math.floor(rectWidth / 10)))
        .attr('class', 'node-text ')
        .attr('style', 'font-size: 12px; fill: var(--ant-color-text-base); pointer-events: none;')
        .attr('x', left + 5)
        .attr('y', 20)

      rect.on('click', function (this: any, e: any) {
        e.stopPropagation()
        const key = d3.select(this).attr('data-key')
        let node = nodes.current.find((opt: any) => opt.key === key)
        setSelectedErrorTreeNode(node)
      })
      rect.on('mouseenter', function (this: any, e: any) {
        e.stopPropagation()
        const { pageX, clientY } = e
        const bodyWidth = document.body.clientWidth
        const tooltip = d3.select('#chart_tooltip')
        const tooltipWidth = tooltip.node()?.getBoundingClientRect().width || 0
        // 判断 tooltip 在页面中左右的位置
        let tooltipX = pageX
        if (pageX + tooltipWidth > bodyWidth) {
          // 如果 tooltip 超出右边界，则左对齐
          tooltipX = pageX - tooltipWidth
        }

        // 设置 tooltip 的准确位置
        tooltip.attr(
          'style',
          `z-index: 99; visibility: visible; position: fixed; left: ${tooltipX}px; top: ${clientY}px`,
        )

        const htmlContent = `<div class='tooltip_content'>
                    <div class='title'>${node.id}</div>
                    <div class='content'>
                        <div class='content-info'>URL: ${node.url}</div>
                        <div class='content-info-row'>
                            <div class='content-info'>是否被Profiled: ${node.isProfiled ? '是' : '否'}</div>
                            <div class='content-info'>本次调用: ${formatUnit(node.totalTime, 'ns')}</div>
                        </div>
                    </div>
                </div>`
        d3.select('#chart_tooltip').html(htmlContent)
      })
      rect.on('mouseleave', (e: any) => {
        d3.select('#chart_tooltip').attr('style', `visibility: hidden;`)
      })
      if (node.isMutated) {
        selectedErrorTreeNode = node
      }
      if (node.children && node.children.length > 0) {
        drawError(node.children, chartWarp, level)
      } else {
        const height = chartRef.current.clientHeight
        let contentHeight = (level + 1) * 20 + 70
        if (contentHeight > height) {
          svgRef.current.style('height', contentHeight)
        }
      }
    })
  }
  const prepareForSlowTrace = (data: any) => {
    const nodeList: any[] = []
    let timeInfo = { earliestStartTime: Infinity, latestEndTime: 0 }
    getAllNodes(nodeList, [data], 0, timeInfo)
    let maxMutatedValue = _.chain(nodeList)
      .map('mutatedValue')
      .filter((v) => v > 0)
      .sum()
      .value()
    let minMutatedValue = _.chain(nodeList)
      .map('mutatedValue')
      .filter((v) => v < 0)
      .sum()
      .value()
    maxandminMutatedValue.current = [minMutatedValue, maxMutatedValue]
    _.forEach(nodeList, (node: any) => {
      if (node.mutatedValue > 0) {
        node.mutatedValuePrecent = node.mutatedValue / maxandminMutatedValue.current[1]
      } else if (node.mutatedValue === 0) {
        node.mutatedValuePrecent = 0
      } else {
        node.mutatedValuePrecent = -node.mutatedValue / maxandminMutatedValue.current[0]
      }
    })
    let mutatedValuePrecentList: number[] = _.chain(nodeList)
      .sortBy('mutatedValuePrecent')
      .map('mutatedValuePrecent')
      .reverse()
      .sortedUniq()
      .value()
    _.forEach(nodeList, (node: any) => {
      node.mutatedValuePrecentSort = mutatedValuePrecentList.indexOf(node.mutatedValuePrecent) + 1
    })
    nodes.current = nodeList

    const container: any = document.getElementById('trace_topo')
    container.innerHTML = ''
    if (_.isEmpty(data)) {
      return
    }

    const values = d3.range(-100, 120, 20)
    const colorRectWidth = 70
    const legendWidth = values.length * colorRectWidth
    const minChartWidth = legendWidth + 220
    const WIDTH = Math.max(chartRef.current.clientWidth, minChartWidth)
    const HEIGHT = chartRef.current.clientHeight
    timeScale.current = d3
      .scaleLinear()
      .domain([timeInfo.earliestStartTime, timeInfo.latestEndTime])
      .range([0, WIDTH])

    const svg = d3
      .select('#trace_topo')
      .append('svg')
      .attr('width', WIDTH)
      .attr('height', HEIGHT - 20)
      .style('width', `${WIDTH}px`)
      .style('height', `${HEIGHT - 20}px`)
    const g = svg.append('g').attr('id', 'chart_warp').attr('transform', `translate(0, 60)`)
    svgRef.current = svg

    // 生成颜色对应比例尺
    let colors: string[] = []
    let alphavalue = -1
    values.forEach((value, idx) => {
      let alpha = alphavalue + idx * 0.2
      if (value < 0) {
        colors.push(rgbaToHex(rgba(greenRGB, Math.abs(alpha))))
      } else if (value === 0) {
        colors.push('#adabab')
      } else {
        colors.push(rgbaToHex(rgba(redRGB, alpha)))
      }
    })
    colorScale.current = d3
      .scaleLinear()
      .domain(values)
      // @ts-ignore
      .range(colors)

    const colorGWarp = svg
      .append('g')
      .attr('transform', `translate(${WIDTH / 2 - legendWidth / 2}, 10)`)
    colorGWarp
      .append('text')
      .text('(-)' + t(`${i18nKey}.decrease`))
      .attr('x', (values.length * colorRectWidth) / 2 - 280)
      .attr('y', 10)
      .attr('style', `font-size: 14px;fill: ${ThemeColor.green}`)
    colorGWarp
      .append('text')
      .text('(+)' + t(`${i18nKey}.increase`))
      .attr('x', (values.length * colorRectWidth) / 2 + 170)
      .attr('y', 10)
      .attr('style', `font-size: 14px;fill: ${ThemeColor.red}`)
    colorGWarp
      .selectAll('rect')
      .data(values)
      .enter()
      .append('rect')
      .attr('x', (d, i) => i * colorRectWidth)
      .attr('y', 20)
      .attr('width', colorRectWidth)
      .attr('height', 20)
      .style('fill', (d) => {
        return colorScale.current(d)
      })

    colorGWarp
      .selectAll('.color_text')
      .data(values)
      .enter()
      .append('text')
      .text((d) => d + '%')
      .attr('class', 'color_text')
      .attr('x', function (d, i) {
        return i * colorRectWidth + colorRectWidth / 2 - this.getComputedTextLength() / 2
      })
      .attr('y', 35)
      .style('fill', 'var(--ant-color-text-base)')

    svg.on('click', () => {
      d3.select('#chart_tooltip').attr('style', `visibility: none;`)
    })
    draw([data], g, 0)
  }
  const prepareForErrorTrace = (data: any) => {
    const nodeList: any[] = []
    let timeInfo = { earliestStartTime: Infinity, latestEndTime: 0 }
    getAllNodes(nodeList, [data], 0, timeInfo)
    nodes.current = nodeList
    const container: any = document.getElementById('trace_topo')
    container.innerHTML = ''
    if (_.isEmpty(data)) {
      return
    }
    const WIDTH = chartRef.current.clientWidth
    const HEIGHT = chartRef.current.clientHeight

    timeScale.current = d3
      .scaleLinear()
      .domain([timeInfo.earliestStartTime, timeInfo.latestEndTime])
      .range([0, WIDTH])
    const svg = d3
      .select('#trace_topo')
      .append('svg')
      .style('width', WIDTH)
      .style('height', HEIGHT - 20)
    svgRef.current = svg
    const g = svg.append('g').attr('id', 'chart_warp').attr('transform', `translate(0, 40)`)

    let legends = svg
      .selectAll('.legend')
      .data(legendItems)
      .enter()
      .append('g')
      .attr('class', 'legend')
      .attr('transform', function (d, i) {
        return 'translate(' + (20 + i * 120) + ',20 )'
      })

    legends
      .append('circle')
      .attr('cx', 0)
      .attr('cy', 0)
      .attr('r', 10)
      .style('fill', function (d) {
        return d.color
      })

    legends
      .append('text')
      .attr('x', 15)
      .attr('y', 5)
      .attr('style', `fill:${token.colorText}`)
      .text(function (d) {
        return d.label
      })

    drawError([data], g, 0)
    setSelectedErrorTreeNode(selectedErrorTreeNode)
  }
  useEffect(() => {
    if (category === 'slow') {
      prepareForSlowTrace(data)
    } else {
      prepareForErrorTrace(data)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data])

  return (
    <div className={category === 'slow' ? 'trace_topo_warp' : ''}>
      <div
        ref={chartRef}
        id="trace_topo"
        style={{ width: '100%', height: '100%', overflowX: 'auto' }}
      ></div>
      <div id="chart_tooltip" className={'custom_tooltip'}></div>
    </div>
  )
}

export default TraceTopologyCom
