import styles from './index.module.scss'
const GlassCard = () => {
  return (
    <div className={styles.glassContainer}>
      <div className={styles.glassEffect}>
        <div>
          {/* <h1>毛玻璃效果</h1> */}
          <p>
            尚未配置大语言模型 API 访问权限，请参阅
            <a
              className="underline text-sky-500"
              target="_blank"
              href="https://kindlingx.com/docs/APO%20向导式可观测性中心/配置指南/APO接入大语言模型分析告警事件"
            >
              接入大语言模型手册
            </a>
            进行配置
          </p>
        </div>
      </div>
    </div>
  )
}
export default GlassCard
