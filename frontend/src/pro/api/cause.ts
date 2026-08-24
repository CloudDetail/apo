import { get, post } from 'src/core/utils/request'

export function getMutatedtracesApi(traceId: string, spanId: string) {
  return get(`/originx/api/query/mutatedtraceswithspan/${traceId}/${spanId}`)
}
// 可以带url参数mutatedType,指定要分析的根因类型;
// /flows/span/run/{traceId:string}/{spanId:string}?mutatedType=
export function getFlowsApi(traceId: string, spanId: string, mutatedType?: string, groupId?: string) {
  return get(`/originx/api/flows/span/run/${traceId}/${spanId}`, { mutatedType, groupId })
}
export function getErrorTraceLogApi(traceId: string, spanId: string) {
  return get(`/originx/api/query/logs/${traceId}/${spanId}`)
}

//错误报告
export function getErrortracesApi(traceId: string) {
  return get(`/originx/api/query/errortraces/${traceId}`)
}
export function getErrorFlowsApi(traceId: string, spanId: string) {
  return get(`/originx/api/flows/error/${traceId}/${spanId}`)
}

export function getEntryEndpointApi(params: any) {
  return post(`/originx/api/query/entryendpoint`, params)
}
