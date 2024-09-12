import { Image, Empty, Modal } from 'antd'
import React, { useEffect, useRef, useState } from 'react'
import AlertPng from 'src/assets/snapshot/alert.png'
import EntryPng from 'src/assets/snapshot/entry.png'
import DashboardImg from 'src/assets/snapshot/dashboard.jpg'
import ExceptionPng from 'src/assets/snapshot/exception.png'
import InstancePng from 'src/assets/snapshot/instance.png'
import K8sPng from 'src/assets/snapshot/k8s.png'
import LogsPng from 'src/assets/snapshot/logs.png'
import PolarisPng from 'src/assets/snapshot/polaris.png'
import TracePng from 'src/assets/snapshot/trace.png'
import CommingSoon from 'src/assets/images/commingSoon.svg'
import CpuPng from 'src/assets/snapshot/cpu.png'
import { QuestionCircleOutlined, EyeOutlined } from '@ant-design/icons'
export default function CoachMask() {
  const [visible, setVisible] = useState(false)
  const list = [
    {
      name: '接口自身的告警信息、应用层告警和资源层告警',
      scene: '告警分析',
      img: [AlertPng],
    },
    {
      name: '接口的影响业务入口黄金指标',
      scene: '影响面分析',
      img: [EntryPng],
    },
    {
      name: '接口的下游依赖告警关联',
      scene: '级联告警影响分析',
    },
    //实例
    {
      name: '接口的实例和节点的资源指标',
      scene: '饱和度分析',
      img: [InstancePng, CpuPng],
    },
    //大盘rtt
    {
      name: '接口的网络指标',
      scene: '网络质量分析',
      img: [DashboardImg],
    },
    //错误实例
    {
      name: '接口的代码Exception，以及含有Exception的日志',
      scene: '错误闭环',
      img: [ExceptionPng],
    },
    //北极性
    {
      name: '接口执行的北极星指标',
      scene: '延时闭环',
      img: [PolarisPng],
    },
    //日志
    {
      name: '接口执行的日志',
      scene: '故障佐证',
      img: [LogsPng],
    },
    //trace
    {
      name: '接口执行的Trace',
      scene: '故障佐证',
      img: [TracePng],
    },
    //🉑k8s
    {
      name: '接口所依赖的容器环境关键事件',
      scene: '环境影响',
      img: [K8sPng],
    },
  ]
  const shouldShowPopup = () => {
    const hasShown = localStorage.getItem('CoachMaskShown')

    if (hasShown) {
      const parsedData = JSON.parse(hasShown)

      // 检查当前时间是否超过过期时间
      if (new Date().getTime() > parsedData.expires) {
        localStorage.removeItem('popupShown')
        return true // 弹窗过期，应该重新显示
      }

      return false // 弹窗已经显示过且未过期，不需要再显示
    }

    return true // 未找到标记，应该显示弹窗
  }
  const setPopupShown = () => {
    const expirationDate = new Date()
    expirationDate.setMonth(expirationDate.getMonth() + 1) // 设置过期时间为一个月后

    const popupData = {
      shown: true,
      expires: expirationDate.getTime(),
    }

    localStorage.setItem('CoachMaskShown', JSON.stringify(popupData))
    setVisible(true)
  }
  useEffect(() => {
    const visible = shouldShowPopup()
    if (visible) {
      setPopupShown()
    }
  })
  return (
    <>
      <QuestionCircleOutlined className="text-lg text-[#6261cc] px-3" onClick={setPopupShown} />
      <Modal
        title={'服务详情指南'}
        open={visible}
        // footer={null}
        // style={{ width: '100vw', height: '100vh' }}
        // bodyStyle={{
        //   height: 'calc(100vh - 125px)',
        //   overflowY: 'auto',
        // }}
        width="100vw"
        onCancel={() => setVisible(false)}
        onOk={() => setVisible(false)}
        destroyOnClose
        centered
        okText={'关闭指南'}
        footer={(_, { OkBtn }) => (
          <>
            <OkBtn />
          </>
        )}
        maskClosable={false}
      >
        <div className="h-[700px] overflow-y-scroll">
          {list.map((item, index) => (
            <div className="flex w-full justify-center " key={index}>
              <div className="w-[400px] text-left p-1">
                <span className="text-[#46A5F7] font-bold text-xl">{item.scene}</span>
                <div className="w-[500px] text-base">{item.name}</div>
              </div>

              <div className="flex-shrink-0 flex justify-center w-[800px] h-[100px] overflow-hidden relative ">
                {item.img ? (
                  item.img.map((src) => (
                    <div className="flex-1 " key={src}>
                      <Image
                        src={src}
                        width={'auto'}
                        height={'auto'}
                        preview={{
                          closeIcon: (
                            <div className="w-full fixed left-0 flex items-center justify-center top-0 bg-slate-600 p-3">
                              <div className="p-1">
                                <span className="text-[#46A5F7] font-bold text-xl pr-5">
                                  {item.scene}
                                </span>
                                <span className="text-base">{item.name}</span>
                              </div>
                            </div>
                          ),
                          mask: (
                            <div className="flex absolute top-12">
                              <EyeOutlined /> <div className="pl-2">点击放大</div>{' '}
                            </div>
                          ),
                        }}
                        // preview={{
                        //   toolbarRender: (_, { image: { url }, transform: { scale } }) => (
                        //     <div className="text-left p-1 flex items-center">
                        //       <span className="text-[#46A5F7] font-bold text-xl">{item.scene}</span>
                        //       <div className="w-[500px] text-base">{item.name}</div>
                        //     </div>
                        //   ),
                        // }}
                      />
                    </div>
                  ))
                ) : (
                  <Empty image={CommingSoon} description="敬请期待" imageStyle={{ height: 70 }} />
                )}
              </div>
            </div>
          ))}
        </div>
      </Modal>
      {/* <div
        className="fixed w-full h-full top-0 left-0 bg-[#000000] bg-opacity-70 flex items-center justify-center"
        style={{ zIndex: 1000 }}
      >
        <div className="bg-black p-3 rounded">
          {list.map((item, index) => (
            <div className="flex w-full mt-6 items-center justify-center ">
              <div className="flex-shrink-0 flex justify-end">
                <div className="w-[20px] bg-[#66bb6a] h-[20px] rounded-full mr-10"></div>
              </div>
              <div className="w-[700px] text-left flex justify-between items-center text-sm">
                <div className="w-[500px]">{item.name}</div>
                <span className="text-[#46A5F7] font-bold">{item.scene}</span>
              </div>
            </div>
          ))}
        </div>
      </div> */}
    </>
  )
}
