import {IAuthResponse} from './auth.types.ts'
import {atom} from 'recoil'
import {getCookie} from "../../utils/getCoockie.ts";

const AuthAtom = atom<Omit<IAuthResponse, 'user'>>({
    key: 'auth',
    default: {
        token_pair: {
            access_token: localStorage.getItem('access_token') || '',
            refresh_token: getCookie('refresh_token') || '',

        }
    }
})
export default AuthAtom
