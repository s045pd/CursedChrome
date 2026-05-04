// All REST endpoint wrappers. One function per route — keep callers
// in components, stores or composables free of URL string handling.
import { api } from './client'
import type {
  BotListResult,
  BotSummary,
  MeResult,
  ScreenshotEntry,
  KeyboardLogEntry,
  AudioSession,
} from '@/types/api'

export type { BotSummary, BotListResult } from '@/types/api'

export interface ListBotsParams {
  page?: number
  limit?: number
  name?: string
  isOnline?: boolean | null
  state?: string
}

function qs(params: Record<string, unknown>): string {
  const usp = new URLSearchParams()
  for (const [k, v] of Object.entries(params)) {
    if (v === undefined || v === null || v === '') continue
    usp.set(k, String(v))
  }
  const s = usp.toString()
  return s ? `?${s}` : ''
}

export const auth = {
  login: (username: string, password: string) =>
    api.post<MeResult>('/api/v1/login', { username, password }),
  logout: () => api.get<Record<string, never>>('/api/v1/logout'),
  me: () => api.get<MeResult>('/api/v1/me'),
  changePassword: (newPassword: string) =>
    api.put<Record<string, never>>('/api/v1/password', { new_password: newPassword }),
}

export const bots = {
  list: (p: ListBotsParams = {}) =>
    api.get<BotListResult>(
      '/api/v1/bots' +
        qs({
          page: p.page,
          limit: p.limit,
          name: p.name,
          is_online: p.isOnline === null ? '' : p.isOnline,
          state: p.state,
        }),
    ),

  update: (botId: string, patch: Partial<Pick<BotSummary, 'name' | 'proxy_username' | 'proxy_password'>> & {
    switch_config?: Record<string, unknown>
    data_config?: Record<string, unknown>
  }) => api.put<Record<string, never>>('/api/v1/bots', { bot_id: botId, ...patch }),

  delete: (botId: string) => api.del<Record<string, never>>('/api/v1/bots', { bot_id: botId }),

  batchDelete: (botIds: string[]) =>
    api.post<{ deletedCount: number }>('/api/v1/bots/batch-delete', { bot_ids: botIds }),

  imageURL: (botId: string) => `/api/v1/bots/image/${botId}`,

  field: <T = unknown>(botId: string, field: string) =>
    api.get<T>(`/api/v1/fields${qs({ field, id: botId })}`),
}

export const settings = {
  getGlobalProxy: () => api.get<string | null>('/api/v1/settings/global-proxy'),
  setGlobalProxy: (botId: string) =>
    api.post<string>('/api/v1/settings/global-proxy', { bot_id: botId }),
}

export const remote = {
  navigate: (botId: string, url: string) =>
    api.post<unknown>('/api/v1/remote-control', { bot_id: botId, url }),
  stop: (botId: string) =>
    api.post<Record<string, never>>('/api/v1/stop-remote-control', { bot_id: botId }),
  startAudio: (botId: string) =>
    api.post<Record<string, never>>('/api/v1/start-audio', { bot_id: botId }),
  stopAudio: (botId: string) =>
    api.post<Record<string, never>>('/api/v1/stop-audio', { bot_id: botId }),
}

export const media = {
  screenshots: (botId: string, limit = 50, offset = 0) =>
    api.get<ScreenshotEntry[]>(
      `/api/v1/screenshots${qs({ id: botId, limit, offset })}`,
    ),
  screenshotImageURL: (id: string) => `/api/v1/screenshots/${id}/image`,
  keyboardLogs: (
    botId: string,
    limit = 50,
    offset = 0,
    range?: { startTime?: string; endTime?: string },
  ) =>
    api.get<KeyboardLogEntry[]>(
      `/api/v1/keyboard-logs${qs({
        id: botId,
        limit,
        offset,
        startTime: range?.startTime,
        endTime: range?.endTime,
      })}`,
    ),
  recordings: (botId: string) =>
    api.get<unknown[]>(`/api/v1/recordings${qs({ id: botId })}`),
  audioSessions: (botId: string) =>
    api.get<AudioSession[]>(`/api/v1/audio-sessions${qs({ id: botId })}`),
  audioSessionURL: (sessionId: string) => `/api/v1/audio-session/${sessionId}`,
  audioChunkURL: (id: string) => `/api/v1/audio/${id}`,
}
