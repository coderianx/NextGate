import { useEffect, useState } from 'react'
import {
  adminRequest,
  apiErrorMessage,
  isUnauthorized,
  type Provider,
  type ProviderModel,
} from './api'
import { ConfirmDialog } from './Dialogs'

type LoadState = 'loading' | 'ready' | 'error'

type ModelsPanelProps = {
  accessToken: string
  provider: Provider
  refreshKey: number
  onAddModel: (provider: Provider) => void
  onNotify: (kind: 'success' | 'error', message: string) => void
  onUnauthorized: () => void
}

type ModelsResponse = {
  models?: ProviderModel[] | null
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

function ModelsPanel({
  accessToken,
  provider,
  refreshKey,
  onAddModel,
  onNotify,
  onUnauthorized,
}: ModelsPanelProps) {
  const [models, setModels] = useState<ProviderModel[]>([])
  const [loadState, setLoadState] = useState<LoadState>('loading')
  const [error, setError] = useState('')
  const [retryKey, setRetryKey] = useState(0)
  const [pendingModelId, setPendingModelId] = useState<string | null>(null)
  const [actionError, setActionError] = useState('')
  const [deleteTarget, setDeleteTarget] = useState<ProviderModel | null>(null)
  const [isDeleting, setIsDeleting] = useState(false)
  const [deleteError, setDeleteError] = useState('')

  useEffect(() => {
    let cancelled = false

    async function loadModels() {
      setLoadState('loading')
      setError('')

      try {
        const payload = await adminRequest<ModelsResponse>(
          accessToken,
          `/api/v1/providers/${provider.id}/models`,
        )

        if (!cancelled) {
          setModels(payload.models ?? [])
          setLoadState('ready')
        }
      } catch (requestError) {
        if (isUnauthorized(requestError)) {
          onUnauthorized()
          return
        }

        if (!cancelled) {
          setLoadState('error')
          setError(
            apiErrorMessage(requestError, 'Models could not be loaded.'),
          )
        }
      }
    }

    void loadModels()

    return () => {
      cancelled = true
    }
  }, [accessToken, onUnauthorized, provider.id, refreshKey, retryKey])

  async function toggleModel(model: ProviderModel) {
    setPendingModelId(model.model_id)
    setActionError('')

    try {
      const enabled = !model.enabled

      await adminRequest<{ message: string; enabled: boolean }>(
        accessToken,
        `/api/v1/providers/${provider.id}/models/${encodeURIComponent(model.model_id)}`,
        {
          method: 'PATCH',
          body: JSON.stringify({ enabled }),
        },
      )

      setModels((current) =>
        current.map((item) =>
          item.model_id === model.model_id ? { ...item, enabled } : item,
        ),
      )
      onNotify(
        'success',
        `${model.name} ${enabled ? 'enabled' : 'disabled'}.`,
      )
    } catch (requestError) {
      if (isUnauthorized(requestError)) {
        onUnauthorized()
        return
      }

      setActionError(
        apiErrorMessage(requestError, 'The model state could not be updated.'),
      )
    } finally {
      setPendingModelId(null)
    }
  }

  async function deleteModel() {
    if (!deleteTarget) {
      return
    }

    setIsDeleting(true)
    setDeleteError('')

    try {
      await adminRequest<{ message: string }>(
        accessToken,
        `/api/v1/providers/${provider.id}/models/${encodeURIComponent(deleteTarget.model_id)}`,
        { method: 'DELETE' },
      )

      setModels((current) =>
        current.filter((item) => item.model_id !== deleteTarget.model_id),
      )
      onNotify('success', `${deleteTarget.name} deleted.`)
      setDeleteTarget(null)
    } catch (requestError) {
      if (isUnauthorized(requestError)) {
        onUnauthorized()
        return
      }

      setDeleteError(
        apiErrorMessage(requestError, 'The model could not be deleted.'),
      )
    } finally {
      setIsDeleting(false)
    }
  }

  return (
    <section className="models-panel" aria-label={`${provider.name} models`}>
      <div className="models-heading">
        <div>
          <h3>Models</h3>
          <span>
            {loadState === 'ready'
              ? `${models.length} configured`
              : 'Provider catalog'}
          </span>
        </div>
        <button
          type="button"
          className="inline-action"
          onClick={() => onAddModel(provider)}
        >
          <PlusIcon />
          Add model
        </button>
      </div>

      {actionError ? (
        <p className="models-action-error" role="alert">
          {actionError}
        </p>
      ) : null}

      {loadState === 'loading' && models.length === 0 ? (
        <p className="models-message" role="status">
          Reading models…
        </p>
      ) : null}

      {loadState === 'error' ? (
        <div className="models-error" role="alert">
          <span>{error}</span>
          <button type="button" onClick={() => setRetryKey((key) => key + 1)}>
            Try again
          </button>
        </div>
      ) : null}

      {loadState === 'ready' && models.length === 0 ? (
        <p className="models-message">
          No models configured. Add one to make it available through this
          provider.
        </p>
      ) : null}

      {models.length > 0 ? (
        <div className="model-list">
          {models.map((model) => {
            const isPending = pendingModelId === model.model_id

            return (
              <div className="model-row" key={model.model_id}>
                <div className="model-identity">
                  <strong>{model.name}</strong>
                  <span>{model.model_id}</span>
                </div>

                <div className="model-actions">
                  <button
                    type="button"
                    role="switch"
                    aria-checked={model.enabled}
                    aria-label={`${model.enabled ? 'Disable' : 'Enable'} ${model.name}`}
                    className={`model-switch${model.enabled ? ' is-on' : ''}`}
                    onClick={() => void toggleModel(model)}
                    disabled={isPending}
                  >
                    <span className="model-switch-track" aria-hidden="true">
                      <span />
                    </span>
                    <span className="model-switch-label">
                      {isPending
                        ? 'Saving'
                        : model.enabled
                          ? 'Enabled'
                          : 'Disabled'}
                    </span>
                  </button>

                  <button
                    type="button"
                    className="icon-button is-danger"
                    aria-label={`Delete ${model.name}`}
                    title={`Delete ${model.name}`}
                    onClick={() => {
                      setDeleteError('')
                      setDeleteTarget(model)
                    }}
                  >
                    <TrashIcon />
                  </button>
                </div>
              </div>
            )
          })}
        </div>
      ) : null}

      {deleteTarget ? (
        <ConfirmDialog
          title={`Delete ${deleteTarget.name}?`}
          description={`This removes ${deleteTarget.model_id} from ${provider.name}. The action cannot be undone.`}
          confirmLabel="Delete model"
          isBusy={isDeleting}
          error={deleteError}
          onClose={() => {
            if (!isDeleting) {
              setDeleteTarget(null)
            }
          }}
          onConfirm={() => void deleteModel()}
        />
      ) : null}
    </section>
  )
}

export default ModelsPanel
