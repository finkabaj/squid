import { PropsWithChildren, useEffect, useLayoutEffect, useRef } from 'react'
import useHttpLoader from '../../../../shared/hooks/httpLoader/useHttpLoader.ts'
import { useRecoilState } from 'recoil'
import { applyInterceptors } from '../../../../services/http'
import authApi from '../../auth.api.ts'
import LoaderPage from '../../../../shared/Loaders/LoaderPage'
import Auth from '../../index.tsx'
import profileAtom from '../../../Profile/profile.atom.ts'

const AuthProvider = (props: PropsWithChildren) => {
  const { wait, loading } = useHttpLoader()

  const [profileState, setProfileState] = useRecoilState(profileAtom)
  const refState = useRef({...profileState})

  useEffect(() => {
    applyInterceptors(refState, setProfileState)
  }, [profileState])

  useLayoutEffect(() => {
    wait(authApi.refresh(), (resp) => {
      if (resp.status === 'success') {
        setProfileState(resp.result)
      }
    })
  }, [])


  if (loading) {
    return <LoaderPage label={`Кажется, вы тут уже были...\nПытаемся авторизоваться...`} />
  }

  if (!profileState|| profileState.id === '') {
    return <Auth />
  }

  return <>{props.children}</>
}

export default AuthProvider
