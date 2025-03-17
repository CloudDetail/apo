import { post } from "../utils/request"


export const workflowLoginApi = (params) =>{
    return post('/dify/console/api/login',params)
}