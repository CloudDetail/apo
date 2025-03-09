import { Col, Radio, Row } from 'antd'
import { useState } from 'react'
import { Trans, useTranslation } from 'react-i18next'
import { useChatAIContext } from 'src/pro/contexts/ChatAIContext'
interface NodeData {
  service: string
  endpoint: string
}
interface CauseConfirmStepProps {
  nextStep: any
}
const i18nKey = 'causeConfirm'
const CauseConfirmStep = ({ nextStep }: CauseConfirmStepProps) => {
  const { t } = useTranslation('pro/chatAI')
  const causeNode = useChatAIContext((ctx) => ctx.causeNode)
  const suspectedNodeList = useChatAIContext((ctx) => ctx.suspectedNodeList)
  const selectedSuspectedNode = useChatAIContext((ctx) => ctx.selectedSuspectedNode)
  const setSelectedSuspectedNode = useChatAIContext((ctx) => ctx.setSelectedSuspectedNode)
  const onSelectNode = (node: NodeData) => {
    setSelectedSuspectedNode(node)
    nextStep()
  }
  const NodeCom = ({ node }) => {
    return (
      <div
        className=" relative w-[200px] m-2 min-h-[30px] p-2 rounded-md border-2 border-solid border-[#6293ff] text-[#6293ff] overflow-hidden cursor-pointer transition-transform duration-300 hover:scale-110"
        style={{ backgroundColor: 'rgba(19, 25, 32, 0.6)' }}
        onClick={() => onSelectNode(node)}
      >
        <div className="text-center text-sm pt-2 px-2 break-all">
          {node.service}
          <div className="text-xs text-[#9e9e9e] text-left  break-all">{node.endpoint}</div>
        </div>
        <Radio
          checked={selectedSuspectedNode?.service === node.service}
          className="absolute right-0 top-0"
        ></Radio>
      </div>
    )
  }
  return (
    <div className="w-full text-center text-xs">
      {t(`${i18nKey}.lastAIDeduction`)}{' '}
      <span className="text-sm font-bold">
        {causeNode?.service}
        <span className="text-xs font-normal">（{causeNode?.endpoint}）</span>
      </span>
      ,
      <Trans
        t={t}
        i18nKey="causeConfirm.notify"
        components={{
          span: <span className="font-extrabold"></span>,
        }}
      />
      <div className="flex justify-around p-4 flex-wrap">
        <Row justify="space-between">
          {suspectedNodeList?.map((node) => (
            <Col>
              <NodeCom node={node} />
            </Col>
          ))}
        </Row>
        {/* {suspectedNodeList?.length <= 2 ? (
          <>
            {suspectedNodeList.map((node) => (
              <NodeCom node={node} />
            ))}
          </>
        ) : (
          <>
            <Row>
              {suspectedNodeList.map((node) => (
                <Col>
                  <NodeCom node={node} />
                </Col>
              ))}
            </Row>
          </>
        )} */}
      </div>
    </div>
  )
}
export default CauseConfirmStep
