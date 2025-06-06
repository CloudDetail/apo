/**
 * Copyright 2024 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */

import { Image, Empty, Modal } from 'antd'
import React, { useEffect, useState } from 'react'
import { QuestionCircleOutlined, EyeOutlined } from '@ant-design/icons'
import commingSoon from 'src/core/assets/images/commingSoon.svg'
import i18n from 'i18next'
import { useSelector } from 'react-redux'

const CoachMask = React.memo(() => {
  const [visible, setVisible] = useState(false)
  const [images, setImages] = useState({})
  const [list, setList] = useState([])
  const { theme } = useSelector((state) => state.settingReducer)

  const imageModules = import.meta.glob('src/core/assets/snapshot/**/*.png', { eager: true })

  const getImagePath = (imageName, language) => {
    const path = `/src/core/assets/snapshot/${theme}/${language}/${imageName}.png`
    const module = imageModules[path]
    console.log('Available modules:', imageModules, path)
    // console.log('Target path:', path)

    if (!module?.default) {
      console.log(`Image not found: ${path}`)
      return ''
    }
    return module.default
  }

  const generateContent = (language) => {
    const images = {
      alert: getImagePath('alert', language),
      entry: getImagePath('entry', language),
      dashboard: getImagePath('dashboard', language),
      exception: getImagePath('exception', language),
      instance: getImagePath('instance', language),
      k8s: getImagePath('k8s', language),
      logs: getImagePath('logs', language),
      polaris: getImagePath('polaris', language),
      trace: getImagePath('trace', language),
      cpu: getImagePath('cpu', language),
    }

    const list = [
      {
        name: i18n.t('core/mask:descriptions.alertInfo'),
        scene: i18n.t('core/mask:scenes.alertAnalysis'),
        img: [images.alert],
      },
      {
        name: i18n.t('core/mask:descriptions.entryImpact'),
        scene: i18n.t('core/mask:scenes.impactAnalysis'),
        img: [images.entry],
      },
      {
        name: i18n.t('core/mask:descriptions.cascadeAlert'),
        scene: i18n.t('core/mask:scenes.cascadeAlertAnalysis'),
      },
      {
        name: i18n.t('core/mask:descriptions.instanceMetrics'),
        scene: i18n.t('core/mask:scenes.saturationAnalysis'),
        img: [images.instance, images.cpu],
      },
      {
        name: i18n.t('core/mask:descriptions.networkMetrics'),
        scene: i18n.t('core/mask:scenes.networkQualityAnalysis'),
        img: [images.dashboard],
      },
      {
        name: i18n.t('core/mask:descriptions.errorLogs'),
        scene: i18n.t('core/mask:scenes.errorClosedLoop'),
        img: [images.exception],
      },
      {
        name: i18n.t('core/mask:descriptions.polarisMetrics'),
        scene: i18n.t('core/mask:scenes.latencyClosedLoop'),
        img: [images.polaris],
      },
      {
        name: i18n.t('core/mask:descriptions.logs'),
        scene: i18n.t('core/mask:scenes.faultEvidence'),
        img: [images.logs],
      },
      {
        name: i18n.t('core/mask:descriptions.trace'),
        scene: i18n.t('core/mask:scenes.faultEvidence'),
        img: [images.trace],
      },
      {
        name: i18n.t('core/mask:descriptions.k8sEvents'),
        scene: i18n.t('core/mask:scenes.environmentImpact'),
        img: [images.k8s],
      },
    ]
    setImages(images)
    setList(list)
  }
  useEffect(() => {
    console.log(i18n.language)
    generateContent(i18n.language)
  }, [i18n.language, theme])

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
        title={i18n.t('core/mask:coachMaskTitle')}
        open={visible}
        width="100vw"
        onCancel={() => setVisible(false)}
        onOk={() => setVisible(false)}
        destroyOnClose
        centered
        okText={i18n.t('core/mask:closeGuide')}
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
                              <EyeOutlined />{' '}
                              <div className="pl-2">{i18n.t('core/mask:clickToEnlarge')}</div>{' '}
                            </div>
                          ),
                        }}
                      />
                    </div>
                  ))
                ) : (
                  <Empty
                    image={commingSoon}
                    description={i18n.t('core/mask:comingSoon')}
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
})
export default CoachMask
