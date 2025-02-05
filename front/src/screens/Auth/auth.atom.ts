import { IAuthResponse } from './auth.types.ts'
import { atom } from 'recoil'
import Cookies from 'js-cookie';

const AuthAtom = atom<Omit<IAuthResponse, 'user'>>({
  key: 'auth',
  default: {
    token_pair: {
      access_token: localStorage.getItem('access_token') || '',
      refresh_token: Cookies.get('refresh_token') || '',
    },
  },
})
export default AuthAtom
