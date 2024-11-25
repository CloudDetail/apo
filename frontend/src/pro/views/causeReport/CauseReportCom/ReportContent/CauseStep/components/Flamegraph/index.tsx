import React, { useEffect, useRef, useState } from 'react'
import * as d3 from 'd3'
import 'd3-flame-graph/dist/d3-flamegraph.css'
import { Empty } from 'antd'
import styles from './flamegraph.module.scss'
import { colorBasedOnPackageName, getPackageNameFromStackTrace } from './color'
import _ from 'lodash'

export const COLLAPSE_THRESHOLD = 5
export default function Flamegraph(props: any) {
  const { data = {} } = props
  const { levels = [], maxSelf = 0, names = [], numTicks = 0 } = data
  const flamegraphRef = useRef<HTMLDivElement>(null)
  const init = useRef(true)
  let chart: any, renderFlag
  const layerHeight = 20

  const drawFlamegraph = (container: any) => {
    const GRAPH_WIDTH = container.offsetWidth - 10
    const GRAPH_HEIGHT = layerHeight * levels?.length + 20
    d3.select('.flamegraph').selectAll('svg').remove()
    const tooltip = d3.select('#tooltip') // 获取 tooltip 元素
    const svg = d3
      .select('.flamegraph')
      .append('svg')
      .attr('width', GRAPH_WIDTH)
      .attr('height', GRAPH_HEIGHT)
    // 定义层级的高度

    const g = svg.append('g')
    // 遍历每一层数据
    const tickWidth = GRAPH_WIDTH / numTicks
    // 设置比例尺，将 value 映射到宽度
    const xScale = d3.scaleLinear().domain([0, numTicks]).range([0, GRAPH_WIDTH])

    //步长
    const step = 4

    levels.forEach((level: any, levelIndex: number) => {
      // 计算每一层的总宽度

      // 当前层的 y 坐标
      const y = levelIndex * layerHeight

      // 绘制每个矩形
      let xOffset = 0 // 起始位置

      for (let j = 0; j < level.length; j += step) {
        //  上一个节点结束位置+ 距离上一个节点结束位置偏移量 = 当前节点开始位置
        xOffset += level[j] * tickWidth
        const rectWidth = xScale(level[j + 1])
        const present = level[j + 1] / numTicks // 计算比例
        const percentage = (present * 100).toFixed(2).replace(/\.?0+$/, '') // 转换为百分比
        const widthPresent = rectWidth / GRAPH_WIDTH
        // 绘制矩形
        const name = names[level[j + 3]]
        const packageName = getPackageNameFromStackTrace('', name) || ''
        const color = colorBasedOnPackageName(packageName).alpha(1)
        // const color = getBarColorByValue(level[j + 1], numTicks)
        const rect = g
          .append('rect')
          // .attr('class', styles.flameRect)
          .attr('x', xOffset)
          .attr('y', y)
          .attr('width', rectWidth)
          .attr('height', layerHeight) // 加上间距
          .attr('fill', rectWidth <= 5 ? '#393e47' : color)

          .on('mouseout', () => {
            const tooltip = d3.select('#chart_tooltip')
            tooltip.style('visibility', 'hidden')
          })
          .on('mousemove', function (this: any, e: any) {
            const tooltip = d3.select('#chart_tooltip')
            const { pageX, clientY } = e
            const bodyWidth = document.body.clientWidth
            const tooltipWidth = tooltip.node()?.getBoundingClientRect().width || 0
            // 判断 tooltip 在页面中左右的位置
            let tooltipX = pageX
            if (pageX + tooltipWidth > bodyWidth) {
              // 如果 tooltip 超出右边界，则左对齐
              tooltipX = pageX - tooltipWidth
            }
            tooltip.attr(
              'style',
              `z-index: 99; visibility: visible; position: fixed; left: ${tooltipX + 10}px; top: ${clientY + 10}px`,
            )
          })
          .on('mouseenter', function (this: any, e: any) {
            const tooltip = d3.select('#chart_tooltip')

            tooltip.style('visibility', 'visible')
            const { pageX, clientY } = e
            const bodyWidth = document.body.clientWidth
            const tooltipWidth = tooltip.node()?.getBoundingClientRect().width || 0
            // 判断 tooltip 在页面中左右的位置
            let tooltipX = pageX
            if (pageX + tooltipWidth > bodyWidth) {
              // 如果 tooltip 超出右边界，则左对齐
              tooltipX = pageX - tooltipWidth
            }

            const htmlContent = `<div class='${styles.tooltip}'">
                    <div class='${styles.title}'>${names[level[j + 3]]}</div>
                    <div class=''>占比
                    ：${percentage} %</div>
                    <div class=''>样本数量：${new Intl.NumberFormat().format(level[j + 1])}</div>
                </div>`
            // 设置 tooltip 的准确位置
            tooltip.attr(
              'style',
              `z-index: 99; visibility: visible; position: fixed; left: ${tooltipX + 10}px; top: ${clientY + 10}px`,
            )
            d3.select('#chart_tooltip').html(htmlContent)
          })

        // 动态边框绘制
        if (rectWidth > 5) {
          svg
            .append('path')
            .attr(
              'd',
              `M${xOffset},${y} L${xOffset},${y + layerHeight} L${xOffset + rectWidth},${y + layerHeight} L${xOffset + rectWidth},${y} Z`,
            )
            .attr('stroke', '#58547c')
            .attr('stroke-width', 0.2)
            .attr('fill', 'none')
        }

        // 定义文字及 clipPath
        g.append('clipPath')
          .attr('id', `clip-${levelIndex}-${j}`)
          .append('rect')
          .attr('x', xOffset)
          .attr('y', y)
          .attr('width', rectWidth)
          .attr('height', layerHeight)

        g.append('text')
          .attr('class', styles.flameText)
          .attr('x', xOffset + 5)
          .attr('y', y + layerHeight / 2)
          .attr('clip-path', `url(#clip-${levelIndex}-${j})`)
          .text(names[level[j + 3]] + `  (${percentage}%)`)
          .style('display', rectWidth < 20 ? 'none' : 'block')
          .style('text-anchor', 'start')
          .style('dominant-baseline', 'middle')
          .style('overflow', 'hidden')
          .style('text-overflow', 'ellipsis')
          .style('white-space', 'nowrap')

        // console.log([xOffset, xOffset + rectWidth])
        xOffset += rectWidth
      }
      // level.forEach((d) => {
      //   const rectWidth = xScale(d.value) // 计算矩形宽度

      //   // 绘制矩形
      //   g.append('rect')
      //     .attr('class', 'flame-rect')
      //     .attr('x', xOffset)
      //     .attr('y', y)
      //     .attr('GRAPH_WIDTH', rectWidth)
      //     .attr('height', layerHeight - 2) // 加上间距
      //     .attr('fill', color(d.name))
      //     .on('click', () => console.log(`Clicked: ${d.name}`))

      //   // 添加文字
      //   g.append('text')
      //     .attr('class', 'flame-text')
      //     .attr('x', xOffset + rectWidth / 2) // 居中对齐
      //     .attr('y', y + layerHeight / 2 + 4) // 垂直居中
      //     .text(d.name)
      //     .style('display', rectWidth < 20 ? 'none' : 'block') // 太窄的矩形隐藏文字

      //   xOffset += rectWidth // 更新起始位置
      // })
    })
  }
  // console.log(props.data, flamegraphRef.current)
  useEffect(() => {
    // window.addEventListener("resize", onResize);

    if (props.data) {
      const dom = flamegraphRef.current
      if (dom) {
        drawFlamegraph(dom)
      }
    }
    return () => {
      d3.select('.flamegraph').selectAll('svg').remove()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [props.data])

  return (
    <div className="event_detail mb-3 max-h-[500px] overflow-y-auto overflow-x-hidden">
      {props.data ? (
        <React.Fragment>
          <div className="flamegraph h-full" ref={flamegraphRef} />
          <div id="chart_tooltip"></div>
        </React.Fragment>
      ) : (
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} />
      )}
    </div>
  )
}
