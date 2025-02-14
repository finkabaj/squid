import { appLinks } from '../../../navigation/app.links'
import NavigationLink from '../NavLink'
import styles from './NavBar.module.css'
const NavBar = () => {
  return (
    <div className={styles.wrapper}>
      {appLinks.map((link) => (
        <NavigationLink key={link.path} link={link} />
      ))}
    </div>
  )
}
export default NavBar
