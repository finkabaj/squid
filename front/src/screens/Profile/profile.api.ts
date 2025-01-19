import http, { handleHttpError, handleHttpResponse } from '../../services/http'
import { IUserPatchData } from './profile.types.ts'

const getCurrentUserId = () => {
    return http.get('/me').then(handleHttpResponse).catch(handleHttpError)
}

const getUserData = (id: string) => {
    return http.get(`/user/${id}`).then(handleHttpResponse).catch(handleHttpError)
}

const patchUserData = (id: string, data: IUserPatchData) => {
    return http.patch(`/update/${id}`, data).then(handleHttpResponse).catch(handleHttpError)
}

const profileApi = {
    getMyId: getCurrentUserId,
    getUser: getUserData,
    patchUser: patchUserData,
}

export default profileApi
