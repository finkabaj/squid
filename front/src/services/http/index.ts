import axios, {AxiosError, AxiosResponse, InternalAxiosRequestConfig} from "axios";
import config from "../../config.ts";
import {IAuthResponse} from "../../screens/Auth/auth.types.ts";
import {Dispatch, SetStateAction} from "react";
import {IHTTPErrorResponse, IHTTPSuccessResponse} from "./http.types.ts";
import authApi from "../../screens/Auth/auth.api.ts";

const http = axios.create({
    baseURL: config.API_URL,
    withCredentials: true,
    headers: {
        'Access-Control-Allow-Origin': '*',
        'Access-Control-Allow-Methods': 'GET, POST, PUT, DELETE, OPTIONS',
        'Access-Control-Allow-Headers': 'Content-Type, Authorization',
    },

})

let interceptorsApplied: boolean = false;

export const applyInterceptors = (
    authState: Omit<IAuthResponse, 'user'>,
    setAuthState: Dispatch<SetStateAction<Omit<IAuthResponse, 'user'>>>,
) => {
    if (interceptorsApplied) {
        return
    }
    interceptorsApplied = true
    let isRefreshing = false
    let refreshRequest = Promise.resolve({
        token_pair: {
            access_token: authState.token_pair.access_token,
            refresh_token: authState.token_pair.refresh_token,
        },
    })

    const ensureAuthorization = (): Promise<Omit<IAuthResponse, 'user'>> => {
        const shouldRefresh =
            authState.token_pair.access_token === ''
        return shouldRefresh ? refreshToken() : Promise.resolve(authState)
    }
    const refreshToken = async (): Promise<Omit<IAuthResponse, 'user'>> => {
        if (isRefreshing) return refreshRequest
        isRefreshing = true

        refreshRequest = authApi.refresh().finally(() => (isRefreshing = false))
        return refreshRequest
    }

    http.interceptors.request.use((config: InternalAxiosRequestConfig) => {
        return ensureAuthorization().then(({token_pair}) => {
            setAuthState((prevAuthState) => ({
                ...prevAuthState,
                token_pair: token_pair
            }))

            config.headers.Authorization = `${token_pair.access_token}`
            return config
        })
    })


    http.interceptors.response.use(
        (response) => response,
        (err) => {
            const shouldLogout = err.response && err.response.status === 401

            if (shouldLogout) {
                localStorage.removeItem('access_token')
                setAuthState({
                    user: null,
                    token_pair: {access_token: '', refresh_token: ''},

                })
            }

            throw err
        }
    )
}

export const handleHttpResponse = <T>(
    response: AxiosResponse<T>
): IHTTPSuccessResponse<T> => {
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