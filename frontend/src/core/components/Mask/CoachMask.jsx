import { Image, Empty, Modal } from 'antd'
import React, { useEffect, useState } from 'react'
import AlertPng from 'src/core/assets/snapshot/alert.png'
import EntryPng from 'src/core/assets/snapshot/entry.png'
import DashboardImg from 'src/core/assets/snapshot/dashboard.jpg'
import ExceptionPng from 'src/core/assets/snapshot/exception.png'
import InstancePng from 'src/core/assets/snapshot/instance.png'
import K8sPng from 'src/core/assets/snapshot/k8s.png'
import LogsPng from 'src/core/assets/snapshot/logs.png'
import PolarisPng from 'src/core/assets/snapshot/polaris.png'
import TracePng from 'src/core/assets/snapshot/trace.png'
import CommingSoon from 'src/core/assets/images/commingSoon.svg'
import CpuPng from 'src/core/assets/snapshot/cpu.png'
import { QuestionCircleOutlined, EyeOutlined } from '@ant-design/icons'
import { useTranslation } from 'react-i18next'

export default function CoachMask() {
  const { t } = useTranslation('core/mask')
  const [visible, setVisible] = useState(false)
  const list = [
    {
      name: t('descriptions.alertInfo'),
      scene: t('scenes.alertAnalysis'),
      img: [AlertPng],
    },
    {
      name: t('descriptions.entryImpact'),
      scene: t('scenes.impactAnalysis'),
      img: [EntryPng],
    },
    {
      name: t('descriptions.cascadeAlert'),
      scene: t('scenes.cascadeAlertAnalysis'),
    },
    {
      name: t('descriptions.instanceMetrics'),
      scene: t('scenes.saturationAnalysis'),
      img: [InstancePng, CpuPng],
    },
    {
      name: t('descriptions.networkMetrics'),
      scene: t('scenes.networkQualityAnalysis'),
      img: [DashboardImg],
    },
    {
      name: t('descriptions.errorLogs'),
      scene: t('scenes.errorClosedLoop'),
      img: [ExceptionPng],
    },
    {
      name: t('descriptions.polarisMetrics'),
      scene: t('scenes.latencyClosedLoop'),
      img: [PolarisPng],
    },
    {
      name: t('descriptions.logs'),
      scene: t('scenes.faultEvidence'),
      img: [LogsPng],
    },
    {
      name: t('descriptions.trace'),
      scene: t('scenes.faultEvidence'),
      img: [TracePng],
    },
    {
      name: t('descriptions.k8sEvents'),
      scene: t('scenes.environmentImpact'),
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
  }, [])

  return (
    <>
      <QuestionCircleOutlined className="text-lg text-[#6261cc] px-3" onClick={setPopupShown} />
      <Modal
        title={t('coachMaskTitle')}
        open={visible}
        width="100vw"
        onCancel={() => setVisible(false)}
        onOk={() => setVisible(false)}
        destroyOnClose
        centered
        okText={t('closeGuide')}
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
                <div className="w-[400px] text-base">{item.name}</div>
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
                              <EyeOutlined /> <div className="pl-2">{t('clickToEnlarge')}</div>{' '}
                            </div>
                          ),
                        }}
                      />
                    </div>
                  ))
                ) : (
                  <Empty
                    image={CommingSoon}
                    description={t('comingSoon')}
                    imageStyle={{ height: 70 }}
                  />
                )}
              </div>
            </div>
          ))}
        </div>
      </Modal>
    </>
  )
}
