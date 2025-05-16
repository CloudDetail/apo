import { Button, Collapse } from 'antd'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'
import SingleLinkTopology from './SvgTopology'
import TraceTopologyModal from './TraceTopology'
import { useState } from 'react'
import { useTranslation } from 'react-i18next'
const i18nKey = 'topology'
const ReportTopology = () => {
  const { t } = useTranslation('pro/causeReport')

  const { isTimeOut, topologyData, category } = useCauseReportContext()
  const [modalVisible, setModalVisible] = useState(false)
  const openModal = (e) => {
    e.stopPropagation()
    setModalVisible(true)
  }
  const items = [
    {
      key: 'topology',
      label: isTimeOut ? t(`${i18nKey}.faultPoint`) : t(`${i18nKey}.faultPropagationChain`),
      children: (
        <div className="h-[300px] overflow-y-hidden">
          <div className="chart_warp">
            {topologyData ? <SingleLinkTopology data={topologyData} category={category} /> : <></>}
          </div>
        </div>
      ),
      classNames: { header: 'text-base' },
      styles: { header: { alignItems: 'center' } },
      extra:
        topologyData?.nodes && topologyData?.nodes.length > 0 ? (
          <Button type="link" className="chart_btn" onClick={openModal}>
            {t(`${i18nKey}.fullLinkInvocation`)}
          </Button>
        ) : null,
    },
  ]
  return (
    <>
      <TraceTopologyModal modalVisible={modalVisible} closeModal={() => setModalVisible(false)} />
      <Collapse className="mb-2 bg-[var(--ant-color-bg-container)]" items={items}></Collapse>
    </>
  )
}
export default ReportTopology
