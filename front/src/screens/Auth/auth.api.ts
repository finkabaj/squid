import axios from 'axios'
import config from '../../config.ts'
import { ILoginValues, IRefreshResponse, IRegisterValues } from './auth.types.ts'
import { handleHttpError, handleHttpResponse } from '../../services/http'
import { IUser } from '../Profile/profile.types.ts'
import { initialProfileState } from '../Profile/profile.atom.ts'

const login = (data: ILoginValues) => {
  return axios
    .post(config.API_URL + '/auth/login', data, { withCredentials: true })
    .then(handleHttpResponse)
    .catch(handleHttpError)
}

const register = (data: IRegisterValues) => {
  return axios
    .post(config.API_URL + '/auth/register', data, { withCredentials: true })
    .then(handleHttpResponse)
    .catch(handleHttpError)
}

const logout = () => {
  return axios
    .post(config.API_URL + '/auth/logout', {}, { withCredentials: true })
    .then(handleHttpResponse)
    .catch(handleHttpError)
}

const refresh = (): Promise<IRefreshResponse> => {
  return axios
    .post<IUser>(
      config.API_URL + '/auth/refresh',
      {},
      {
        withCredentials: true,
      }
    )
    .then((r) => ({
      status: 'success',
      result: r.data,
    }))
    .catch(() => ({
      status: 'error',
      result: initialProfileState,
    }))
}

const authApi = {
  login,
  register,
  refresh,
  logout,
}

export default authApi
