import { TooltipItemData } from '.././type'
import styles from '.././index.module.scss'
import { useEffect, useRef } from 'react'

interface FlamegraphTooltipProps {
  position: { x: number; y: number } | null
  tooltipItem?: TooltipItemData | null
}
const FlamegraphTooltip = (props: FlamegraphTooltipProps) => {
  const { position, tooltipItem } = props
  const tooltipRef = useRef(null)
  useEffect(() => {
    // 如果有 TooltipItem 并且 tooltipRef 存在
    if (tooltipItem && tooltipRef.current) {
      const tooltip = tooltipRef.current

      const { x, y } = position
      const bodyWidth = document.body.clientWidth
      const tooltipWidth = tooltip.getBoundingClientRect().width

      // 计算 tooltip 的水平位置
      let tooltipX = x
      if (x + tooltipWidth > bodyWidth) {
        tooltipX = x - tooltipWidth // 如果超出右边界，调整到左侧
      }
      // 设置 tooltip 的样式
      tooltip.style.left = `${tooltipX + 10}px` // 鼠标右下角偏移
      tooltip.style.top = `${y + 10}px`
      tooltip.style.visibility = 'visible'
    } else if (tooltipRef.current) {
      // 隐藏 tooltip
      tooltipRef.current.style.visibility = 'hidden'
    }
  }, [position, tooltipItem])
  return (
    <>
      <div ref={tooltipRef} className={`${styles.tooltip}`}>
        <div className={styles.title}>{tooltipItem?.funcName}</div>
        <div>占比：{tooltipItem?.percentage}%</div>
        <div>样本数量：{new Intl.NumberFormat().format(tooltipItem?.totalSamples || 0)}</div>
      </div>
    </>
  )
}
export default FlamegraphTooltip
