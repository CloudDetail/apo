import styles from '../../index.module.scss'
import type { ReportTimelineItem } from '../../reportViewModel'

interface AnalysisProcessProps {
  timeline?: ReportTimelineItem[]
}

function timelineRange(time?: string) {
  if (!time?.includes('-')) return undefined
  const [start, end] = time.split('-').map((part) => part.trim())
  if (!start || !end) return undefined
  return { start, end, label: `${start} - ${end}` }
}

const AnalysisProcess = ({ timeline }: AnalysisProcessProps) => {
  if (!timeline?.length) {
    return null
  }

  return (
    <ol className={styles.timelineList}>
      {timeline.map((item, index) => {
        const range = timelineRange(item.time)
        return (
          <li
            className={range ? styles.timelineInterval : undefined}
            key={`${item.time || 'step'}-${item.title}-${index}`}
          >
            <span className={styles.timelineTime}>
              {range?.start || item.time || `步骤 ${index + 1}`}
            </span>
            <span className={styles.timelineMarker} aria-hidden="true" />
            <div className={styles.timelineContent}>
              {range && <em className={styles.timelineDuration}>{range.label} 持续异常</em>}
              <strong>{item.title}</strong>
              {item.detail && <p>{item.detail}</p>}
            </div>
          </li>
        )
      })}
    </ol>
  )
}

export default AnalysisProcess
