export type Provider = {
  id: number
  name: string
  base_url: string
  created_at: string
}

export type ProviderModel = {
  provider_id: number
  model_id: string
  name: string
  enabled: boolean
}

export type AddProviderInput = {
  name: string
  base_url: string
  api_key: string
}

export type AddModelInput = {
  model_id: string
  name: string
}

type ApiErrorBody = {
  error?: string
}

export class AdminApiError extends Error {
  readonly status: number
  readonly code: string

  constructor(status: number, code: string) {
    super(code)
    this.name = 'AdminApiError'
    this.status = status
    this.code = code
  }
}

const errorMessages: Record<string, string> = {
  invalid_request: 'Check the form and try again.',
  name_is_required: 'Enter a name.',
  base_url_is_required: 'Enter a base URL.',
  api_key_is_required: 'Enter an API key.',
  model_id_is_required: 'Enter a model ID.',
  invalid_provider_id: 'The provider ID is invalid.',
  provider_not_found: 'The provider no longer exists.',
  model_not_found: 'The model no longer exists.',
  database_not_connected: 'The database is not connected.',
  database_query_error: 'The requested data could not be read.',
  database_insert_error:
    'The provider could not be saved. Its endpoint may already exist.',
  database_update_error: 'The model state could not be updated.',
  database_delete_error: 'The item could not be deleted.',
  database_error: 'The database operation failed.',
  aes_256_gcm_get_secret_key:
    'The encryption key is unavailable. Check the OS keyring and restart NextGate.',
  encrypt_api_key_error: 'The API key could not be encrypted.',
  you_are_already_exists_have_this_model:
    'This model already exists for the provider.',
  request_failed: 'The request failed.',
}

export async function adminRequest<T>(
  accessToken: string,
  path: string,
  init: RequestInit = {},
): Promise<T> {
  const headers = new Headers(init.headers)
  headers.set('Authorization', `Bearer ${accessToken}`)

  if (init.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json')
  }

  const response = await fetch(path, { ...init, headers })
  let payload: unknown = null

  try {
    payload = await response.json()
  } catch {
    payload = null
  }

  if (!response.ok) {
    const body = payload as ApiErrorBody | null
    throw new AdminApiError(
      response.status,
      body?.error ?? 'request_failed',
    )
  }

  return (payload ?? {}) as T
}

export function isUnauthorized(error: unknown) {
  return (
    error instanceof AdminApiError &&
    (error.status === 401 || error.status === 403)
  )
}

export function apiErrorMessage(error: unknown, fallback: string) {
  if (error instanceof AdminApiError) {
    return errorMessages[error.code] ?? fallback
  }

  if (error instanceof TypeError) {
    return 'The admin API is unreachable. Check that NextGate is running.'
  }

  return fallback
}
