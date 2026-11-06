import { useState } from 'react'
import type { FormEvent } from 'react'
import logo from './assets/nextgate-logo.png'
import OrbitGraphic from './OrbitGraphic'

type LoginPageProps = {
  onAuthenticated: (token: string) => void
}

type LoginResponse = {
  accessToken?: string
  error?: string
}

const errorMessages: Record<string, string> = {
  invalid_request: 'Enter your administrator password.',
  password_is_required: 'Enter your administrator password.',
  invalid_credentials: 'That password is not valid.',
  keyring_admin_error:
    'The administrator password could not be read from the system keyring.',
  database_not_connected: 'The admin database is not available.',
  jwt_error: 'The administrator session could not be created.',
}

function errorMessage(code?: string) {
  return (
    (code ? errorMessages[code] : undefined) ??
    'Sign-in failed. Check the server logs and try again.'
  )
}

function LoginPage({ onAuthenticated }: LoginPageProps) {
  const [password, setPassword] = useState('')
  const [showPassword, setShowPassword] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [error, setError] = useState('')

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (!password) {
      setError('Enter your administrator password.')
      return
    }

    setIsSubmitting(true)
    setError('')

    try {
      const response = await fetch('/api/v1/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ password }),
      })

      let payload: LoginResponse = {}

      try {
        payload = (await response.json()) as LoginResponse
      } catch {
        setError('The admin API returned an unreadable response.')
        return
      }

      if (!response.ok) {
        setError(errorMessage(payload.error))
        return
      }

      if (!payload.accessToken) {
        setError('The admin API returned an incomplete session.')
        return
      }

      onAuthenticated(payload.accessToken)
    } catch {
      setError('The admin API is unreachable. Check that NextGate is running.')
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <div className="gate-page">
      <OrbitGraphic />

      <header className="site-header">
        <img className="brand-logo" src={logo} alt="NextGate" />
        <div className="service-state">
          <span className="service-dot" />
          Admin service available
        </div>
      </header>

      <main className="access-layout">
        <section className="access-copy" aria-labelledby="page-title">
          <h1 id="page-title">
            Administrator
            <span>access.</span>
          </h1>
          <p>
            Manage provider credentials, models, and local access from one
            controlled surface.
          </p>
        </section>

        <section className="auth-stage" aria-labelledby="sign-in-title">
          <div className="sign-in-state">
            <div className="form-heading">
              <h2 id="sign-in-title">Sign in</h2>
              <span>60-minute session</span>
            </div>
            <p className="form-intro">
              Use the administrator password configured on this host.
            </p>

            <form className="login-form" onSubmit={handleSubmit} noValidate>
              <div className="field-group">
                <div className="field-label-row">
                  <label htmlFor="admin-password">Administrator password</label>
                  <button
                    type="button"
                    className="text-button"
                    aria-label={
                      showPassword ? 'Hide password' : 'Show password'
                    }
                    aria-pressed={showPassword}
                    onClick={() => setShowPassword((visible) => !visible)}
                  >
                    {showPassword ? 'Hide' : 'Show'}
                  </button>
                </div>
                <input
                  id="admin-password"
                  name="password"
                  type={showPassword ? 'text' : 'password'}
                  value={password}
                  onChange={(event) => setPassword(event.target.value)}
                  autoComplete="current-password"
                  autoFocus
                  required
                  aria-invalid={error ? true : undefined}
                  aria-describedby={error ? 'login-error' : undefined}
                  placeholder="Enter password"
                />
              </div>

              {error ? (
                <div className="form-error" id="login-error" role="alert">
                  <span>{error}</span>
                </div>
              ) : null}

              <button
                type="submit"
                className="primary-button"
                disabled={isSubmitting || password.length === 0}
              >
                {isSubmitting ? (
                  <>
                    <span className="button-spinner" aria-hidden="true" />
                    Checking password
                  </>
                ) : (
                  'Sign in'
                )}
              </button>
            </form>

            <p className="form-note">
              The password is verified against the local OS keyring. Access
              tokens expire after 60 minutes.
            </p>
          </div>
        </section>
      </main>

      <footer className="page-meta">
        <div>
          <span>Endpoint</span>
          <strong>127.0.0.1:9090</strong>
        </div>
        <div>
          <span>Session</span>
          <strong>60 minutes</strong>
        </div>
        <div>
          <span>Storage</span>
          <strong>Local keyring</strong>
        </div>
      </footer>
    </div>
  )
}

export default LoginPage
