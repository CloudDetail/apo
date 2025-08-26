import { Divider, Modal, Select } from 'antd'
import { useState, useMemo } from 'react'
import { useTranslation } from 'react-i18next'
import { IoSettingsOutline, IoWarning } from 'react-icons/io5'
import styles from './index.module.scss'
import { ProviderIcon } from './constants'
import ProviderCard from './ProviderCard'
import { useAGUIPermission } from 'src/pro/contexts/AGUIContext'

const AgentModel = () => {
  const { t } = useTranslation('pro/agent')
  const { allModels, originalProviders, defaultModel, changeDefaultModel, refreshModelList } =
    useAGUIPermission()
  const [open, setOpen] = useState(false)

  // 分组：已配置 / 未配置
  const { configuredProviders, unconfiguredProviders } = useMemo(() => {
    const configured = originalProviders.filter((p) => p.setUp)
    const unconfigured = originalProviders.filter((p) => !p.setUp)
    return { configuredProviders: configured, unconfiguredProviders: unconfigured }
  }, [originalProviders])
  return (
    <div className="cursor-pointer">
      {defaultModel?.model ? (
        <Select
          size="small"
          variant="filled"
          options={allModels}
          // 修正：value 传字符串（模型名），与 options.value 对齐
          value={defaultModel}
          className={styles.modelSelect}
          onSelect={(value, option) => {
            changeDefaultModel(option)
          }}
          popupMatchSelectWidth={false}
          labelRender={() => {
            console.log(defaultModel)
            const { model } = defaultModel
            return (
              <div className="flex items-center gap-2">
                <img
                  src={ProviderIcon[defaultModel.provider]}
                  alt={defaultModel.provider}
                  width={14}
                  height={14}
                />

                <span className="text-xs text-[var(--ant-color-text-tertiary)]">{model}</span>
              </div>
            )
          }}
          // 可选：自定义 Option 渲染（带供应商图标）
          optionRender={(option) => {
            const data = option.data as any
            const providerName = data?.provider as keyof typeof ProviderIcon
            const icon = ProviderIcon[providerName]
            return (
              <div className="flex items-center gap-2">
                {icon ? <img src={icon} alt={String(providerName)} width={14} height={14} /> : null}
                <span className="text-xs text-[var(--ant-color-text-tertiary)]">
                  {option.label}
                </span>
              </div>
            )
          }}
          dropdownRender={(menu) => {
            return (
              <div className="text-xs">
                {menu}
                {/* <div className="flex items-center gap-2 w-full p-2">
                  <div className="flex items-center gap-2">
                    <img
                      src={ProviderIcon[defaultModel.provider as keyof typeof ProviderIcon]}
                      alt={defaultModel.provider}
                      width={14}
                      height={14}
                    />
                    <span className="text-xs text-[var(--ant-color-text-tertiary)]">
                      {defaultModel.model}
                    </span>
                  </div>
                </div> */}
                <Divider style={{ margin: '8px 0' }} />
                <div
                  className="flex items-center gap-2 w-full p-2 cursor-pointer"
                  onClick={() => setOpen(true)}
                >
                  <IoSettingsOutline className="text-sm text-[var(--ant-color-text-base)]" />
                  {t('modelSettings.settingModels')}
                </div>
              </div>
            )
          }}
        />
      ) : (
        <div className="flex items-center" onClick={() => setOpen(true)}>
          <IoWarning className="text-yellow-500  mr-1" />
          <span className="text-xs text-[var(--ant-color-text-tertiary)]">
            {t('modelSettings.setDefaultModel')}
          </span>
        </div>
      )}

      <Modal
        open={open}
        onCancel={() => setOpen(false)}
        width="60vw"
        destroyOnClose
        centered
        footer={() => <></>}
        maskClosable={true}
        title={t('modelSettings.title')}
        styles={{ body: { height: '80vh', overflowY: 'auto', overflowX: 'hidden' } }}
      >
        <div className="w-full space-y-4">
          {/* 第一部分：已配置 */}
          <Divider orientation="left" plain>
            {t('modelSettings.configured')}
          </Divider>
          <div className="flex items-center gap-2 flex-col w-full">
            {configuredProviders.map((item) => (
              <ProviderCard provider={item} key={item.name} onRefresh={refreshModelList} />
            ))}
            {configuredProviders.length === 0 && (
              <div className="text-xs text-[var(--ant-color-text-tertiary)] py-2">
                {t('modelSettings.noConfiguredProviders')}
              </div>
            )}
          </div>

          {/* 第二部分：未配置 */}
          <Divider orientation="left" plain>
            {t('modelSettings.unconfigured')}
          </Divider>
          <div className="flex items-center gap-2 flex-col w-full">
            {unconfiguredProviders.map((item) => (
              <ProviderCard provider={item} key={item.name} onRefresh={refreshModelList} />
            ))}
            {unconfiguredProviders.length === 0 && (
              <div className="text-xs text-[var(--ant-color-text-tertiary)] py-2">
                {t('modelSettings.noUnconfiguredProviders')}
              </div>
            )}
          </div>
        </div>
      </Modal>
    </div>
  )
}

export default AgentModel
