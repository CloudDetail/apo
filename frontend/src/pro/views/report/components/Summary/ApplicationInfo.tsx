import { Row, Col } from 'antd'
import { AiOutlineNodeIndex } from 'react-icons/ai'
import { LuServer, LuHash, LuBox, LuContainer } from 'react-icons/lu'
import { TagsParams } from '../../type'
import { t } from 'i18next'

const i18nPrefix = 'pro/report:'
const InfoItemIconMap = {
  serviceName: <LuServer size={24} className="text-emerald-400" />,
  pod: <LuBox size={24} className="text-lime-400" />,
  containerId: <LuContainer size={24} className="text-lime-400" />,
  nodeName: <AiOutlineNodeIndex size={24} className="text-cyan-400" />,
  pid: <LuHash size={24} className="text-violet-400" />,
}
const InfoItem = ({ type, value }) => {
  return (
    <div className="flex">
      <div className="rounded-md bg-[var(--ant-color-bg-layout)] p-2 ">{InfoItemIconMap[type]}</div>
      <div className="ml-4 flex flex-col justify-between">
        <div className="text-[var(--ant-color-text-secondary)] text-xs">{t(i18nPrefix + type)}</div>
        <div style={{ color: 'var(--ant-color-text-base)' }}>{value}</div>
      </div>
    </div>
  )
}
const TagKeyOrder: (keyof TagsParams)[] = ['serviceName', 'pod', 'containerId', 'nodeName', 'pid']
const ApplicationInfo = ({ tags }: { tags: TagsParams }) => {
  return (
    <Row className="w-full my-2" justify="space-between">
      {TagKeyOrder.filter((key) => tags[key] !== undefined).map((key) => (
        <Col key={key} flex={1}>
          <InfoItem type={key} value={tags[key]} />
        </Col>
      ))}
    </Row>
  )
}
export default ApplicationInfo
