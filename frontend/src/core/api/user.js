import { postFormData, get } from "../utils/request"

const loginApi = (params) => {
    return postFormData(`/api/user/login`,params)
}

const logoutApi = (params) => {
    return postFormData(`/api/user/logout`,params)
}

const updateEmailApi = (params) => {
    return postFormData(`/api/user/update/email`,params)
}

const updateCorporationApi = (params) => {
    return postFormData(`/api/user/update/info`,params)
}

const updatePasswordApi = (params) => {
    return postFormData(`/api/user/update/password`,params)
}

const updatePhoneApi = (params) => {
    return postFormData(`/api/user/update/phone`,params)
}

const createUserApi = (params) => {
    return postFormData(`/api/user/create`,params)
}

const getUserInfoApi = () => {
    return get(`api/user/info`)
}

const getUserListApi = (params,signal) => {
    return get(`/api/user/list`,params,{signal})
}

const removeUserApi = (params) => {
    return postFormData(`/api/user/remove`,params)
}

const updatePasswordWithNoOldPwd = (params) => {
    return postFormData(`/api/user/reset`,params)
}

export {
    loginApi,
    logoutApi,
    updateEmailApi,
    updateCorporationApi,
    updatePasswordApi,
    updatePhoneApi,
    createUserApi,
    getUserInfoApi,
    getUserListApi,
    removeUserApi,
    updatePasswordWithNoOldPwd
}