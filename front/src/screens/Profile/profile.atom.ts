import { atom } from 'recoil'
import decodeJWT from '../../utils/decodeToken.ts'

const getUserId = () => {
  const tokenItem = localStorage.getItem('access_token')
  if (!tokenItem) {
    return ''
  }
  const user_id = decodeJWT(tokenItem)
  if (!user_id) {
    return ''
  }
  return user_id
}

const profileAtom = atom({
  key: 'current_user',
  default: {
    user_id: getUserId(),
    username: '',
    first_name: '',
    last_name: '',
    date_of_birth: '',
    email: '',
  },
})

export default profileAtom
