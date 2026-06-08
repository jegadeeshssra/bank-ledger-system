import { useAuthStore } from '@/stores/authStore'

const API_BASE_PATH = import.meta.env.VITE_API_BASE_PATH || '/api/v1'
const BASE_URL = import.meta.env.VITE_API_BASE_URL || API_BASE_PATH

export function generateIdempotencyKey() {
  if (typeof crypto !== 'undefined' && typeof crypto.randomUUID === 'function') {
    return crypto.randomUUID()
  }

  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0
    const v = c === 'x' ? r : (r & 0x3) | 0x8
    return v.toString(16)
  })
}

// Flag to prevent multiple refresh attempts
let isRefreshing = false
let refreshPromise = null

async function refreshAccessToken() {
  if (isRefreshing) {
    return refreshPromise
  }

  isRefreshing = true
  refreshPromise = (async () => {
    try {
      const { refreshToken, token } = useAuthStore.getState()

      if (!refreshToken) {
        throw new Error('No refresh token available')
      }

      const headers = {
        'Content-Type': 'application/json',
        Authorization: `Bearer ${token}`,
      }

      const res = await fetch(`${BASE_URL}/auth/refresh`, {
        method: 'POST',
        headers,
        body: JSON.stringify({ refresh_token: refreshToken }),
      })

      if (!res.ok) {
        // Refresh failed, logout user
        useAuthStore.getState().logout()
        window.location.href = '/login'
        throw new Error('Failed to refresh token')
      }

      const data = await res.json()
      useAuthStore.getState().setToken(data.access_token)
      return data.access_token
    } catch (error) {
      isRefreshing = false
      refreshPromise = null
      throw error
    } finally {
      isRefreshing = false
      refreshPromise = null
    }
  })()

  return refreshPromise
}

async function request(path, options = {}) {
  let { token } = useAuthStore.getState()

  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  }

  let res = await fetch(`${BASE_URL}${path}`, { ...options, headers })

  // Handle token expiry and refresh
  if (res.status === 401) {
    try {
      const newToken = await refreshAccessToken()
      // Retry the original request with new token
      headers.Authorization = `Bearer ${newToken}`
      res = await fetch(`${BASE_URL}${path}`, { ...options, headers })
    } catch (error) {
      useAuthStore.getState().logout()
      window.location.href = '/login'
      throw new Error('Session expired. Please log in again.')
    }
  }

  if (res.status === 429) {
    throw new Error('Too many requests. Please wait a moment and try again.')
  }

  if (!res.ok) {
    if (res.status >= 500) {
      throw new Error('Something went wrong. Please try again.')
    }

    let message
    try {
      const text = await res.text()
      // Try to parse as JSON error
      try {
        const json = JSON.parse(text)
        message = json.error || json.message || text
      } catch {
        message = text
      }
    } catch {
      message = res.statusText
    }
    throw new Error(message || 'Something went wrong. Please try again.')
  }

  if (res.status === 204) return null
  return res.json()
}

// ── Auth ──────────────────────────────────────────────────────────────
export const register = (data) =>
  request('/auth/register', { method: 'POST', body: JSON.stringify(data) })

export const login = (data) =>
  request('/auth/login', { method: 'POST', body: JSON.stringify(data) })

export const refresh = (refreshToken) =>
  request('/auth/refresh', { method: 'POST', body: JSON.stringify({ refresh_token: refreshToken }) })

// ── Accounts ─────────────────────────────────────────────────────────
export const listAccounts = () => request('/accounts')

export const getAccount = (id) => request(`/accounts/${id}`)

export const createAccount = (data) =>
  request('/accounts', { method: 'POST', body: JSON.stringify(data) })

export const deleteAccount = (id) =>
  request(`/accounts/${id}`, { method: 'DELETE' })

// ── Ledger operations ─────────────────────────────────────────────────
export const deposit = (id, amount, idempotencyKey) =>
  request(`/accounts/${id}/deposit`, {
    method: 'POST',
    body: JSON.stringify({ amount }),
    headers: idempotencyKey
      ? { 'X-Idempotency-Key': idempotencyKey }
      : undefined,
  })

export const withdraw = (id, amount, idempotencyKey) =>
  request(`/accounts/${id}/withdraw`, {
    method: 'POST',
    body: JSON.stringify({ amount }),
    headers: idempotencyKey
      ? { 'X-Idempotency-Key': idempotencyKey }
      : undefined,
  })

export const transfer = (id, data, idempotencyKey) =>
  request(`/accounts/${id}/transfers`, {
    method: 'POST',
    body: JSON.stringify(data),
    headers: idempotencyKey
      ? { 'X-Idempotency-Key': idempotencyKey }
      : undefined,
  })

export const reconcile = (id) =>
  request(`/accounts/${id}/reconcile`, { method: 'PUT' })

// ── Read-only data ────────────────────────────────────────────────────
export const getEntries = (id) => request(`/accounts/${id}/entries`)

export const getTransaction = (accountId, transactionId) =>
  request(`/accounts/${accountId}/transactions/${transactionId}`)
