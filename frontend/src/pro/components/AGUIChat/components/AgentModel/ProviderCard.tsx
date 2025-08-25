import { Button, Card, Collapse, Form, Input, Tabs, Tag } from 'antd'
import { useState, useEffect, useRef } from 'react'
import { useTranslation } from 'react-i18next'
import { addModelInfoApi, deleteModelInfoApi, updateModelInfoApi } from 'src/pro/api/modelSetting'
import { IoSettingsOutline } from 'react-icons/io5'
import loaderStyles from '../DataStatus.module.scss'
import styles from './index.module.scss'
import { notify } from 'src/core/utils/notify'
import { Provider } from './types'
import { ProviderIcon } from './constants'
import IconWrapper from './IconWrapper'

interface ProviderCardProps {
  provider: Provider
  onRefresh?: () => void
}

const ProviderCard: React.FC<ProviderCardProps> = ({ provider, onRefresh }) => {
  const { t } = useTranslation('pro/agent')
  const [showConfig, setShowConfig] = useState(false)

  const DefaultConfig = ({ defaultConfig, add }: { defaultConfig?: any; add?: boolean }) => {
    const [initialValue, setInitialValue] = useState<any>(null)
    const [loading, setLoading] = useState(false)
    const [form] = Form.useForm()

    useEffect(() => {
      if (defaultConfig) {
        const newValue = { ...defaultConfig, model: defaultConfig.name, provider: provider.name }
        setInitialValue(newValue)
        // 动态设置表单值
        form.setFieldsValue(newValue)
      } else {
        const newValue = provider.config.reduce(
          (acc, item) => {
            acc[item.name] = item.value
            return acc
          },
          { provider: provider.name },
        )
        setInitialValue(newValue)
        // 动态设置表单值
        form.setFieldsValue(newValue)
      }
    }, [defaultConfig])

    const handleSubmit = async () => {
      try {
        setLoading(true)
        const values = await form.validateFields()

        if ((provider.setUp || defaultConfig) && !add) {
          await updateModelInfoApi(values)
          notify({
            message: t('modelSettings.updateSuccess'),
            type: 'success',
          })
        } else {
          await addModelInfoApi(values)
          notify({
            message: t('modelSettings.addSuccess'),
            type: 'success',
          })
        }

        onRefresh?.()
      } catch (error) {
        console.error('保存失败:', error)
        notify({
          message: t('modelSettings.saveError'),
          type: 'error',
        })
      } finally {
        setLoading(false)
      }
    }

    return (
      <div className="px-4 ">
        <Form layout="vertical" size="small" initialValues={initialValue} form={form}>
          <Form.Item name="provider" hidden>
            <Input />
          </Form.Item>
          {provider.config?.map((item) => (
            <Form.Item
              key={item.name}
              name={item.name}
              label={
                <div className="flex items-center gap-2">
                  {item.label}
                  <span className="text-[11px] text-[var(--ant-color-text-tertiary)]">
                    {item.placeholder}
                  </span>
                </div>
              }
              rules={[{ required: item.required }]}
            >
              <Input />
            </Form.Item>
          ))}
        </Form>
        <div className="w-full text-right">
          <Button
            type="primary"
            size="small"
            className="ml-2"
            onClick={handleSubmit}
            loading={loading}
          >
            {t('modelSettings.submit')}
          </Button>
        </div>
      </div>
    )
  }

  const CompatibleConfig = () => {
    const [activeKey, setActiveKey] = useState(provider?.models[0]?.name || 'newTab0')
    const [items, setItems] = useState(() => {
      if (provider.models.length === 0) {
        // 如果 models 为空数组，默认创建一个添加新模型的 tab
        return [
          {
            key: 'newTab0',
            label: t('modelSettings.addNewModel'),
            children: <DefaultConfig add={true} />,
          },
        ]
      }
      return provider.models.map((model) => ({
        key: model.name,
        label: model.name,
        children: <DefaultConfig defaultConfig={model} />,
      }))
    })
    const newTabIndex = useRef(provider.models.length === 0 ? 1 : 0)

    const onEdit = (
      targetKey: React.MouseEvent | React.KeyboardEvent | string,
      action: 'add' | 'remove',
    ) => {
      if (action === 'add') {
        add()
      } else {
        remove(targetKey)
      }
    }

    const add = () => {
      const newActiveKey = `newTab${newTabIndex.current++}`

      setItems((prev) => {
        return [
          ...prev,
          {
            label: t('modelSettings.addNewModel'),
            children: <DefaultConfig add={true} />,
            key: newActiveKey,
          },
        ]
      })
      setActiveKey(newActiveKey)
    }

    const remove = (targetKey) => {
      let newActiveKey = activeKey
      let lastIndex = -1
      items.forEach((item, i) => {
        if (item.key === targetKey) {
          lastIndex = i - 1
        }
      })
      const newPanes = items.filter((item) => item.key !== targetKey)
      if (newPanes.length && newActiveKey === targetKey) {
        if (lastIndex >= 0) {
          newActiveKey = newPanes[lastIndex].key
        } else {
          newActiveKey = newPanes[0].key
        }
      }

      // 检查是否是新加的 tab（key 以 'newTab' 开头）
      const isNewTab = targetKey.startsWith('newTab')

      if (isNewTab) {
        // 如果是新加的 tab，直接删除，不调用 API
        setItems(newPanes)
        setActiveKey(newActiveKey)
      } else {
        // 如果是已存在的模型，调用删除 API
        deleteModelInfoApi({ provider: provider.name, model: targetKey }).then(() => {
          notify({
            message: t('modelSettings.deleteSuccess'),
            type: 'success',
          })
          setItems(newPanes)
          setActiveKey(newActiveKey)
        })
      }
    }
    return (
      <div className="w-full">
        <Tabs
          size="small"
          type="editable-card"
          onChange={(newActiveKey) => setActiveKey(newActiveKey)}
          activeKey={activeKey}
          onEdit={onEdit}
          items={items}
        />
      </div>
    )
  }

  return (
    <Card
      type="inner"
      className="w-full"
      key={provider.name}
      classNames={{
        actions: styles.providerCardActions,
      }}
      actions={[
        <Collapse
          key="collapse"
          size="small"
          ghost
          className="m-0"
          activeKey={showConfig ? provider.name : ''}
          onChange={(key) => setShowConfig(key.includes(provider.name))}
          items={[
            {
              key: provider.name,
              label: (
                <span className="text-xs font-bold text-[var(--ant-color-text-tertiary)]">
                  {t('modelSettings.showConfigDetail')}
                </span>
              ),
              children:
                provider.name === 'openai-api-compatible' ? (
                  <CompatibleConfig />
                ) : (
                  <DefaultConfig />
                ),
            },
          ]}
        />,
      ]}
    >
      <div className="flex items-center  justify-between mb-2">
        <div className="flex items-center gap-2">
          <IconWrapper
            icon={
              <img src={ProviderIcon[provider.name]} alt={provider.name} width={32} height={32} />
            }
            size="32px"
          />
          <div className="flex flex-col">
            <span className="font-bold text-base">{provider.name}</span>
            <span className="text-xs text-[var(--ant-color-text-secondary)]">
              {provider.description}
            </span>
          </div>
        </div>
        <div
          className="flex items-center gap-2 flex-col cursor-pointer"
          onClick={() => setShowConfig(true)}
        >
          <div className="text-xs text-[var(--ant-color-text-tertiary)]">
            <div className={loaderStyles.loaderContainer}>
              <span className="text-xs">API-KEY</span>
              {provider.setUp ? (
                <div className={loaderStyles.loader}></div>
              ) : (
                <div className={loaderStyles.errorLoader}></div>
              )}
              <IoSettingsOutline className="text-sm text-[var(--ant-color-text-base)]" />
            </div>
          </div>
        </div>
      </div>
      {provider.models.map((model) => (
        <Tag bordered={false} key={model.name}>
          {model.name}
        </Tag>
      ))}
    </Card>
  )
}

export default ProviderCard
