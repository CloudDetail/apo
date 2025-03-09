import Color from 'color'
import { getColor, signleFlame } from '.././flamebearer'
import {
  CanvasRenderProps,
  LEVEL_HEIGHT,
  MAIN_BAR_WIDTH,
  MAIN_TEXT_BAR_WIDTH,
  TooltipItemData,
} from '.././type'

function CanvasRender(props: CanvasRenderProps) {
  const {
    canvas,
    levels,
    names,
    numTicks,
    rangeMin,
    rangeMax,
    searchValue,
    focusedItem,
    setbarPositions,
  } = props
  const ratio = window.devicePixelRatio || 1
  const barPositions = []
  const sf = signleFlame
  const graphWidth = canvas?.clientWidth || 0
  canvas.width = graphWidth * ratio
  const canvasHeight = LEVEL_HEIGHT * levels.length
  //  const canvasHeight = PX_PER_LEVEL * (levels.length - topLevel);
  canvas.height = canvasHeight * ratio

  //每一个刻度的宽度
  const tickWidth = graphWidth / numTicks / (rangeMax - rangeMin)
  const ctx = canvas.getContext('2d')
  if (!ctx) {
    throw new Error('Could not get ctx')
  }
  ctx.scale(ratio, ratio)
  for (let lIndex = 0; lIndex < levels.length; lIndex += 1) {
    const level = levels[lIndex]
    let offset = 0
    const y = lIndex * LEVEL_HEIGHT
    for (let bIndex = 0; bIndex < level.length; bIndex += sf.step) {
      offset += sf.getOffset(level, bIndex)
      const x = getBarX(offset, numTicks, rangeMin, tickWidth)

      let totalSamples = sf.getTotalSamples(level, bIndex)

      const funcName = getFunctionName(names, bIndex, level)

      const highlightModeOn = (!!searchValue && searchValue.length > 0) || focusedItem

      const isHighlighted =
        nodeIsInQuery(funcName, searchValue) && lIndex > (focusedItem?.lIndex || 0) - 1

      // 判断是否需要省略呈灰色展示
      const isFold = tickWidth * totalSamples <= MAIN_BAR_WIDTH
      if (isFold) {
        while (
          bIndex < level.length - sf.step &&
          sf.getOffset(level, bIndex + sf.step) === 0 &&
          tickWidth * sf.getTotalSamples(level, bIndex + sf.step) <= MAIN_BAR_WIDTH
        ) {
          bIndex += sf.step
          totalSamples += sf.getTotalSamples(level, bIndex)
        }
      }
      const sw = tickWidth * totalSamples
      //
      const color = isFold ? Color.rgb(200, 200, 200).alpha(0.4) : getColor(funcName)

      ctx.beginPath()
      ctx.rect(x, y, sw, LEVEL_HEIGHT)
      ctx.fillStyle = color.string()
      ctx.fill()

      // 判断是否需要叠加灰色蒙层
      if (highlightModeOn && !isHighlighted) {
        ctx.beginPath()
        ctx.rect(x, y, sw, LEVEL_HEIGHT)
        ctx.fillStyle = Color.rgb(0, 0, 0).alpha(0.4).string() // 灰色蒙层，半透明
        ctx.fill()
      }
      const present = totalSamples / numTicks // 计算比例
      const percentage = (present * 100).toFixed(2).replace(/\.?0+$/, '') // 转换为百分比
      if (!isFold) {
        const data: TooltipItemData = {
          funcName,
          totalSamples,
          selfSamples: sf.getSelfSamples(level, bIndex),
          percentage: percentage,
          offset: offset,
        }
        barPositions.push({
          x,
          y,
          width: sw,
          height: LEVEL_HEIGHT,
          lIndex: lIndex, // 当前层级
          bIndex: bIndex,
          data: data,
        })
      }
      offset += totalSamples

      // 绘字
      if (isFold || sw < MAIN_TEXT_BAR_WIDTH) {
        continue
      }
      ctx.textBaseline = 'middle'
      ctx.font = '400 11.5px SFMono-Regular, Consolas, Liberation Mono, Menlo, monospace'
      const characterSize = ctx.measureText('a').width
      ctx.save()
      ctx.clip()
      ctx.fillStyle = 'black'
      const namePosX = Math.round(Math.max(x, 0))
      ctx.fillText(funcName, namePosX + 5, y + LEVEL_HEIGHT / 2 + 1)
      ctx.restore()
    }
  }
  setbarPositions(barPositions)
}
function getBarX(offset: number, totalTicks: number, rangeMin: number, perTick: number) {
  return (offset - totalTicks * rangeMin) * perTick
}
export function getFunctionName(names: string[], index: number, level: number[]) {
  let l = signleFlame.getFuncNameIndex(level, index)
  if (l === undefined) {
    l = -1
  }
  const name = names[l]

  if (!name) {
    console.warn('Error', level)
    return ''
  }
  return name
}
function nodeIsInQuery(funcName: string, query: string) {
  if (!funcName) {
    return false
  }

  return funcName.toLowerCase().includes(query.toLowerCase())
}

export default CanvasRender
