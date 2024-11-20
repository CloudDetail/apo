import { post, get } from "../utils/request"

const loginApi = (params) => {
    return post(`/api/user/login?username=${params.username}&password=${params.password}`)
}

const logoutApi = (params) => {
    return post(`/api/user/logout?accessToken=${params.accessToken}&refreshToken=${params.refreshToken}`)
}

const updateEmailApi = (params) => {
    return post(`/api/user/update/email?email=${params.email}`)
}

const updateCorporationApi = (params) => {
    return post(`/api/user/update/info?corporation=${params.corporation}`)
}

const updatePasswordApi = (params) => {
    return post(`/api/user/update/password?oldPassword=${params.oldPassword}&newPassword=${params.newPassword}`)
}

const updatePhoneApi = (params) => {
    return post(`/api/user/update/phone?phone=${params.phone}`)
}

const createUserApi = (params) => {
    return post(`/api/user/create?username=${params.username}&password=${params.password}&confirmPassword=${params.confirmPassword}`)
}

const getUserInfoApi = () => {
    return get(`api/user/info`)
}

const getUserListApi = (params) => {
    return get(`/api/user/list?currentPage=${params.currentPage}&pageSize=${params.pageSize}&username=${params.username}&role=${params.role}&corporation=${params.corporation}`)
}

const removeUserApi = (params) => {
    return post(`/api/user/remove?username=${params.username}`)
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
    removeUserApi
}
