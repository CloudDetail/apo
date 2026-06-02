import { convertTime } from 'src/core/utils/time'

export interface MetricValueFormatterMessages {
  unit: string
  timesPerMinute: string
}

export const formatMetricValue = (
  value: number,
  type: string,
  messages: MetricValueFormatterMessages,
): string => {
  switch (type) {
    case 'logs':
      return Math.floor(value) + messages.unit

    case 'latency':
    case 'p90':
      if (value > 0 && value < 10) {
        return '< 0.01 ms'
      }
      return convertTime(value, 'ms', 2) + 'ms'

    case 'errorRate':
      if (value > 0 && value < 0.01) {
        return '< 0.01%'
      }
      return parseFloat(value.toFixed(2)) + '%'

    case 'tps':
      if (value > 0 && value < 0.01) {
        return '< 0.01' + messages.timesPerMinute
      }
      return parseFloat((Math.floor(value * 100) / 100).toString()) + messages.timesPerMinute

    default:
      return String(value)
  }
}
