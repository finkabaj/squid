import axios from 'axios'
import config from '../../config.ts'
import { ILoginValues, IRefreshResponse, IRegisterValues, IUser } from './auth.types.ts'
import { handleHttpError, handleHttpResponse } from '../../services/http'
import Cookies from 'js-cookie';

const login = (data: ILoginValues) => {
  return axios
    .post(config.API_URL + '/auth/login', data, {withCredentials: true})
    .then(handleHttpResponse)
    .catch(handleHttpError)
}

const register = (data: IRegisterValues) => {
  return axios
    .post(config.API_URL + '/auth/register', data, {withCredentials: true})
    .then(handleHttpResponse)
    .catch(handleHttpError)
}


const refresh = (): Promise<IRefreshResponse> => {
  const refreshToken = Cookies.get('refresh_token')
  return axios
    .post<IUser>(
      config.API_URL + '/auth/refresh',
      {
        refresh_token: refreshToken,
      },
      {
        withCredentials: true
      }
    )
    .then((r) => ({
      status: "success",
      result: r.data
    }))
    .catch(() => ({
      status: 'error',
      result: {
        id: '',
        username: '',
        first_name: '',
        last_name: '',
        date_of_birth: '',
        email: '',
      }
      
    }))
}

const authApi = {
  login,
  register,
  refresh,
}

export default authApi
