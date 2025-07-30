import { Button, Form, Steps } from 'antd'
import React, { useMemo, useState } from 'react'
import { useTranslation } from 'react-i18next'
import { BasicCard } from 'src/core/components/Card/BasicCard'
import ProbeForm from './ProbeForm'
import DataplaneStep from './DataplaneStep'
import { createDataIntegrationApi } from 'src/core/api/integration'
import { useNavigate } from 'react-router-dom'
import InstallCmd from 'src/core/components/InstallCmd'

const ProbeInstallerPage = () => {
  const navigate = useNavigate()
  const { t } = useTranslation('pro/probeInstaller')
  const [current, setCurrent] = useState(0)
  const [readonly, setReadonly] = useState(false)
  const [probeParams, setProbeParams] = useState<any>(null)
  const [form] = Form.useForm()
  const [installStep, setInstallStep] = useState<any>(null) // 安装完成步骤（可选）

  const onSave = (apoConfig?: any) => {
    const params = { ...probeParams }
    if (apoConfig) {
      Object.keys(apoConfig).forEach((key) => {
        params[key] = apoConfig[key]
      })
    }
    createDataIntegrationApi(params).then((res) => {
      setReadonly(true)
      setCurrent(2)
      setInstallStep({
        key: 'third',
        title: '安装完成',
        content: (
          <div className="flex flex-col h-full">
            <div className="flex-1 h-0">
              <InstallCmd
                clusterId={res.id}
                clusterType={res.clusterType}
                apoCollector={probeParams?.apoCollector}
              />
            </div>
            <div className="flex justify-end mt-2 flex-0">
              <Button
                type="primary"
                onClick={() => {
                  navigate('/agent')
                }}
              >
                返回Agent SRE
              </Button>
            </div>
          </div>
        ),
      })
    })
  }

  const onNext = (formParams: any) => {
    setProbeParams(formParams)
    setCurrent(1)
  }

  const items = useMemo(() => {
    const baseSteps = [
      {
        key: 'first',
        title: '探针安装',
        content: <ProbeForm onNext={onNext} form={form} readonly={readonly} />,
      },
      {
        key: 'second',
        title: '接入数据源',
        content: <DataplaneStep probeParams={probeParams} onSave={onSave} readonly={readonly} />,
      },
    ]
    return installStep ? [...baseSteps, installStep] : baseSteps
  }, [readonly, probeParams, installStep])

  return (
    <BasicCard>
      <div className="flex justify-between flex-col h-full">
        <Steps
          current={current}
          items={items}
          className="w-full my-2"
          onChange={(stepIndex) => {
            form.validateFields().then(() => {
              setCurrent(stepIndex)
            })
          }}
        />
        <div className="h-0 flex-1 relative">
          {items.map((item, index) => (
            <div
              key={item.key}
              style={{ display: current === index ? 'block' : 'none' }}
              className="absolute inset-0"
            >
              {item.content}
            </div>
          ))}
        </div>
      </div>
    </BasicCard>
  )
}

export default ProbeInstallerPage
