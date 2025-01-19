import { PropsWithChildren, useEffect, useLayoutEffect } from 'react'
import useHttpLoader from '../../../../shared/hooks/httpLoader/useHttpLoader.ts'
import { useRecoilState, useSetRecoilState } from 'recoil'
import authAtom from '../../auth.atom.ts'
import { applyInterceptors } from '../../../../services/http'
import authApi from '../../auth.api.ts'
import LoaderPage from '../../../../shared/'
import Auth from '../../index.tsx'
import profileApi from '../../../Profile/profile.api.ts'
import ProfileAtom from '../../../Profile/profile.atom.ts'

const AuthProvider = (props: PropsWithChildren) => {
    const { wait, loading } = useHttpLoader()

    const [authState, setAuthState] = useRecoilState(authAtom)
    const setProfileState = useSetRecoilState(ProfileAtom)

    useEffect(() => {
        applyInterceptors(authState, setAuthState)
    }, [])

    useLayoutEffect(() => {
        wait(authApi.refresh(), (resp) => {
            if (resp.token_pair.access_token) {
                setAuthState((prevState) => ({
                    ...prevState,
                    token_pair: {
                        access_token: resp.token_pair.access_token,
                        refresh_token: resp.token_pair.refresh_token
                    },
                }))
            }
        }).then(() => {
            profileApi.getMyId().then((resp: any) => {
                setProfileState((prev) => ({ ...prev, user_id: resp.body.result.user_id }))
            })
        })
    }, [])

    if (loading) {
        return (
            <LoaderPage
                inscription={`Кажется, вы тут уже были...\nПытаемся авторизоваться...`}
            />
        )
    }

    if (!authState.token_pair.access_token || authState.token_pair.access_token === '') {
        return <Auth />
    }

    return <>{props.children}</>
}

export default AuthProvider
