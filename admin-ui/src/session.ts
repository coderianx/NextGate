const adminTokenKey = 'nextgate.admin.accessToken'

export function readAdminToken() {
  try {
    return sessionStorage.getItem(adminTokenKey)
  } catch {
    return null
  }
}

export function saveAdminToken(token: string) {
  sessionStorage.setItem(adminTokenKey, token)
}

export function clearAdminToken() {
  sessionStorage.removeItem(adminTokenKey)
}
