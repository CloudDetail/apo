import { useEffect, useRef, useState } from 'react'
import { FlamegraphCanvasProps, FocusedItemData, LEVEL_HEIGHT, TooltipItemData } from './type'
import Toolbox from './components/Toolbox'
import CanvasRender from './components/CanvasRender'
import FlamegraphTooltip from './components/FlamegraphTooltip'

function FlamegraphCanvas(props: FlamegraphCanvasProps) {
  const canvasRef = useRef<HTMLCanvasElement>(null)
  const { levels, names, numTicks, timeRange } = props
  const [barPositions, setbarPositions] = useState<any>()
  const [tooltipItem, setTooltipItem] = useState<TooltipItemData | null>(null)
  const [focusedItem, setFocusedItem] = useState<FocusedItemData | null>(null)
  const [mousePosition, setMousePosition] = useState<{ x: number; y: number } | null>(null)
  const [searchValue, setSearchValue] = useState<string>('')
  const [rangeMin, setRangeMin] = useState<number>(0)
  const [rangeMax, setRangeMax] = useState<number>(1)
  const onMouseMove = (e: any) => {
    const rect = canvasRef.current?.getBoundingClientRect()
    if (!rect) {
      throw new Error('Could not get rect')
    }
    const mouseX = e.clientX - rect?.left || 0
    const mouseY = e.clientY - rect?.top || 0

    // 遍历条目，检查鼠标是否悬浮在某个条目上
    const hoveredBar = barPositions.find(
      (bar: any) =>
        mouseX >= bar.x &&
        mouseX <= bar.x + bar.width &&
        mouseY >= bar.y &&
        mouseY <= bar.y + bar.height,
    )

    if (hoveredBar) {
      setMousePosition({ x: e.clientX, y: e.clientY })
      setTooltipItem(hoveredBar.data)
      // 在这里可以更新一个提示框或高亮当前条目
    } else {
      setTooltipItem(null)
      setMousePosition(null)
    }
  }
  const onFocus = (e: any) => {
    const rect = canvasRef.current?.getBoundingClientRect()
    if (!rect) {
      throw new Error('Could not get rect')
    }
    const mouseX = e.clientX - rect?.left || 0
    const mouseY = e.clientY - rect?.top || 0

    // 遍历条目，检查鼠标是否悬浮在某个条目上
    const clickBar = barPositions.find(
      (bar: any) =>
        mouseX >= bar.x &&
        mouseX <= bar.x + bar.width &&
        mouseY >= bar.y &&
        mouseY <= bar.y + bar.height,
    )

    if (clickBar) {
      setMousePosition(null)
      setTooltipItem(null)
      setFocusedItem({ lIndex: clickBar.lIndex, bIndex: clickBar.bIndex })
      const zoomMin = clickBar.data.offset / numTicks
      const zoomMax = (clickBar.data.offset + clickBar.data.totalSamples) / numTicks
      setRangeMin(zoomMin)
      setRangeMax(zoomMax)
      // 在这里可以更新一个提示框或高亮当前条目
    }
  }
  const resetView = () => {
    setFocusedItem(null)
    setRangeMin(0)
    setRangeMax(1)
  }

  const onMouseOut = () => {
    setTooltipItem(null)
    setMousePosition(null)
  }
  useEffect(() => {
    if (canvasRef.current) {
      const canvasProps = {
        canvas: canvasRef.current,
        levels,
        names,
        numTicks,
        rangeMin,
        rangeMax,
        searchValue,
        focusedItem,
        setbarPositions,
      }
      CanvasRender(canvasProps) // 在确保 canvas 不为空时调用
    }
  }, [
    canvasRef.current, // 依赖 canvasRef.current 的变化
    levels,
    names,
    numTicks,
    rangeMin,
    rangeMax,
    searchValue,
    focusedItem,
  ])

  return (
    <>
      <Toolbox
        searchValue={searchValue}
        setSearchValue={setSearchValue}
        isFocused={focusedItem !== null}
        resetView={resetView}
        timeRange={timeRange}
      />
      <FlamegraphTooltip position={mousePosition} tooltipItem={tooltipItem} />
      <div className="h-full overflow-auto">
        <div className="px-2" style={{ height: levels.length * LEVEL_HEIGHT + 100 }}>
          <canvas
            ref={canvasRef}
            className="w-full"
            onMouseMove={onMouseMove}
            onClick={onFocus}
            onMouseOut={onMouseOut}
          />
        </div>
      </div>
    </>
  )
}
export default FlamegraphCanvas
