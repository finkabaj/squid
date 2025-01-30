import { createBrowserRouter } from 'react-router'
import Layout from '../shared/Layout'
import appRoutes from './app.routes.tsx'

const appRouter = createBrowserRouter([
    {
        element: <Layout />,
        children: appRoutes,
    },
])
export default appRouter
