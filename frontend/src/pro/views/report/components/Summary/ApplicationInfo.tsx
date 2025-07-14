import { Row, Col, Button, Tooltip } from 'antd'
import { AiOutlineNodeIndex } from 'react-icons/ai'
import { LuServer, LuHash, LuBox, LuContainer } from 'react-icons/lu'
import { BsChevronLeft } from 'react-icons/bs'
import { MdOutlineMore } from 'react-icons/md'
import { TagsParams } from '../../type'
import { t } from 'i18next'
import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'

const i18nPrefix = 'pro/report:'
const InfoItemIconMap = {
  service: <LuServer size={24} className="text-emerald-400" />,
  pod: <LuBox size={24} className="text-lime-400" />,
  containerId: <LuContainer size={24} className="text-lime-400" />,
  nodeName: <AiOutlineNodeIndex size={24} className="text-cyan-400" />,
  pid: <LuHash size={24} className="text-violet-400" />,
}
const InfoItem = ({ type, value }: { type: string; value: string }) => {
  const isServiceName = type === 'service'
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
const TagKeyOrder: (keyof TagsParams)[] = ['service', 'pod', 'containerId', 'nodeName', 'pid']
const ApplicationInfo = ({ tags }: { tags: TagsParams }) => {
  const hasService =
    tags.service !== undefined &&
    tags.service !== null &&
    tags.service !== '' &&
    Object.keys(tags.service || {}).length > 0
  const [showMore, setShowMore] = useState(!hasService)
  const [showInitialTooltip, setShowInitialTooltip] = useState(true)

  useEffect(() => {
    if (showInitialTooltip) {
      const timer = setTimeout(() => {
        setShowInitialTooltip(false)
      }, 3000)
      return () => clearTimeout(timer)
    }
  }, [showInitialTooltip])

  return (
    <Row className="w-full mt-2" align="middle">
      {TagKeyOrder.filter(
        (key) =>
          key === 'service' &&
          tags[key] !== undefined &&
          tags[key] !== null &&
          tags[key] !== '' &&
          Object.keys(tags[key] || {}).length > 0,
      ).map((key) => (
        <Col key={key} flex={showMore ? 1 : 0}>
          <InfoItem type={key} value={tags[key]} />
        </Col>
      ))}

      {showMore &&
        TagKeyOrder.filter(
          (key) =>
            key !== 'service' &&
            tags[key] !== undefined &&
            tags[key] !== null &&
            tags[key] !== '' &&
            Object.keys(tags[key] || {}).length > 0,
        ).map((key) => (
          <Col key={key} flex={1}>
            <InfoItem type={key} value={tags[key]} />
          </Col>
        ))}
      {hasService && (
        <Col>
          <Tooltip
            title={showMore ? t(i18nPrefix + 'collapse') : t(i18nPrefix + 'showMore')}
            placement="right"
            trigger={['hover', 'click']}
            open={showInitialTooltip && !showMore ? true : undefined}
          >
            <Button
              type="link"
              size="small"
              onClick={() => {
                setShowMore((v) => !v)
              }}
            >
              {showMore ? <BsChevronLeft size={32} /> : <MdOutlineMore size={32} />}
            </Button>
          </Tooltip>
        </Col>
      )}
    </Row>
  )
}
export default ApplicationInfo
