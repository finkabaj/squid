import axios, { AxiosError, AxiosResponse, InternalAxiosRequestConfig } from 'axios'
import config from '../../config.ts'
import { IRefreshResponse} from '../../screens/Auth/auth.types.ts'
import {IUser} from '../../screens/Profile/profile.types.ts'
import { Dispatch, SetStateAction, MutableRefObject } from 'react'
import { IHTTPErrorResponse, IHTTPSuccessResponse } from './http.types.ts'
import authApi from '../../screens/Auth/auth.api.ts'
import { initialProfileState } from '../../screens/Profile/profile.atom.ts'


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
    const shouldRefresh = profileState.current.id === ''
    return shouldRefresh ? refreshToken() : Promise.resolve({ result: profileState.current })
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
        setProfileState(initialProfileState)
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
