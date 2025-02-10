import { Button, Card, Collapse } from 'antd'
import { useCauseReportContext } from 'src/pro/contexts/CauseReportContext'
import SingleLinkTopology from './SvgTopology'
import TraceTopologyModal from './TraceTopology'
import { useState } from 'react'

const ReportTopology = () => {
  const { isTimeOut, topologyData, category } = useCauseReportContext()
  const [modalVisible, setModalVisible] = useState(false)
  const openModal = (e) => {
    e.stopPropagation()
    setModalVisible(true)
  }
  return (
    <>
      <TraceTopologyModal modalVisible={modalVisible} closeModal={() => setModalVisible(false)} />
      <Collapse
        className="mb-2 bg-[#141414]"
        items={[
          {
            key: 'topology',
            label: isTimeOut ? '故障现象点' : '故障传播链',
            children: (
              <div className=" h-[300px]  overflow-y-hidden">
                <div className="chart_warp">
                  {topologyData ? (
                    <SingleLinkTopology data={topologyData} category={category} />
                  ) : (
                    <></>
                  )}
                </div>
              </div>
            ),
            classNames: { header: 'text-base' },
            styles: { header: { alignItems: 'center' } },
            extra:
              topologyData?.nodes && topologyData?.nodes.length > 0 ? (
                <Button type="link" className="chart_btn" onClick={openModal}>
                  全链路调用
                </Button>
              ) : null,
          },
        ]}
      ></Collapse>
    </>
    // <Card
    //   className="h-full"
    //   title={isTimeOut ? '故障现象点' : '故障传播链'}
    //   classNames={{
    //     body: 'p-0',
    //     header: 'pr-0',
    //   }}
    //   styles={{
    //     body: {
    //       height: 'calc(100% - 56px)',
    //     },
    //   }}
    //   extra={topologyData?.nodes && topologyData?.nodes.length > 0 ? <TraceTopologyModal /> : null}
    // >
    //   {/* <Card.Description className="topology_card_description"> */}
    //   <div className="chart_warp">
    //     {topologyData ? <SingleLinkTopology data={topologyData} category={category} /> : <></>}
    //   </div>
    //   {/* </Card.Description> */}
    // </Card>
  )
}
export default ReportTopology
