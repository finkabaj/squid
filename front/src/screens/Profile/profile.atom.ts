import { atom } from 'recoil'

const profileAtom = atom({
  key: 'current_user',
  default: {
    id: '',
    username: '',
    first_name: '',
    last_name: '',
    date_of_birth: '',
    email: '',
  },
})

export default profileAtom
