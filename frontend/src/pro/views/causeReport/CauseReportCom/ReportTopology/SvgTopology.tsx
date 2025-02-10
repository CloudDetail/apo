import React, { useEffect, useRef } from 'react'
import * as d3 from 'd3'
import _ from 'lodash'

import nodeGreen from 'src/core/assets/images/node/hexagon-green.svg'
import nodeRed from 'src/core/assets/images/node/hexagon-red.svg'
import nodeGrey from 'src/core/assets/images/node/hexagon-grey.svg'
import unknowIcon from 'src/core/assets/images/unknown.png'
import { formatUnit } from 'src/core/utils/time'
import { ThemeColor, ThemeStyle } from 'src/constants'

interface Props {
  data: any
  category: string
  chartId?: string
}
function SingleLinkTopology({ data, category, chartId = 'nodeChart' }: Props) {
  //   const { theme } = useSelector((state) => state.settingReducer)
  const chartRef = useRef<any>()
  const theme = ThemeStyle['dark']
  const getNodeOrder = ({ nodes, edges }: any) => {
    let allNodeIds = nodes.map((node: any) => node.chartId)
    // 构造根据edge的调用顺序从头到底排序的nodeId数组
    let nodeIdByEdge: string[] = []
    allNodeIds.forEach((id: string) => {
      if (_.findIndex(edges, { target: id }) === -1) {
        nodeIdByEdge.push(id)
      }
    })
    while (_.findIndex(edges, { source: nodeIdByEdge[nodeIdByEdge.length - 1] }) > -1) {
      let edge = _.find(edges, { source: nodeIdByEdge[nodeIdByEdge.length - 1] })
      nodeIdByEdge.push(edge.target)
    }
    return nodeIdByEdge
  }
  const draw = () => {
    const container: any = document.getElementById(chartId)
    if (data.nodes.length === 0) {
      container.innerHTML = ''
      return
    }
    container.innerHTML = ''
    const nodeOrder = getNodeOrder(data)
    const maxTime = _.chain(data.nodes)
      .map((node) => [node.p90, node.totalTime])
      .flatten()
      .max()
      .value()
    const maxTotalTime = Math.max(...data.nodes.map((node: any) => node.totalTime))
    // const WIDTH = chartRef.current.clientWidth
    const WIDTH = data.nodes.length * 310
    let HEIGHT = data.nodes.length * 310
    if (category === 'chain') {
      HEIGHT -= 50
    }
    const svg = d3
      .select(`#${chartId}`)
      .append('svg')
      .style('height', '100%')
      .style('width', HEIGHT)
    const g = svg.append('g').attr('id', `${chartId}-chart_warp`).attr('class', 'chart_warp')
    const nodeWarp = g.append('g').attr('id', `${chartId}-node_warp`)
    const LineWarp = g.append('g').attr('id', `${chartId}-line_warp`)

    const defs = LineWarp.append('defs') // defs定义可重复使用的元素
    const arrowheads = defs
      .append('marker') // 创建箭头
      .attr('id', `${chartId}-arrow`)
      .attr('markerUnits', 'strokeWidth')
      .attr('markerWidth', 16)
      .attr('markerHeight', 12)
      .attr('viewBox', '0 0 10 10')
      .attr('refX', 9)
      .attr('refY', 5)
      .attr('orient', 'auto')
    arrowheads
      .append('path')
      .attr('d', 'M 0 0 L 10 5 L 0 10 L 4 5 z')
      .attr('fill', theme.colors.text.primary)

    // const fontSize = 14;

    function drawNode(node: any, idx: number) {
      let nodeG = nodeWarp
        .append('g')
        .attr('class', 'cluster')
        .attr('transform', `translate(${idx * 300},0)`)
      nodeG
        .append('image')
        .attr('id', `${chartId}-node${idx}`)
        .attr('class', 'node_image')
        .attr('x', 20)
        .attr('y', 30)
        .attr('width', 50)
        .attr('height', 50)
        .attr('xlink:href', () => {
          let imageIcon: any = null
          if (category === 'error') {
            imageIcon = node.isError ? nodeRed : nodeGreen
          } else if (category === 'slow') {
            imageIcon = node.isMutated ? nodeRed : nodeGreen
          } else {
            return nodeRed
          }
          return imageIcon
        })
      function splitTextToFitWidth(text: string, maxWidth: number) {
        const maxCharsPerLine = Math.floor(maxWidth / 10)
        const maxLines = 2 // 限制行数
        const words = text.split('')
        const lines = []
        let currentLine = ''

        words.forEach((word) => {
          if ((currentLine + word).length > maxCharsPerLine) {
            lines.push(currentLine)
            currentLine = word
          } else {
            currentLine += word
          }

          // 检查是否已达到最大行数
          if (lines.length === maxLines - 1 && currentLine.length > maxCharsPerLine) {
            // 如果即将超出两行，则截断并添加省略号
            currentLine = currentLine.slice(0, maxCharsPerLine - 1) + '...'
            lines.push(currentLine)
            currentLine = ''
          }
        })

        // 把最后一行加入到结果中（如果有的话）
        if (currentLine && lines.length < maxLines) {
          lines.push(currentLine)
        }

        return lines
      }

      // 使用上述函数分割文本
      const splitText = splitTextToFitWidth(node.id, 300)
      nodeG
        .append('text')
        .text('节点名:')
        .attr('x', 30) // X位置，根据需要调整
        .attr('y', 100) // Y位置，简单地用索引*行高来计算
        .style('font-size', '14')
        .style('fill', theme.colors.text.primary)
      // 为每部分文本创建一个<text>元素
      splitText.forEach((part, index) => {
        nodeG
          .append('text')
          .text(part)
          .attr('x', 80) // X位置，根据需要调整
          .attr('y', 100 + index * 15) // Y位置，简单地用索引*行高来计算
          .style('font-size', '14')
          .style('fill', theme.colors.text.primary)
      })
      const timeY = 145
      const timeX = 125
      // const selfTimeMaxWidth = (WIDTH - timeX) * 0.4
      // const otherTimeMaxWidth = (WIDTH - timeX) * 0.6
      const selfTimeMaxWidth = (300 - timeX) * 0.4
      const otherTimeMaxWidth = (300 - timeX) * 0.6
      const timeTextYOffest = 15
      // 中线
      if (category === 'slow') {
        nodeG
          .append('line')
          .attr('class', 'time_line')
          .attr(
            'style',
            `stroke-width: 1px;stroke: ${theme.colors.text.primary};stroke-dasharray: 5, 2;`,
          )
          .attr('x1', timeX + selfTimeMaxWidth + 1)
          .attr('y1', timeY - 20)
          .attr('x2', timeX + selfTimeMaxWidth + 1)
          .attr('y2', timeY + 65)
        nodeG
          .append('text')
          .text('selfTime')
          .attr('style', `font-size: 11px;font-weight: 500; fill: ${theme.colors.text.primary}`)
          .attr('x', function () {
            let textlen: any = this.getComputedTextLength()
            return timeX + selfTimeMaxWidth - textlen - 5
          })
          .attr('y', timeY - 10)

        nodeG
          .append('text')
          .text('otherTime')
          .attr('style', `font-size: 11px; font-weight: 500;fill: ${theme.colors.text.primary}`)
          .attr('x', timeX + selfTimeMaxWidth + 7)
          .attr('y', timeY - 10)

        nodeG
          .append('text')
          .text('历史基线')
          .attr('style', `font-size: 11px; fill: ${theme.colors.text.primary}`)
          .attr('x', 70)
          .attr('y', timeY + timeTextYOffest)
        const selfP90Width = selfTimeMaxWidth * (node.selfP90 / maxTime)
        const p90OtherWidth = otherTimeMaxWidth * ((node.p90 - node.selfP90) / maxTime)
        nodeG
          .append('rect')
          .attr('class', 'time_rect')
          .attr('width', selfP90Width < 2 ? 2 : selfP90Width)
          .attr('height', 20)
          .attr('fill', ThemeColor.green)
          .attr('x', timeX + selfTimeMaxWidth - selfP90Width - 2)
          .attr('y', timeY)
          .attr('rx', 2)
          .attr('ry', 2)

        nodeG
          .append('text')
          .text(formatUnit(node.selfP90, 'ns'))
          .attr('style', `font-size: 11px; font-weight: 500;fill: ${theme.colors.text.primary}`)
          .attr('x', function () {
            let textlen: any = this.getComputedTextLength()
            return timeX + selfTimeMaxWidth - textlen - 4
          })
          .attr('y', timeY + timeTextYOffest - 2)
        if (node.p90 - node.selfP90 >= 0) {
          nodeG
            .append('rect')
            .attr('class', 'time_rect')
            .attr('width', p90OtherWidth < 2 ? 2 : p90OtherWidth)
            .attr('height', 20)
            .attr('fill', ThemeColor.green)
            .attr('x', timeX + selfTimeMaxWidth + 4)
            .attr('y', timeY)
            .attr('rx', 2)
            .attr('ry', 2)
          nodeG
            .append('text')
            .text(formatUnit(node.p90 - node.selfP90, 'ns'))
            .attr('style', `font-size: 11px; font-weight: 500; fill: ${theme.colors.text.primary}`)
            .attr('x', timeX + selfTimeMaxWidth + 6)
            .attr('y', timeY + timeTextYOffest - 2)
        }
        // 绘制故障时间
        const selfWidth = selfTimeMaxWidth * (node.selfTime / maxTime)
        const OtherWidth = otherTimeMaxWidth * ((node.totalTime - node.selfTime) / maxTime)

        nodeG
          .append('text')
          .text('本次耗时')
          .attr('style', `font-size: 11px;font-weight: 500; fill: ${theme.colors.text.primary}`)
          .attr('x', 70)
          .attr('y', timeY + 34 + timeTextYOffest)
        nodeG
          .append('rect')
          .attr('class', 'time_rect')
          .attr('width', selfWidth < 2 ? 2 : selfWidth)
          .attr('height', 20)
          .attr('fill', '#f3485c')
          .attr('x', timeX + selfTimeMaxWidth - selfWidth - 2)
          .attr('y', timeY + 32)
          .attr('rx', 2)
          .attr('ry', 2)
        nodeG
          .append('text')
          .text(formatUnit(node.selfTime, 'ns'))
          .attr('style', `font-size: 11px; font-weight: 500;fill: ${theme.colors.text.primary}`)
          .attr('x', function () {
            let textlen: any = this.getComputedTextLength()
            return timeX + selfTimeMaxWidth - textlen - 4
          })
          .attr('y', timeY + 32 + timeTextYOffest)
        if (node.totalTime - node.selfTime >= 0) {
          nodeG
            .append('rect')
            .attr('class', 'time_rect')
            .attr('width', OtherWidth < 2 ? 2 : OtherWidth)
            .attr('height', 20)
            .attr('fill', '#f3485c')
            .attr('style', 'border-radius: 2px;')
            .attr('x', timeX + selfTimeMaxWidth + 4)
            .attr('y', timeY + 32)
            .attr('rx', 2)
            .attr('ry', 2)
          nodeG
            .append('text')
            .text(formatUnit(node.totalTime - node.selfTime, 'ns'))
            .attr('style', `font-size: 11px; font-weight: 500;fill: ${theme.colors.text.primary}`)
            .attr('x', timeX + selfTimeMaxWidth + 6)
            .attr('y', timeY + 32 + timeTextYOffest)
        }
      } else if (category === 'chain') {
        let totalTimeWidth
        if (node.totalTime === -1) {
          totalTimeWidth = 4 * selfTimeMaxWidth
        } else {
          totalTimeWidth = 2 * selfTimeMaxWidth * (node.totalTime / maxTotalTime)
        }
        nodeG
          .append('text')
          .text('服务端耗时')
          .attr('style', `font-size: 11px;font-weight: 500; fill: ${theme.colors.text.primary}`)
          .attr('x', 20)
          .attr('y', timeY)
        nodeG
          .append('rect')
          .attr('class', 'time_rect')
          .attr('width', totalTimeWidth)
          .attr('height', 20)
          .attr('fill', '#f3485c')
          .attr('style', 'border-radius: 2px;')
          .attr('x', 80)
          .attr('y', timeY - 15)
          .attr('rx', 2)
          .attr('ry', 2)
        nodeG
          .append('text')
          .text(node.totalTime === -1 ? '∞' : formatUnit(node.totalTime, 'ms'))
          .attr('style', `font-size: 11px; font-weight: 500;fill: ${theme.colors.text.primary}`)
          .style('font-size', node.totalTime === -1 ? 18 : 12)
          .attr('x', 80 + totalTimeWidth + 10)
          .attr('y', timeY)
      }
    }

    function drawEdge(edge: any, idx: number) {
      let points: any[] = []
      // @ts-ignore
      let nodeWarpRect = d3.select(`#${chartId}`).node()!.getBoundingClientRect()
      // @ts-ignore
      let sNodeRect: any = d3
        .select(`#${chartId}-node${nodeOrder.indexOf(edge.source)}`)
        .node()!
        .getBoundingClientRect()
      // @ts-ignore
      let tNodeRect: any = d3
        .select(`#${chartId}-node${nodeOrder.indexOf(edge.target)}`)
        .node()!
        .getBoundingClientRect()

      let sx = sNodeRect.x - nodeWarpRect.x,
        tx = tNodeRect.x - nodeWarpRect.x,
        sy = sNodeRect.bottom - nodeWarpRect.y - 25,
        ty = tNodeRect.top - nodeWarpRect.y + 25

      points.push({ x: sx + 50, y: sy }, { x: tx, y: ty })
      if (edge.missVNode || category === 'chain') {
        let startX = sx + 50,
          startY = sy,
          endX = tx,
          endY = ty

        // 计算中间点
        let midX = (startX + endX) / 2
        let midY = (endY + startY) / 2
        // 创建两条线段
        d3.select(`#${chartId}-line_warp`)
          .append('path')
          .attr('d', `M${startX},${startY} L${midX - 50},${midY}`)
          .attr('stroke', theme.colors.text.primary)
          .attr('stroke-width', 1)
          .attr('fill', 'none')

        d3.select(`#${chartId}-line_warp`)
          .append('path')
          .attr('d', `M${midX + 50},${midY} L${endX},${endY}`)
          .attr('stroke', theme.colors.text.primary)
          .attr('stroke-width', 1)
          .attr('fill', 'none')
          .attr('marker-end', `url(#${chartId}-arrow)`)
        if (category === 'chain') {
          d3.select(`#${chartId}-line_warp`)
            .append('text') // 添加文本元素
            .attr('x', midX) // 将文本放在中点的右侧，这里的 10 可以根据需要调整
            .attr('y', midY - 4)
            .attr('style', `font-size: 14px; font-weight: 500;fill: ${theme.colors.text.primary}`)
            .attr('text-anchor', 'middle') // 文本居中对齐
            .text(`${formatUnit(edge.nextCallTime, 'ms')}`) // 设置文本内容
          d3.select(`#${chartId}-line_warp`)
            .append('text') // 添加文本元素
            .attr('x', midX) // 将文本放在中点的右侧，这里的 10 可以根据需要调整
            .attr('y', midY + 14)
            .attr('style', `font-size: 11px; font-weight: 500;fill: ${theme.colors.text.primary}`)
            .attr('text-anchor', 'middle') // 文本居中对齐
            .text(`客户端请求耗时`) // 设置文本内容
        } else {
          // 添加省略号文本
          d3.select(`#${chartId}-line_warp`)
            .append('text')
            .attr('x', midX)
            .attr('y', midY)
            // .attr('dy', '.35em') // 根据需要调整以垂直居中
            .attr('fill', theme.colors.text.primary) // 使用fill属性设置文本颜色
            .style('font-size', '24px') // 放大文本，这里以24px为例
            .attr('text-anchor', 'middle') // 文本居中对齐
            // .attr('transform', `rotate(90, ${midX}, ${midY})`)
            .text('...')
        }
      } else {
        const dpath = d3
          .line()
          .x((d: any) => d.x)
          .y((d: any) => d.y)
          .curve(d3.curveBasis)(points)
        d3.select(`#${chartId}-line_warp`)
          .append('path')
          .attr('id', `#${chartId}-line${idx}`)
          .attr('class', 'line_path')
          .attr('style', `opacity: 1;stroke-width: 1;stroke: ${theme.colors.text.primary}`)
          .attr('fill', 'none')
          .attr('d', dpath)
          .attr('marker-end', `url(#${chartId}-arrow)`)
      }
    }
    function drawMissingNode(node: any, idx: number) {
      let nodeG = nodeWarp
        .append('g')
        .attr('class', 'cluster')
        .attr('transform', `translate(0, ${idx * 300})`)

      nodeG
        .append('image')
        .attr('id', `${chartId}-node${idx}`)
        .attr('class', 'node_image')
        .attr('x', 20)
        .attr('y', 30)
        .attr('width', 50)
        .attr('height', 50)
        .attr('xlink:href', () => {
          return nodeGrey
        })

      const unknowText = nodeG
        .append('text')
        .text('未知服务实例')
        .attr('x', 70) // X位置，根据需要调整
        .attr('y', 50) // Y位置，简单地用索引*行高来计算
        .style('font-size', '14')
        .style('fill', theme.colors.text.primary)
      const textWidth = unknowText.node()?.getComputedTextLength() ?? 0

      const iconX = 70 + textWidth + 10

      nodeG
        .append('image')
        .attr('x', iconX)
        .attr('y', 38)
        .attr('width', 14)
        .attr('height', 14)
        .attr('xlink:href', unknowIcon)
        .on('mouseenter', function (e) {
          // @ts-ignore
          let nodeWarpRect: any = d3.select(`#${chartId}`).node()!.getBoundingClientRect()
          let node: any = d3.select(this).node()!.getBoundingClientRect()
          d3.select(`#${chartId}unknowIconTooltip`).attr(
            'style',
            `visibility: visible; top: ${node.y - nodeWarpRect.y + 10}px; left: ${node.x - nodeWarpRect.x + 10}px;`,
          )
        })
        .on('mouseleave', (e: any) => {
          d3.select(`#${chartId}unknowIconTooltip`).attr('style', `visibility: hidden;`)
        })
    }
    // 绘制node节点

    _.forEach(nodeOrder, (id, idx) => {
      let node = _.find(data.nodes, { chartId: id })
      if (node.isMissing) {
        drawMissingNode(node, idx)
      } else {
        drawNode(node, idx)
      }
    })
    _.forEach(data.edges, (edge, idx: number) => {
      drawEdge(edge, idx)
    })
  }

  useEffect(() => {
    if (Object.keys(data).length > 0) {
      draw()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data])

  return (
    <div style={{ position: 'relative', width: '100%', height: '100%' }}>
      <div ref={chartRef} id={chartId} style={{ height: '100%' }}></div>
      <div id={chartId + 'unknowIconTooltip'} className="custom_tooltip">
        未采样或数据缺失
      </div>
    </div>
  )
}
export default SingleLinkTopology
