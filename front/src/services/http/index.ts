import axios, { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import config from '../../config.ts'
import { IRefreshResponse, IUser } from '../../screens/Auth/auth.types.ts'
import { Dispatch, SetStateAction, MutableRefObject } from 'react'
import { IHTTPErrorResponse, IHTTPSuccessResponse } from './http.types.ts'
import authApi from '../../screens/Auth/auth.api.ts'
import Cookies from 'js-cookie';


const http = axios.create({
  baseURL: config.API_URL,
  withCredentials: true,
})

let interceptorsApplied: boolean = false

export const applyInterceptors = (profileState: MutableRefObject<IUser>, setProfileState: Dispatch<SetStateAction<IUser>>) => {
  if (interceptorsApplied) {
    return
  }
  interceptorsApplied = true
  let isRefreshing = false
  let refreshRequest = Promise.resolve({
    result: {
      id: profileState.current.id,
      username: profileState.current.username,
      email: profileState.current.email,
      first_name: profileState.current.first_name,
      last_name: profileState.current.last_name,
      date_of_birth: profileState.current.date_of_birth
    },
  })

  const ensureAuthorization = (): Promise<Pick<IRefreshResponse,'result'>> => {
    const access_token = Cookies.get('access_token')
    const shouldRefresh = profileState.current.id === '' && access_token === ''
    return shouldRefresh ? refreshToken() : Promise.resolve(profileState.current)
  }
  const refreshToken = async (): Promise<Pick<IRefreshResponse, 'result'>>=> {
    if (isRefreshing) return refreshRequest
    isRefreshing = true

    refreshRequest = authApi.refresh().finally(() => (isRefreshing = false))
    return refreshRequest
  }

  http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
    return ensureAuthorization().then(({ result}) => {
      setProfileState(result)

      return config
    })
  })

  http.interceptors.response.use(
    (response) => response,
    (err) => {
      const shouldLogout = err.response && err.response.status === 401

      if (shouldLogout) {
        setProfileState({
          id: '',
          username: '',
          date_of_birth: '',
          first_name: '',
          last_name: '',
          email: ''
        })
      }

      throw err
    }
  )
}

export const handleHttpResponse = <T>(response: AxiosResponse<T>): IHTTPSuccessResponse<T> => {
  return {
    status: 'success',
    body: response.data,
  }
}

export const handleHttpError = (error: AxiosError): IHTTPErrorResponse => {
  return {
    status: 'error',
    message: error?.message ?? '',
    code: error?.response?.status,
  }
}
export default http
