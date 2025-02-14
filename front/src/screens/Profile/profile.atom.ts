import { atom } from 'recoil'
import { IUser } from './profile.types'

export const initialProfileState = {
  id: '',
  username: '',
  first_name: '',
  last_name: '',
  date_of_birth: '',
  email: '',
}

const profileAtom = atom<IUser>({
  key: 'current_user',
  default: initialProfileState,
})

export default profileAtom
