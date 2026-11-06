import { Fragment, useCallback, useEffect, useState } from 'react'
import logo from './assets/nextgate-logo.png'
import {
  adminRequest,
  apiErrorMessage,
  isUnauthorized,
  type AddModelInput,
  type AddProviderInput,
  type Provider,
} from './dashboard/api'
import {
  AddModelDialog,
  AddProviderDialog,
  ConfirmDialog,
} from './dashboard/Dialogs'
import ModelsPanel from './dashboard/ModelsPanel'

type DashboardProps = {
  accessToken: string
  onSignOut: () => void
}

type ProvidersResponse = {
  providers?: Provider[] | null
}

type LoadState = 'loading' | 'ready' | 'error'

type Notice = {
  kind: 'success' | 'error'
  message: string
}

const providerDateFormatter = new Intl.DateTimeFormat('en', {
  dateStyle: 'medium',
})

function formatProviderDate(value: string) {
  const date = new Date(value)
  return Number.isNaN(date.getTime())
    ? 'Unknown'
    : providerDateFormatter.format(date)
}

function providerMark(name: string) {
  return name.slice(0, 2).toUpperCase()
}

function ProvidersIcon() {
  return (
    <svg
      className="sidebar-icon"
      viewBox="0 0 20 20"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.5"
      aria-hidden="true"
    >
      <rect x="3" y="3" width="5.5" height="5.5" rx="1" />
      <rect x="11.5" y="3" width="5.5" height="5.5" rx="1" />
      <rect x="3" y="11.5" width="5.5" height="5.5" rx="1" />
      <rect x="11.5" y="11.5" width="5.5" height="5.5" rx="1" />
    </svg>
  )
}

function RefreshIcon() {
  return (
    <svg
      viewBox="0 0 16 16"
      width="14"
      height="14"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.4"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M13.1 5.3A5.5 5.5 0 1 0 13.4 9" />
      <path d="M13.4 2.7v3.4H10" />
    </svg>
  )
}

function PlusIcon() {
  return (
    <svg
      viewBox="0 0 16 16"
      width="13"
      height="13"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.4"
      strokeLinecap="round"
      aria-hidden="true"
    >
      <path d="M8 2.6v10.8M2.6 8h10.8" />
    </svg>
  )
}

function ChevronIcon({ isOpen }: { isOpen: boolean }) {
  return (
    <svg
      className={`row-chevron${isOpen ? ' is-open' : ''}`}
      viewBox="0 0 16 16"
      width="13"
      height="13"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.3"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="m5.5 6.5 2.5 2.5 2.5-2.5" />
    </svg>
  )
}

function TrashIcon() {
  return (
    <svg
      viewBox="0 0 16 16"
      width="14"
      height="14"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.25"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M2.8 4.3h10.4M6.2 4.3V2.8h3.6v1.5M4.2 4.3l.6 8.3h6.4l.6-8.3M6.5 6.6v3.8M9.5 6.6v3.8" />
    </svg>
  )
}

function SignOutIcon() {
  return (
    <svg
      viewBox="0 0 16 16"
      width="14"
      height="14"
      fill="none"
      stroke="currentColor"
      strokeWidth="1.4"
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
    >
      <path d="M6.3 2.7H3.5a1 1 0 0 0-1 1v8.6a1 1 0 0 0 1 1h2.8" />
      <path d="M9.6 5.3 12.3 8l-2.7 2.7M6 8h6.1" />
    </svg>
  )
}

function Dashboard({ accessToken, onSignOut }: DashboardProps) {
  const [providers, setProviders] = useState<Provider[]>([])
  const [loadState, setLoadState] = useState<LoadState>('loading')
  const [error, setError] = useState('')
  const [refreshKey, setRefreshKey] = useState(0)
  const [expandedProviderId, setExpandedProviderId] = useState<number | null>(
    null,
  )
  const [modelRefreshKey, setModelRefreshKey] = useState(0)
  const [isAddingProvider, setIsAddingProvider] = useState(false)
  const [modelProvider, setModelProvider] = useState<Provider | null>(null)
  const [providerToDelete, setProviderToDelete] = useState<Provider | null>(
    null,
  )
  const [isDeletingProvider, setIsDeletingProvider] = useState(false)
  const [deleteError, setDeleteError] = useState('')
  const [notice, setNotice] = useState<Notice | null>(null)

  const notify = useCallback((kind: Notice['kind'], message: string) => {
    setNotice({ kind, message })
  }, [])

  useEffect(() => {
    let cancelled = false

    async function loadProviders() {
      setLoadState('loading')
      setError('')

      try {
        const payload = await adminRequest<ProvidersResponse>(
          accessToken,
          '/api/v1/providers',
        )

        if (!cancelled) {
          setProviders(payload.providers ?? [])
          setLoadState('ready')
        }
      } catch (requestError) {
        if (isUnauthorized(requestError)) {
          onSignOut()
          return
        }

        if (!cancelled) {
          setLoadState('error')
          setError(
            apiErrorMessage(requestError, 'Providers could not be loaded.'),
          )
        }
      }
    }

    void loadProviders()

    return () => {
      cancelled = true
    }
  }, [accessToken, onSignOut, refreshKey])

  useEffect(() => {
    if (!notice) {
      return
    }

    const timer = window.setTimeout(() => setNotice(null), 4500)
    return () => window.clearTimeout(timer)
  }, [notice])

  async function addProvider(input: AddProviderInput) {
    try {
      await adminRequest<{ message: string }>(accessToken, '/api/v1/providers', {
        method: 'POST',
        body: JSON.stringify(input),
      })

      setIsAddingProvider(false)
      setRefreshKey((key) => key + 1)
      notify('success', `${input.name} added.`)
    } catch (requestError) {
      if (isUnauthorized(requestError)) {
        onSignOut()
      }

      throw requestError
    }
  }

  async function addModel(input: AddModelInput) {
    if (!modelProvider) {
      return
    }

    try {
      await adminRequest<{ message: string }>(
        accessToken,
        `/api/v1/providers/${modelProvider.id}/models`,
        {
          method: 'POST',
          body: JSON.stringify(input),
        },
      )

      const providerName = modelProvider.name
      setModelProvider(null)
      setModelRefreshKey((key) => key + 1)
      notify('success', `${input.name} added to ${providerName}.`)
    } catch (requestError) {
      if (isUnauthorized(requestError)) {
        onSignOut()
      }

      throw requestError
    }
  }

  async function deleteProvider() {
    if (!providerToDelete) {
      return
    }

    setIsDeletingProvider(true)
    setDeleteError('')

    try {
      await adminRequest<{ message: string }>(
        accessToken,
        `/api/v1/providers/${providerToDelete.id}`,
        { method: 'DELETE' },
      )

      if (expandedProviderId === providerToDelete.id) {
        setExpandedProviderId(null)
      }

      notify('success', `${providerToDelete.name} deleted.`)
      setProviderToDelete(null)
      setRefreshKey((key) => key + 1)
    } catch (requestError) {
      if (isUnauthorized(requestError)) {
        onSignOut()
        return
      }

      setDeleteError(
        apiErrorMessage(requestError, 'The provider could not be deleted.'),
      )
    } finally {
      setIsDeletingProvider(false)
    }
  }

  const providerCount = loadState === 'ready' ? providers.length : null

  return (
    <div className="admin-shell">
      <aside className="admin-sidebar">
        <div className="sidebar-brand">
          <img className="sidebar-logo" src={logo} alt="NextGate" />
          <span className="sidebar-edition">Admin</span>
        </div>

        <nav className="sidebar-nav" aria-label="Administrator navigation">
          <p className="sidebar-nav-label">Workspace</p>
          <a
            className="sidebar-link is-active"
            href="#providers"
            aria-current="page"
          >
            <ProvidersIcon />
            <span>Providers</span>
            <span className="sidebar-count">
              {providerCount === null ? '—' : providerCount}
            </span>
          </a>
        </nav>

        <div className="sidebar-foot">
          <div className="sidebar-server">
            <span>Local admin server</span>
            <strong>127.0.0.1:9090</strong>
          </div>

          <div className="sidebar-session">
            <span className="sidebar-live" aria-hidden="true" />
            <div>
              <strong>Administrator</strong>
              <span>JWT session active</span>
            </div>
          </div>

          <button
            type="button"
            className="sidebar-signout"
            onClick={onSignOut}
          >
            <SignOutIcon />
            Sign out
          </button>
        </div>
      </aside>

      <div className="admin-workspace">
        <header className="workspace-bar">
          <p>
            NextGate <span>/</span> Providers
          </p>
          <span className="workspace-local">
            <i aria-hidden="true" />
            Local only
          </span>
        </header>

        <main className="workspace" id="providers">
          <header className="workspace-heading">
            <div>
              <h1>Providers</h1>
              <p>Manage upstream APIs and the models each one exposes.</p>
            </div>

            <div className="heading-actions">
              <button
                type="button"
                className="action-button is-primary"
                onClick={() => setIsAddingProvider(true)}
              >
                <PlusIcon />
                New provider
              </button>
              <button
                type="button"
                className={`refresh-button${loadState === 'loading' ? ' is-loading' : ''}`}
                onClick={() => setRefreshKey((key) => key + 1)}
                disabled={loadState === 'loading'}
              >
                <RefreshIcon />
                Refresh
              </button>
            </div>
          </header>

          <div className="registry-meta" aria-live="polite">
            <span>
              <strong>{providerCount === null ? '—' : providerCount}</strong>{' '}
              configured
            </span>
            <span>API keys stay hidden</span>
          </div>

          {notice ? (
            <div
              className={`action-notice is-${notice.kind}`}
              role={notice.kind === 'error' ? 'alert' : 'status'}
            >
              <span>{notice.message}</span>
              <button
                type="button"
                aria-label="Dismiss message"
                onClick={() => setNotice(null)}
              >
                ×
              </button>
            </div>
          ) : null}

          {loadState === 'loading' && providers.length === 0 ? (
            <div className="registry-loading" role="status">
              <span className="loading-track" aria-hidden="true">
                <i />
              </span>
              Reading the provider registry
            </div>
          ) : null}

          {loadState === 'error' ? (
            <div className="registry-error" role="alert">
              <p>{error}</p>
              <button
                type="button"
                onClick={() => setRefreshKey((key) => key + 1)}
              >
                Try again
              </button>
            </div>
          ) : null}

          {providers.length > 0 ? (
            <div className="registry-table-wrap">
              <table className="registry-table has-actions">
                <thead>
                  <tr>
                    <th scope="col">Provider</th>
                    <th scope="col">Endpoint</th>
                    <th scope="col">Created</th>
                    <th scope="col">
                      <span className="sr-only">Actions</span>
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {providers.map((provider) => {
                    const isExpanded = expandedProviderId === provider.id

                    return (
                      <Fragment key={provider.id}>
                        <tr>
                          <td>
                            <div className="provider-identity">
                              <span
                                className="provider-mark"
                                aria-hidden="true"
                              >
                                {providerMark(provider.name)}
                              </span>
                              <strong>{provider.name}</strong>
                            </div>
                          </td>
                          <td>
                            <span className="provider-endpoint">
                              {provider.base_url}
                            </span>
                          </td>
                          <td>
                            <span className="provider-date">
                              {formatProviderDate(provider.created_at)}
                            </span>
                          </td>
                          <td>
                            <div className="provider-row-actions">
                              <button
                                type="button"
                                className="row-action"
                                aria-expanded={isExpanded}
                                aria-controls={`models-${provider.id}`}
                                onClick={() =>
                                  setExpandedProviderId((current) =>
                                    current === provider.id ? null : provider.id,
                                  )
                                }
                              >
                                Models
                                <ChevronIcon isOpen={isExpanded} />
                              </button>
                              <button
                                type="button"
                                className="icon-button is-danger"
                                aria-label={`Delete ${provider.name}`}
                                title={`Delete ${provider.name}`}
                                onClick={() => {
                                  setDeleteError('')
                                  setProviderToDelete(provider)
                                }}
                              >
                                <TrashIcon />
                              </button>
                            </div>
                          </td>
                        </tr>

                        {isExpanded ? (
                          <tr className="models-row">
                            <td colSpan={4} id={`models-${provider.id}`}>
                              <ModelsPanel
                                accessToken={accessToken}
                                provider={provider}
                                refreshKey={modelRefreshKey}
                                onAddModel={setModelProvider}
                                onNotify={notify}
                                onUnauthorized={onSignOut}
                              />
                            </td>
                          </tr>
                        ) : null}
                      </Fragment>
                    )
                  })}
                </tbody>
              </table>
            </div>
          ) : null}

          {loadState === 'ready' && providers.length === 0 ? (
            <div className="registry-empty">
              <h2>No providers yet.</h2>
              <p>
                Connect the first upstream API to start exposing models through
                NextGate.
              </p>
              <button
                type="button"
                className="empty-action"
                onClick={() => setIsAddingProvider(true)}
              >
                <PlusIcon />
                Add the first provider
              </button>
            </div>
          ) : null}
        </main>
      </div>

      {isAddingProvider ? (
        <AddProviderDialog
          onClose={() => setIsAddingProvider(false)}
          onSubmit={addProvider}
        />
      ) : null}

      {modelProvider ? (
        <AddModelDialog
          providerName={modelProvider.name}
          onClose={() => setModelProvider(null)}
          onSubmit={addModel}
        />
      ) : null}

      {providerToDelete ? (
        <ConfirmDialog
          title={`Delete ${providerToDelete.name}?`}
          description={`This removes the provider and its models from NextGate. The action cannot be undone.`}
          confirmLabel="Delete provider"
          isBusy={isDeletingProvider}
          error={deleteError}
          onClose={() => {
            if (!isDeletingProvider) {
              setProviderToDelete(null)
            }
          }}
          onConfirm={() => void deleteProvider()}
        />
      ) : null}
    </div>
  )
}

export default Dashboard
