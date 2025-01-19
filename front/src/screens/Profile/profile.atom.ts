import {atom}from 'recoil'
export interface ICurrentUserID {
    user_id: string
}

const CurrentUserIDAtom = atom<ICurrentUserID>({
    key: 'current_user',
    default: {
        user_id: '',
    },
})

export default CurrentUserIDAtom
