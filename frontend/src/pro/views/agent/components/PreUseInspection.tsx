import { useTranslation } from 'react-i18next'
import GlassCard from 'src/core/components/GlassCard'

const PreUseInspection = () => {
  const { t } = useTranslation('pro/agent')
  return (
    <GlassCard
      content={
        <div>
          <p>
            情况1:尚未安装minimal-agent,请先
            <a className="text-blue-500" href="/#/probe-installer">
              安装
            </a>
          </p>
          <p>
            情况2:尚未对接任何可观测数据,请先
            <a className="text-blue-500" href="/#/agent-datasource">
              配置
            </a>
          </p>
        </div>
      }
    />
  )
}

export default PreUseInspection
