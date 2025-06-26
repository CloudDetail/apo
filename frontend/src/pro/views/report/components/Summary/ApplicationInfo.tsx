import { Row, Col, Button } from 'antd'
import { AiOutlineNodeIndex } from 'react-icons/ai'
import { LuServer, LuHash, LuBox, LuContainer } from 'react-icons/lu'
import { TagsParams } from '../../type'
import { t } from 'i18next'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

const i18nPrefix = 'pro/report:'
const InfoItemIconMap = {
  serviceName: <LuServer size={24} className="text-emerald-400" />,
  pod: <LuBox size={24} className="text-lime-400" />,
  containerId: <LuContainer size={24} className="text-lime-400" />,
  nodeName: <AiOutlineNodeIndex size={24} className="text-cyan-400" />,
  pid: <LuHash size={24} className="text-violet-400" />,
}
const InfoItem = ({ type, value }: { type: string; value: string }) => {
  const isServiceName = type === 'serviceName'
  const navigate = useNavigate()
  return (
    <div className="flex mb-2 items-center">
      <div className="rounded-md bg-[var(--ant-color-bg-layout)] p-2 ">{InfoItemIconMap[type]}</div>
      <div className="ml-4 flex flex-col justify-between">
        <div className="text-[var(--ant-color-text-secondary)] text-xs">
          {t(i18nPrefix + type)}
          {isServiceName && (
            <Button
              size="small"
              type="link"
              className="ml-2 text-xs"
              onClick={() => navigate('/demo/service-config')}
            >
              {t(i18nPrefix + 'configBusinessName')}
            </Button>
          )}
        </div>
        <div style={{ color: 'var(--ant-color-text-base)' }}>{value}</div>
      </div>
    </div>
  )
}
const TagKeyOrder: (keyof TagsParams)[] = ['serviceName', 'pod', 'containerId', 'nodeName', 'pid']
const ApplicationInfo = ({ tags }: { tags: TagsParams }) => {
  const [showMore, setShowMore] = useState(false)

  return (
    <Row className="w-full mt-2" justify="space-between" align="middle">
      {TagKeyOrder.filter((key) => key === 'serviceName' && tags[key] !== undefined).map((key) => (
        <Col key={key} flex={1}>
          <InfoItem type={key} value={tags[key]} />
        </Col>
      ))}

      {showMore &&
        TagKeyOrder.filter((key) => key !== 'serviceName' && tags[key] !== undefined).map((key) => (
          <Col key={key} flex={1}>
            <InfoItem type={key} value={tags[key]} />
          </Col>
        ))}
      <Col className="mb-2">
        <Button size="small" type="link" onClick={() => setShowMore((v) => !v)}>
          {showMore ? t(i18nPrefix + 'collapse') : t(i18nPrefix + 'showMore')}
        </Button>
      </Col>
    </Row>
  )
}
export default ApplicationInfo
