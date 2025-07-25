import { useState, useEffect, useCallback } from 'react'
import { useTranslation } from 'react-i18next'
import { checkDataplaneDataApi } from '../api/dataplane'

interface AIPermissionState {
    hasPermission: boolean | null
    isLoading: boolean
    error: string | null
    agentReady: boolean | null
    dataPlaneExist: boolean | null
}

interface AIPermissionActions {
    checkPermission: () => Promise<void>
    refreshPermission: () => Promise<void>
}

export const useAIPermission = (): AIPermissionState & AIPermissionActions => {
    const [hasPermission, setHasPermission] = useState<boolean | null>(null)
    const [isLoading, setIsLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [agentReady, setAgentReady] = useState<boolean | null>(null)
    const [dataPlaneExist, setDataPlaneExist] = useState<boolean | null>(null)
    const { t } = useTranslation('pro/chatAI')

    // 检查AI权限的函数
    const checkPermission = useCallback(async () => {
        setIsLoading(true)
        setError(null)

        try {
            // 检查数据平面状态
            const dataplaneData = await checkDataplaneDataApi()
            const { agentReady: ready, dataPlaneExist: exist } = dataplaneData.data || dataplaneData
            // const ready = false
            // const exist = false
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

    // 刷新权限检查
    const refreshPermission = useCallback(async () => {
        await checkPermission()
    }, [checkPermission])

    // 组件挂载时自动检查权限
    useEffect(() => {
        checkPermission()
    }, [checkPermission])

    return {
        hasPermission,
        isLoading,
        error,
        agentReady,
        dataPlaneExist,
        checkPermission,
        refreshPermission,
    }
}

export default useAIPermission 