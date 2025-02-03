import http, { handleHttpError, handleHttpResponse } from '../../services/http'
import { IUserPatchData, IUserPatchPassword } from './profile.types.ts'

const getUserData = (id: string) => {
  return http.get(`/auth/user/${id}`).then(handleHttpResponse).catch(handleHttpError)
}

const patchUserData = (id: string, data: IUserPatchData) => {
  return http.patch(`/auth/user/${id}`, data).then(handleHttpResponse).catch(handleHttpError)
}
const pathUserPassword = (id: string, data: IUserPatchPassword) => {
  return http.patch(`/auth/password/${id}`, data).then(handleHttpResponse).catch(handleHttpError)
}
const profileApi = {
  getUser: getUserData,
  patchUser: patchUserData,
  pathUserPassword: pathUserPassword,
}

export default profileApi
