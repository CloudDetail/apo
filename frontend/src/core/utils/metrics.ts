/**
 * Copyright 2025 CloudDetail
 * SPDX-License-Identifier: Apache-2.0
 */
import { t } from "i18next"


export const MetricsValue = (originalValue: number | null, type: 'latency' | 'errorRate' | 'tps' | 'logs') => {
    if (originalValue === null) {
        return
    } else {
        let value: number | string = parseFloat((Math.floor(originalValue * 100) / 100).toString())
        switch (type) {
            case 'latency': {
                let convertValue = Math.floor((originalValue / 1000) * 100) / 100
                if (originalValue > 0 && originalValue < 10) {
                    value = '< 0.01 ms'
                } else {
                    value = parseFloat(convertValue.toString()) + 'ms'
                }

                break
            }
            case 'errorRate':
                if (originalValue > 0 && originalValue < 0.01) {
                    value = '< 0.01%'
                } else {
                    value += '%'
                }
                break
            case 'tps':
                if (originalValue > 0 && originalValue < 0.01) {
                    value = '< 0.01'
                }
                value += t('common:tempCell.times')

                break
            case 'logs':
                if (originalValue > 0 && originalValue < 0.01) {
                    value = '< 0.01'
                }
                value += t('common:tempCell.unitsText')
                break
        }
        return value
    }
}