import { post } from "src/core/utils/request";

export function getReportApi(params) {
    return post('api/alerts/events/report/get', params)
}