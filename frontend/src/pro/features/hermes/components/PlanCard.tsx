import { CheckCircleOutlined, OrderedListOutlined } from '@ant-design/icons'
import { Tag } from 'antd'
import type { HermesPlanTimelineItem } from '../client/types.ts'
import styles from '../HermesAgentWorkspace.module.scss'

export default function PlanCard({ item }: { item: HermesPlanTimelineItem }) {
  return (
    <section className={styles.planCard}>
      <div className={styles.planCardIcon}>
        <OrderedListOutlined />
      </div>
      <div className={styles.planCardContent}>
        <div className={styles.planCardHeading}>
          <span>模型规划</span>
          <strong>分析方案</strong>
          <Tag bordered={false} color="processing">
            <CheckCircleOutlined /> 已生成
          </Tag>
        </div>
        <ol className={styles.planSteps}>
          {item.plan.steps.map((step) => (
            <li key={step.id}>{step.title}</li>
          ))}
        </ol>
      </div>
    </section>
  )
}
