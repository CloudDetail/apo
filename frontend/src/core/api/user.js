import { post, get } from "../utils/request"

const loginApi = (params) => {
    return post(`/api/user/login`,params,{},true)
}

const logoutApi = (params) => {
    return post(`/api/user/logout`,params,{},true)
}

const updateEmailApi = (params) => {
    return post(`/api/user/update/email`,params,{},true)
}

const updateCorporationApi = (params) => {
    return post(`/api/user/update/info`,params,{},true)
}

const updatePasswordApi = (params) => {
    return post(`/api/user/update/password`,params,{},true)
}

const updatePhoneApi = (params) => {
    return post(`/api/user/update/phone`,params,{},true)
}

const createUserApi = (params) => {
    return post(`/api/user/create`,params,{},true)
}

const getUserInfoApi = () => {
    return get(`api/user/info`)
}

const getUserListApi = (params,signal) => {
    return get(`/api/user/list`,params,{signal})
}

const removeUserApi = (params) => {
    return post(`/api/user/remove`,params,{},true)
}

const updatePasswordWithNoOldPwd = (params) => {
    return post(`/api/user/reset`,params,{},true)
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