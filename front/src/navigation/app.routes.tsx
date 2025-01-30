import {IRoute} from './navigation.types.ts'
import {urls} from './app.urls.ts'
import {lazy} from 'react'

const Main = lazy(() => import('../screens/Main'))
const Profile = lazy(() => import('../screens/Profile'))
const Project = lazy(() => import('../screens/Project'))

const appRoutes: IRoute[] = [
    {
        path: urls.profile,
        element: <Profile/>,
    },
    {
        path: urls.main,
        element: <Main/>
    },
    {
        path: urls.project,
        element: <Project/>

    }
]

export default appRoutes
