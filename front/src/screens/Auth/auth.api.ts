import axios from "axios";
import config from "../../config.ts";
import {IAuthResponse, ILoginValues, IRegisterValues} from "./auth.types.ts";
import {handleHttpError, handleHttpResponse} from "../../services/http";
import {getCookie} from "../../utils/getCoockie.ts";

const login = (data: ILoginValues) => {
    return axios
        .post(config.API_URL + '/auth/login', data, {
            withCredentials: true,
        })
        .then(handleHttpResponse)
        .catch(handleHttpError)
}

const register = (data: IRegisterValues) => {
    return axios
        .post(config.API_URL + '/auth/register', data, {
            withCredentials: true,
        })
        .then(handleHttpResponse)
        .catch(handleHttpError)
}

const refresh = (): Promise<Pick<IAuthResponse, 'token_pair'>> => {
    const refreshToken = getCookie('refresh_token');
    return axios
        .post<IAuthResponse>(config.API_URL + '/auth/refresh', {
            refresh_token: refreshToken
        }, {
            headers: {
                Authorization: localStorage.getItem('access_token'),
            },
            withCredentials: true,
        })
        .then((r) => ({
            token_pair: r.data.token_pair,
        }))
        .catch(() => ({
            token_pair: {
                access_token: '',
                refresh_token: '',
            },
        }))
}


const authApi = {
    login,
    register,
    refresh,
}

export default authApi