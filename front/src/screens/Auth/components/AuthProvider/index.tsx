import { PropsWithChildren, useEffect, useLayoutEffect } from 'react'
import useHttpLoader from '../../../../shared/hooks/httpLoader/useHttpLoader.ts'
import { useRecoilState } from 'recoil'
import authAtom from '../../auth.atom.ts'
import { applyInterceptors } from '../../../../services/http'
import authApi from '../../auth.api.ts'
import LoaderPage from '../../../../shared/Loaders/LoaderPage'
import Auth from '../../index.tsx'
import profileApi from '../../../Profile/profile.api.ts'
import profileAtom from '../../../Profile/profile.atom.ts'

const AuthProvider = (props: PropsWithChildren) => {
  const { wait, loading } = useHttpLoader()

  const [authState, setAuthState] = useRecoilState(authAtom)
  const [profileState, setProfileState] = useRecoilState(profileAtom)

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
            refresh_token: resp.token_pair.refresh_token,
          },
        }))
      }
    }).then(() => {
      profileApi.getUser(profileState.user_id).then((resp) => {
        setProfileState((prevState) => ({
          ...prevState,
          ...resp.body,
        }))
      })
    })
  }, [])
  if (loading) {
    return <LoaderPage label={`Кажется, вы тут уже были...\nПытаемся авторизоваться...`} />
  }

  if (!authState.token_pair.access_token || authState.token_pair.access_token === '') {
    return <Auth />
  }

  return <>{props.children}</>
}

export default AuthProvider
