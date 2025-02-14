import { NavLink } from 'react-router'
import { ILink } from '../../../navigation/navigation.types'
interface IProps {
  link: ILink
}

const NavigationLink = (props: IProps) => {
  return <NavLink to={props.link.path}>{props.link?.icon}</NavLink>
}

export default NavigationLink
