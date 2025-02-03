import { RouterProvider } from 'react-router'
import appRouter from './navigation/app.router.tsx'
import { urls } from './navigation/app.urls.ts'
import { useEffect } from 'react'
import AuthProvider from './screens/Auth/components/AuthProvider'

const App = () => {
  useEffect(() => {
    if (location.pathname === '/') {
      location.replace(urls.main)
    }
  }, [])

  return (
    <AuthProvider>
      <RouterProvider router={appRouter} />
    </AuthProvider>
  )
}

export default App
