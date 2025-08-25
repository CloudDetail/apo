import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { checkDataplaneDataApi } from '../api/dataplane'
import { ModelOption } from '../components/AGUIChat/types'
import { Provider } from '../components/AGUIChat/types'
import { getModelListApi, setDefaultModelApi } from '../api/modelSetting'
import { Model } from '../components/AGUIChat/types'

interface AIPermissionState {
    hasPermission: boolean | null
    isLoading: boolean
    error: string | null
    agentReady: boolean | null
    dataPlaneExist: boolean | null
    allModels: ModelOption[]
    originalProviders: Provider[]
    defaultModel: ModelOption | null
}

interface AIPermissionActions {
    checkPermission: () => Promise<void>
    refreshPermission: () => Promise<void>
    refreshModelList: () => Promise<void>
    changeDefaultModel: ({ provider, model }: { provider: string; model: string }) => void
}

export const useAIPermission = (): AIPermissionState & AIPermissionActions => {
    const [hasPermission, setHasPermission] = useState<boolean | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [agentReady, setAgentReady] = useState<boolean | null>(null)
    const [dataPlaneExist, setDataPlaneExist] = useState<boolean | null>(null)
    const { t } = useTranslation('pro/chatAI')
    const [allModels, setAllModels] = useState<ModelOption[]>([])
    const [originalProviders, setOriginalProviders] = useState<Provider[]>([])
    const [defaultModel, setDefaultModel] = useState<ModelOption | null>(null)
    // 统一的刷新函数：拉取 providers + default + 组装 allModels
    const fetchModelList = useCallback(async () => {
        const res = await getModelListApi()
        const providers: Provider[] = res?.data?.providers || []
        setOriginalProviders(providers)

        // 可能后端的 default 不是完整的 ModelOption，这里做一次兜底组装
        const defaultFromApi = res?.data?.default || null
        setDefaultModel(defaultFromApi)

        const models: ModelOption[] = []
        providers.forEach((provider: Provider) => {
            if (provider.name === 'openai-api-compatible') {
                provider.models.forEach((model: Model) => {
                    models.push({
                        provider: provider.name,
                        model: model.name,
                        label: `${model.name}`,
                        value: `${model.name}`,
                    })
                })
            } else {
                if (provider.setUp) {
                    provider.models.forEach((model: Model) => {
                        models.push({
                            provider: provider.name,
                            model: model.name,
                            label: `${model.name}`,
                            value: `${model.name}`,
                        })
                    })
                }
            }
        })
        setAllModels(models)
    }, [])
    // 检查AI权限的函数
    const checkPermission = useCallback(async () => {
        setIsLoading(true)
        setError(null)

        try {
            // 检查数据平面状态
            const dataplaneData = await checkDataplaneDataApi()
            const { agentReady: ready, dataPlaneExist: exist } = dataplaneData.data || dataplaneData
            setAgentReady(ready)
            setDataPlaneExist(exist)

            // 根据数据平面状态决定权限
            if (!ready) {
                // agentReady为false，提示未安装minimal
                setError(t('agentNotReady', 'Agent未就绪，请先安装minimal'))
                setHasPermission(false)
            } else if (!exist) {
                // 只有dataPlaneExist为false，可以正常使用agent，仅提示
                setHasPermission(true)
                setError(t('dataPlaneNotExist', '数据平面不存在，但可以正常使用Agent'))
            } else {
                // 都正常
                setHasPermission(true)
            }
        } catch (err) {
            console.error('数据平面状态检查失败:', err)
            setError(t('dataplaneCheckError', '数据平面状态检查失败，请稍后重试'))
            setHasPermission(false)
        } finally {
            setIsLoading(false)
        }
    }, [t])
    const changeDefaultModel = ({ provider, model }: { provider: string; model: string }) => {
        setDefaultModelApi({ provider, model }).then(() => {
            fetchModelList()
        })
    }
    // 刷新权限检查
    const refreshPermission = useCallback(async () => {
        await checkPermission()
    }, [checkPermission])

    const refreshModelList = useCallback(async () => {
        await fetchModelList()
    }, [fetchModelList])
    // 组件挂载时自动检查权限
    useEffect(() => {
        checkPermission()
        fetchModelList()
    }, [checkPermission, fetchModelList])

    return {
        hasPermission,
        isLoading,
        error,
        agentReady,
        dataPlaneExist,
        allModels,
        originalProviders,
        defaultModel,
        checkPermission,
        refreshPermission,
        refreshModelList,
        changeDefaultModel
    }
}

export default useAIPermission 