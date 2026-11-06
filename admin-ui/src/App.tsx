import { useCallback, useEffect, useState } from 'react'
import Dashboard from './Dashboard'
import LoginPage from './LoginPage'
import './App.css'
import {
  clearAdminToken,
  readAdminToken,
  saveAdminToken,
} from './session'

type View = 'login' | 'dashboard'

function viewFromLocation(): View {
  const path = window.location.pathname.replace(/\/+$/, '')
  return path === '/dashboard' ? 'dashboard' : 'login'
}

function App() {
  const [view, setView] = useState<View>(viewFromLocation)
  const [accessToken, setAccessToken] = useState<string | null>(readAdminToken)
  const activeView: View =
    view === 'dashboard' && accessToken ? 'dashboard' : 'login'

  useEffect(() => {
    function handlePopState() {
      setView(viewFromLocation())
    }

    window.addEventListener('popstate', handlePopState)
    return () => window.removeEventListener('popstate', handlePopState)
  }, [])

  useEffect(() => {
    document.title =
      activeView === 'dashboard'
        ? 'Providers | NextGate Admin'
        : 'NextGate Admin'
  }, [activeView])

  useEffect(() => {
    if (view === 'dashboard' && !accessToken) {
      window.history.replaceState(null, '', '/')
    }
  }, [accessToken, view])

  const handleAuthenticated = useCallback((token: string) => {
    saveAdminToken(token)
    setAccessToken(token)
    window.history.pushState(null, '', '/dashboard')
    setView('dashboard')
  }, [])

  const handleSignOut = useCallback(() => {
    clearAdminToken()
    setAccessToken(null)
    window.history.replaceState(null, '', '/')
    setView('login')
  }, [])

  if (activeView === 'dashboard' && accessToken) {
    return <Dashboard accessToken={accessToken} onSignOut={handleSignOut} />
  }

  return <LoginPage onAuthenticated={handleAuthenticated} />
}

export default App
