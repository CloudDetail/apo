import {
  Form,
  Segmented,
  Input,
  Typography,
  Collapse,
  Button,
  Checkbox,
  Alert,
  FormInstance,
} from 'antd'
import MetricsFormItem from 'src/core/components/Integration/MetricsFormItem'
import TraceFormItem from 'src/core/components/Integration/TraceFormItem'
import LogsFormItem from 'src/core/components/Integration/LogsFormItem'
import { useTranslation } from 'react-i18next'
import APOCollectorFormItem from 'src/core/components/Integration/APOCollectorFormItem'
import { portsDefault } from 'src/core/views/IntegrationCenter/constant'
import { useState, useCallback, useMemo } from 'react'

// 类型定义
type ConfigType = 'trace' | 'metric' | 'log'

interface ProbeFormProps {
  onNext: (formParams: any) => void
  form: FormInstance<FormValues>
  readonly: boolean
}

interface FormValues {
  id?: string
  name: string
  clusterType: 'k8s' | 'vm'
  apoCollector: {
    ports?: Record<string, any>
    collectorAddr: string
  }
  metric: {
    dsType: string
    mode: string
    name: string
  }
  log: {
    dbType: string
    mode: string
    name: string
  }
}

const ProbeForm = ({ onNext, form, readonly = false }: ProbeFormProps) => {
  const { t } = useTranslation('core/dataIntegration')
  const [apoConfig, setApoConfig] = useState<ConfigType[]>([])

  // 集群类型选项
  const clusterTypeOptions = useMemo(
    () => [
      {
        label: 'Kubernetes',
        value: 'k8s',
      },
      {
        label: t('vm'),
        value: 'vm',
      },
    ],
    [t],
  )

  // 数据源配置项
  const dataSourceConfigs = useMemo(
    () => [
      {
        key: 'trace' as const,
        label: t('traceIntegration'),
        component: TraceFormItem,
      },
      {
        key: 'metric' as const,
        label: t('metricsIntegration'),
        component: MetricsFormItem,
      },
      {
        key: 'log' as const,
        label: t('logsIntegration'),
        component: LogsFormItem,
      },
    ],
    [t],
  )

  // 处理配置切换
  const handleConfigToggle = useCallback((type: ConfigType) => {
    setApoConfig((prev) =>
      prev.includes(type) ? prev.filter((item) => item !== type) : [...prev, type],
    )
  }, [])

  // 处理表单提交
  const handleSubmit = useCallback(async () => {
    try {
      const values = await form.validateFields()
      const params = { ...values }

      // 设置默认端口配置
      if (!values.apoCollector.ports) {
        params.apoCollector.ports = portsDefault?.reduce((map, item) => {
          map[item.key] = item.value
          return map
        }, {})
      }

      dataSourceConfigs.map(({ key, component: Component }) => {
        if (!apoConfig.includes(key)) {
          delete params[key]
        }
      })

      onNext(params)
      console.log('Form submitted:', params)
    } catch (errorInfo: any) {
      // 滚动到第一个错误字段
      if (errorInfo.errorFields?.length > 0) {
        form.scrollToField(errorInfo.errorFields[0].name[0])
      }
    }
  }, [form, onNext, apoConfig])

  // Collapse 额外内容组件
  const CollapseExtra = useCallback(
    ({ type }: { type: ConfigType }) => (
      <Checkbox
        checked={apoConfig.includes(type)}
        onClick={(e) => {
          e.stopPropagation()
          handleConfigToggle(type)
        }}
      >
        配置
      </Checkbox>
    ),
    [apoConfig, handleConfigToggle],
  )

  // 表单初始值
  const initialValues = useMemo(
    () => ({
      clusterType: 'k8s' as const,
      metric: {
        dsType: 'self-collector',
        mode: 'pql',
        name: 'APO-DEFAULT-VM',
      },
      log: {
        dbType: 'self-collector',
        mode: 'sql',
        name: 'APO-DEFAULT-CH',
      },
      apoCollector: {
        collectorAddr: 'apo-nginx-proxy-svc',
      },
    }),
    [],
  )

  return (
    <div className="h-full flex flex-col">
      <Form
        className="h-full overflow-auto mx-5 my-2"
        scrollToFirstError
        form={form}
        layout="vertical"
        validateMessages={{
          required: '${label}' + t('required'),
        }}
        initialValues={initialValues}
        disabled={readonly}
      >
        <Form.Item name="id" hidden>
          <Input />
        </Form.Item>

        <Form.Item name="name" label={t('clusterName')} rules={[{ required: true }]}>
          <Input />
        </Form.Item>

        <Form.Item name="clusterType" label={t('clusterType')} rules={[{ required: true }]}>
          <Segmented options={clusterTypeOptions} />
        </Form.Item>

        <APOCollectorFormItem />
        <Alert
          className="mb-2"
          showIcon
          message="默认仅采集核心关键数据，您可以通过配置允许探针采集更多数据，增强智能分析效果"
          type="info"
        />
        <Typography.Title level={5}>
          {t('可选数据采集配置')}
          <span className="text-xs ml-2 text-[var(--ant-color-text-secondary)]">
            {t('(推荐在没有其他可观测性数据源时使用）')}
          </span>
        </Typography.Title>

        {/* 数据源配置折叠面板 */}
        {dataSourceConfigs.map(({ key, label, component: Component }) => (
          <Collapse
            key={key}
            className="mb-2"
            items={[
              {
                forceRender: true,
                key,
                label,
                children: <Component />,
                extra: <CollapseExtra type={key} />,
              },
            ]}
          />
        ))}
      </Form>
      {!readonly && (
        <div className="flex justify-end mt-2">
          <Button type="primary" onClick={handleSubmit}>
            {t('下一步')}
          </Button>
        </div>
      )}
    </div>
  )
}

export default ProbeForm
