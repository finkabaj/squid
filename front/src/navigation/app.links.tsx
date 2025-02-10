import { ILink } from './navigation.types.ts'
import { urls } from './app.urls.ts'
import ProfileIcon from "../assets/icons/ProfileIcon.svg?react"
import HomeIcon from "../assets/icons/HomeIcon.svg?react"

export const appLinks: ILink[] = [
  {
    path: urls.main,
    name: 'Главная',
    icon: <HomeIcon/>
  },
  {
    path: urls.profile,
    name: 'Профиль',
    icon: <ProfileIcon/>
  },
  {
    path: urls.project,
    name: 'Проект',
  },
]
