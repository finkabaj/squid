import { ILink } from './navigation.types.ts'
import { urls } from './app.urls.ts'

export const appLinks: ILink[] = [
    {
        path: urls.main,
        name: 'Главная',
    },
    {
        path: urls.profile,
        name: 'Профиль',
    },
    {
        path: urls.project,
        name: 'Проект',
    },
]
