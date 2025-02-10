export interface TooltipItemData {
  funcName: string
  totalSamples: number
  selfSamples: number
  percentage: string
  offset: number
}
export interface FocusedItemData {
  lIndex: number
  bIndex: number
}

export interface FlamegraphCanvasProps {
  levels: number[][]
  names: string[]
  numTicks: number
}
export interface CanvasRenderProps extends FlamegraphCanvasProps {
  canvas: any
  rangeMax: number
  rangeMin: number
  searchValue: string
  focusedItem: FocusedItemData | null
  setbarPositions: any
}
export const LEVEL_HEIGHT = 22
export const MAIN_TEXT_BAR_WIDTH = 30
export const MAIN_BAR_WIDTH = 5
