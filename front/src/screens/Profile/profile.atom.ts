import {atom} from 'recoil'
import decodeJWT from '../../utils/decodeToken.ts'

export interface ICurrentUser {
    user_id: string
}

const profileAtom = atom<ICurrentUser>({
    key: 'current_user',
    default: {
        user_id: decodeJWT(localStorage.getItem('access_token')),
        username: '',
        first_name: '',
        last_name: '',
        date_of_birth: '',
        email: '',
    },
})

export default profileAtom
