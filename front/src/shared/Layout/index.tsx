import { Suspense } from 'react'
import LoaderSpinner from '../Loaders/LoderSpinner/index'
import styles from './layout.module.css'
import NavBar from './NavBar'
import { Outlet } from 'react-router'
import Exit from "../../assets/icons/exitIcon.svg?react"
import authApi from '../../screens/Auth/auth.api'
import { useResetRecoilState } from 'recoil'
import profileAtom from '../../screens/Profile/profile.atom'

const Layout = () => {
  const resetProfile = useResetRecoilState(profileAtom)
  const resetAuthState = () =>{
    resetProfile()
  }
  return (
    <div className={styles.wrapper}>
      <div className={styles.left_bar}>
        <NavBar/>
        <button onClick={() => authApi.logout().then(()=> resetAuthState())} className={styles.exit_btn}>
          <Exit/>
        </button>
      </div>
      <div className={styles.page_content}></div>
      <Suspense fallback={<LoaderSpinner />}>
        <Outlet />
      </Suspense>
    </div>
  )
}

export default Layout
