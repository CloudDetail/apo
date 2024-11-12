import * as d3 from 'd3'
import dagreD3 from 'dagre-d3'
import React, { useEffect, useRef } from 'react'
import _ from 'lodash'
import nodeRed from 'src/core/assets/images/hexagon-red.svg'
import { useNavigate, useSearchParams } from 'react-router-dom'
import { usePropsContext } from 'src/core/contexts/PropsContext'
import Empty from 'src/core/components/Empty/Empty'
import { useSelector } from 'react-redux'
import { TimestampToISO } from 'src/core/utils/time'
function escapeId(id) {
  return id.replace(/[^a-zA-Z0-9-_]/g, '_')
}
function splitTextToFitWidth(text, maxWidth) {
  const maxCharsPerLine = Math.floor(maxWidth / 10)
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
  })

  if (currentLine) {
    lines.push(currentLine)
  }

  return lines
}
export const ErrotChain = React.memo(function ErrotChain(props) {
  const { data, chartId = 'errorChain' } = props
  const { serviceName, endpoint } = usePropsContext()
  const [searchParams, setSearchParams] = useSearchParams()
  // const mutatedRef = useRef(props.instance)
  // mutatedRef.current = props.instance
  const navigate = useNavigate()
  const { startTime, endTime } = useSelector((state) => state.timeRange)
  function draw() {
    const container = d3.select(`#${escapeId(chartId)}`)
    // 清除之前的内容
    container.selectAll('*').remove()
    const svg = container.append('svg').attr('width', '100%').attr('height', '100%')
    // const svgWidth = svg.node()?.getBoundingClientRect().width ?? 0;
    // const svgHeight = svg.node()?.getBoundingClientRect().height ?? 0;
    const labelSize = 16
    const inner = svg.append('g')

    const g = new dagreD3.graphlib.Graph()
      .setGraph({
        // width: svgWidth,
        // height: svgHeight,
        rankdir: 'LR',
        // 适应窄宽区域的边距和节点间距
        // marginx: 20,
        // marginy: 20,
        edgesep: 20,
        nodesep: 60,
        ranksep: 100,
      })
      .setDefaultEdgeLabel(function () {
        return {}
      })
    const nodeWidth = 50
    const nodeHeight = 50

    const current = data.current

    // 处理数据
    g.setNode(current.instance, {
      label: current.instance,
      width: nodeWidth,
      height: nodeHeight,
      id: escapeId(chartId + 'node-current-' + current.instance),
      service: current.service,
    })
    data.parents.forEach((data) => {
      g.setNode(data.instance, {
        label: data.instance,
        width: nodeWidth,
        height: nodeHeight,
        id: escapeId(chartId + 'node-parent-' + data.instance),
        service: data.service,
      })
      g.setEdge(data.instance, current.instance, {
        //                     style: "stroke: #f66; stroke-width: 3px; stroke-dasharray: 5, 5;",
        //   arrowheadStyle: "fill: #f66"

        id: escapeId(chartId + 'edge-parent-' + data.instance + '-current-' + current.instance),
        arrowhead: 'vee',
        arrowheadStyle: () => {
          let color = 'rgba(42, 130, 228, 1)'
          return `fill:${color};stroke: none;`
        },
        curve: d3.curveBasis,
        // curve: d3.curveLinear,
        arrowheadId: escapeId(
          chartId + 'arrow-parent-' + data.instance + '-current-' + current.instance,
        ),
        style:
          'stroke: rgba(42, 130, 228, 1); stroke-width: 3px; stroke-dasharray: 5, 5;fill: none', // 5px实线，5px空白
      })
    })
    data.children.forEach((data) => {
      g.setNode(data.instance, {
        label: data.instance,
        width: nodeWidth,
        height: nodeHeight,
        id: escapeId(chartId + 'node-child-' + data.instance),
        service: data.service,
      })
      g.setEdge(current.instance, data.instance, {
        //                     style: "stroke: #f66; stroke-width: 3px; stroke-dasharray: 5, 5;",
        //   arrowheadStyle: "fill: #f66"
        id: escapeId(chartId + 'edge-current-' + current.instance + '-child-' + data.instance),
        arrowhead: 'vee',
        arrowheadStyle: () => {
          let color = 'rgba(42, 130, 228, 1)'
          return `fill:${color};stroke: none;`
        },
        curve: d3.curveBasis,
        // curve: d3.curveLinear,
        arrowheadId: escapeId(
          chartId + 'arrow-current-' + current.instance + '-child-' + data.instance,
        ),
        style:
          'stroke: rgba(42, 130, 228, 1); stroke-width: 3px; stroke-dasharray: 5, 5;fill: none', // 5px实线，5px空白
      })
    })
    // 创建渲染器并准备 SVG 容器
    const render = new dagreD3.render()
    // 运行渲染器
    //@ts-ignore
    render(inner, g)
    inner
      .selectAll('g.node')
      .append('circle')
      .attr('x', 0)
      .attr('y', 0)
      .attr('r', 25)
      .attr('width', nodeWidth)
      .attr('height', nodeHeight)
      .style('rx', '50%')
      .style('ry', '50%')
      .style('fill', 'none')
      .style('stroke', 'none')
    // .each(function (d, i) {
    //   if (d === instance) {
    //     d3.select(this)
    //       .transition()
    //       .ease(d3.easeExpOut)
    //       .duration(1000)
    //       .style('fill', 'rgba(255, 208, 0, 0.3)')
    //       .style('stroke', 'rgba(255, 208, 0, 0.2)')
    //       .style('stroke-width', '15px')
    //   }
    // })
    // 定义交互筛选

    // const allNodeNames = g.nodes()
    // console.log(allNodeNames)
    // allNodeNames.forEach((nodeName) => {
    //   d3.selectAll(`g[id^="${escapeId(chartId)}"].node`)
    //     .on('click', function () {
    //       d3.selectAll('circle').interrupt()
    //       d3.selectAll('.bg-circle')
    //         .style('fill', 'none')
    //         .style('stroke', 'none')
    //         .style('stroke-width', '0')
    //       d3.select(this)
    //         .select('circle')
    //         .transition()
    //         .ease(d3.easeExpOut) // 缓动函数，可以根据需要调整
    //         .duration(1000) // 动画持续时间
    //         .attr('class', 'bg-circle')
    //         .style('fill', 'rgba(255, 208, 0, 0.3)')
    //         .style('stroke', 'rgba(255, 208, 0, 0.2)')
    //         .style('stroke-width', '15px')
    //       const nodeData = g.node(nodeName) // 获取节点的所有数据
    //       console.log(nodeName, nodeData)
    //       // navigate(`/logs?service=${nodeData.service}&endpoint=${endpoint}&instance=${nodeName}`)
    //     })
    //     .on('mouseover', function (d) {
    //       d3.select(this).style('cursor', 'pointer')
    //     })
    // })
    inner.selectAll('g.node').each(function (nodeName) {
      const node = d3.select(this)
      // 为当前节点添加点击事件
      node
        .on('click', () => {
          // console.log(`${nodeName} 被点击了`)
          const nodeData = g.node(nodeName)
          // console.log('节点数据:', nodeData)
          const from = TimestampToISO(startTime)
          const to = TimestampToISO(endTime)

          navigate(
            `/logs/fault-site?service=${nodeData.service}&endpoint=${endpoint}&instance=${nodeName}&logs-from=${from}&logs-to=${to}`,
          )
        })
        .on('mouseover', function (d) {
          d3.select(this).style('cursor', 'pointer')
        })
    })
    // // 自定义节点样式为图像

    inner
      .selectAll('g.node')
      .append('image')
      .attr('x', 0)
      .attr('y', 0)
      .attr('xlink:href', nodeRed)
      .attr('width', nodeWidth)
      .attr('height', nodeHeight)
      .attr('transform', `translate(0,0)`)
    inner.selectAll('g.node').selectAll('g.node rect').style('fill', 'none').style('stroke', 'none')
    inner.selectAll('g.node tspan').style('fill', '#ffffff').style('font-size', labelSize)
    inner.selectAll('g.label').attr('transform', `translate(${nodeWidth},${nodeHeight - 5})`)
    inner.selectAll('g.node').each(function (d) {
      const node = _.find(data, { instance: d })
      if (typeof d === 'string') {
        const splitText = splitTextToFitWidth(d, 160)
        d3.select(this).select('tspan').remove()
        const text = d3
          .select(this)
          .append('g')
          .attr('transform', `translate(${-nodeWidth - labelSize},${nodeHeight / 2 + labelSize})`)

        splitText.forEach((part, index) => {
          text
            .append('text')
            .text(part)
            .attr('x', 0)
            .attr('y', 0 + index * labelSize)
            .style('font-size', labelSize)
            .style('fill', '#ffffff')
        })
      }
    })
    inner.selectAll('g.edgePath path').style('stroke-width', 2).style('stroke-dasharray', '5,5')
    inner
      .selectAll('g.node image')
      .attr('x', -nodeWidth / 2)
      .attr('y', -nodeHeight / 2)
      .attr('width', nodeWidth)
      .attr('height', nodeHeight)

    //@ts-ignore
    let graphWidth = +container.node()?.getBoundingClientRect().width
    //@ts-ignore
    let graphHeight = +container.node()?.getBoundingClientRect().height
    let padding = g.nodes().length > 2 ? 0 : 100
    let zoomScale = Math.min(
      3,
      0.8 *
        //@ts-ignore
        Math.min(
          (graphWidth - padding) / g.graph().width,
          (graphHeight - padding) / g.graph().height,
        ),
    )
    //@ts-ignore
    const xCenterOffset = (graphWidth - g.graph().width * zoomScale) / 2
    //@ts-ignore
    const yCenterOffset = (graphHeight - g.graph().height * zoomScale) / 2
    inner.attr(
      'transform',
      'translate(' + xCenterOffset + ',' + yCenterOffset / 2 + ') scale(' + zoomScale + ')',
    )

    // 缩放
    const zoom = d3.zoom().on('zoom', function () {
      let currentZoomTransform = d3.zoomTransform(this)
      inner.attr(
        'transform',
        `translate(${currentZoomTransform.x + xCenterOffset},${
          currentZoomTransform.y + yCenterOffset / 2
        }) scale(${zoomScale + currentZoomTransform.k - 1})`,
      )
    })
    // @ts-ignore
    svg.call(zoom)
  }
  useEffect(() => {
    if (data?.current) {
      draw()
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data])

  // useEffect(() => {
  //   // console.log(d3.select('#arrow'))
  //   d3.selectAll('circle').interrupt()
  //   d3.selectAll('.bg-circle')
  //     .style('fill', 'none')
  //     .style('stroke', 'none')
  //     .style('stroke-width', '0')
  //   d3.select('g#node-' + instance + '.node')
  //     .select('circle')
  //     .transition()
  //     .ease(d3.easeExpOut) // 缓动函数，可以根据需要调整
  //     .duration(1000) // 动画持续时间
  //     .attr('class', 'bg-circle')
  //     .style('fill', 'rgba(255, 208, 0, 0.3)')
  //     .style('stroke', 'rgba(255, 208, 0, 0.2)')
  //     .style('stroke-width', '15px')
  // }, [instance])
  return (
    <div style={{ width: '100%', height: '100%' }}>
      {data?.current ? (
        <div
          id={escapeId(chartId)}
          className="topology-container"
          style={{ width: '100%', height: '100%' }}
        ></div>
      ) : (
        <Empty />
      )}
    </div>
  )
})
