import { useId, useState } from 'react'
import type { FormEvent } from 'react'
import Modal from './Modal'
import {
  apiErrorMessage,
  type AddModelInput,
  type AddProviderInput,
} from './api'

type AddProviderDialogProps = {
  onClose: () => void
  onSubmit: (input: AddProviderInput) => Promise<void>
}

export function AddProviderDialog({
  onClose,
  onSubmit,
}: AddProviderDialogProps) {
  const nameId = useId()
  const baseUrlId = useId()
  const apiKeyId = useId()
  const [name, setName] = useState('')
  const [baseUrl, setBaseUrl] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [error, setError] = useState('')
  const [isSaving, setIsSaving] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (!name.trim()) {
      setError('Enter a provider name.')
      return
    }

    if (!baseUrl.trim() && name.trim().toLowerCase() !== 'groq') {
      setError('Enter a base URL.')
      return
    }

    if (!apiKey.trim()) {
      setError('Enter an API key.')
      return
    }

    setIsSaving(true)
    setError('')

    try {
      await onSubmit({
        name: name.trim(),
        base_url: baseUrl.trim(),
        api_key: apiKey.trim(),
      })
    } catch (requestError) {
      setError(apiErrorMessage(requestError, 'The provider could not be added.'))
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <Modal
      title="Add provider"
      description="Connect an upstream API. The API key is encrypted before storage."
      onClose={onClose}
    >
      <form className="modal-form" onSubmit={handleSubmit} noValidate>
        <div className="modal-field">
          <label htmlFor={nameId}>Provider name</label>
          <input
            id={nameId}
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="groq"
            autoComplete="off"
            autoFocus
          />
        </div>

        <div className="modal-field">
          <label htmlFor={baseUrlId}>Base URL</label>
          <input
            id={baseUrlId}
            value={baseUrl}
            onChange={(event) => setBaseUrl(event.target.value)}
            placeholder="https://api.example.com/v1"
            inputMode="url"
            autoComplete="off"
          />
          <span className="field-hint">
            Leave blank for Groq; the server supplies its default endpoint.
          </span>
        </div>

        <div className="modal-field">
          <label htmlFor={apiKeyId}>API key</label>
          <input
            id={apiKeyId}
            type="password"
            value={apiKey}
            onChange={(event) => setApiKey(event.target.value)}
            placeholder="Paste the provider API key"
            autoComplete="new-password"
          />
        </div>

        {error ? (
          <p className="modal-error" role="alert">
            {error}
          </p>
        ) : null}

        <div className="modal-actions">
          <button
            type="button"
            className="modal-button is-secondary"
            onClick={onClose}
            disabled={isSaving}
          >
            Cancel
          </button>
          <button
            type="submit"
            className="modal-button is-primary"
            disabled={isSaving}
          >
            {isSaving ? 'Adding…' : 'Add provider'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

type AddModelDialogProps = {
  providerName: string
  onClose: () => void
  onSubmit: (input: AddModelInput) => Promise<void>
}

export function AddModelDialog({
  providerName,
  onClose,
  onSubmit,
}: AddModelDialogProps) {
  const modelId = useId()
  const nameId = useId()
  const [modelIdentifier, setModelIdentifier] = useState('')
  const [name, setName] = useState('')
  const [error, setError] = useState('')
  const [isSaving, setIsSaving] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()

    if (!modelIdentifier.trim()) {
      setError('Enter a model ID.')
      return
    }

    if (!name.trim()) {
      setError('Enter a display name.')
      return
    }

    setIsSaving(true)
    setError('')

    try {
      await onSubmit({
        model_id: modelIdentifier.trim(),
        name: name.trim(),
      })
    } catch (requestError) {
      setError(apiErrorMessage(requestError, 'The model could not be added.'))
    } finally {
      setIsSaving(false)
    }
  }

  return (
    <Modal
      title="Add model"
      description={`Add a model to ${providerName}.`}
      onClose={onClose}
    >
      <form className="modal-form" onSubmit={handleSubmit} noValidate>
        <div className="modal-field">
          <label htmlFor={modelId}>Model ID</label>
          <input
            id={modelId}
            value={modelIdentifier}
            onChange={(event) => setModelIdentifier(event.target.value)}
            placeholder="llama-3.3-70b-versatile"
            autoComplete="off"
            autoFocus
          />
          <span className="field-hint">
            Use the exact identifier sent to the provider API.
          </span>
        </div>

        <div className="modal-field">
          <label htmlFor={nameId}>Display name</label>
          <input
            id={nameId}
            value={name}
            onChange={(event) => setName(event.target.value)}
            placeholder="Llama 3.3 70B"
            autoComplete="off"
          />
        </div>

        {error ? (
          <p className="modal-error" role="alert">
            {error}
          </p>
        ) : null}

        <div className="modal-actions">
          <button
            type="button"
            className="modal-button is-secondary"
            onClick={onClose}
            disabled={isSaving}
          >
            Cancel
          </button>
          <button
            type="submit"
            className="modal-button is-primary"
            disabled={isSaving}
          >
            {isSaving ? 'Adding…' : 'Add model'}
          </button>
        </div>
      </form>
    </Modal>
  )
}

type ConfirmDialogProps = {
  title: string
  description: string
  confirmLabel: string
  isBusy: boolean
  error?: string
  onClose: () => void
  onConfirm: () => void
}

export function ConfirmDialog({
  title,
  description,
  confirmLabel,
  isBusy,
  error,
  onClose,
  onConfirm,
}: ConfirmDialogProps) {
  return (
    <Modal title={title} onClose={onClose}>
      <p className="confirm-copy">{description}</p>

      {error ? (
        <p className="modal-error" role="alert">
          {error}
        </p>
      ) : null}

      <div className="modal-actions">
        <button
          type="button"
          className="modal-button is-secondary"
          onClick={onClose}
          disabled={isBusy}
        >
          Cancel
        </button>
        <button
          type="button"
          className="modal-button is-danger"
          onClick={onConfirm}
          disabled={isBusy}
        >
          {isBusy ? 'Deleting…' : confirmLabel}
        </button>
      </div>
    </Modal>
  )
}
