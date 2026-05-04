// Thin fetch wrapper. Cookie-based session auth; same-origin so the
// browser handles the cookie automatically. We never read or store
// the session cookie in JS.
import type { ApiEnvelope } from '@/types/api'

class ApiError extends Error {
  status: number
  constructor(status: number, message: string) {
    super(message)
    this.status = status
  }
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  init?: RequestInit,
): Promise<T> {
  const res = await fetch(path, {
    method,
    credentials: 'same-origin',
    headers: body ? { 'Content-Type': 'application/json', ...(init?.headers ?? {}) } : init?.headers,
    body: body !== undefined ? JSON.stringify(body) : undefined,
    ...init,
  })

  // Bubble 401 up so the auth store can redirect to login.
  if (res.status === 401) {
    throw new ApiError(401, 'unauthorized')
  }

  // Binary endpoints (audio, image): caller will handle blob themselves.
  const ct = res.headers.get('content-type') ?? ''
  if (!ct.includes('application/json')) {
    if (!res.ok) throw new ApiError(res.status, `HTTP ${res.status}`)
    return (await res.blob()) as unknown as T
  }

  const json = (await res.json()) as ApiEnvelope<T>
  if (!json.success) {
    throw new ApiError(res.status, json.error ?? `HTTP ${res.status}`)
  }
  return json.result as T
}

export const api = {
  get: <T>(path: string) => request<T>('GET', path),
  post: <T>(path: string, body?: unknown) => request<T>('POST', path, body),
  put: <T>(path: string, body?: unknown) => request<T>('PUT', path, body),
  del: <T>(path: string, body?: unknown) => request<T>('DELETE', path, body),
  raw: request,
}

export { ApiError }
