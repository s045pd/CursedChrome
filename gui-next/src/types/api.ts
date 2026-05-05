// API contract shared with Go backend (cursed-go/internal/api).
// Keep this file the single source of truth and keep field names
// aligned with botSummary in cursed-go/internal/api/bots.go.

export interface ApiEnvelope<T> {
  success: boolean
  result?: T
  error?: string
}

export interface BotSummary {
  id: string
  name: string
  browser_id: string
  is_online: boolean
  last_online: string
  last_active_at?: string
  createdAt: string
  proxy_username: string
  proxy_password: string
  state: string
  user_agent: string
  current_tab: { id?: number; url?: string; title?: string } | null
  current_tab_image: boolean
  tabs: number
  history: number
  switch_config: Record<string, unknown>
  data_config: Record<string, unknown>
}

export interface Pagination {
  total: number
  page: number
  limit: number
  totalPages: number
}

export interface BotListResult {
  bots: BotSummary[]
  pagination: Pagination
}

export interface MeResult {
  username: string
  password_should_be_changed: boolean
}

export type SwitchConfigKey =
  | 'SYNC'
  | 'SYNC_HUGE'
  | 'REALTIME_IMG'
  | 'NOTIFICATION'
  | 'PERSISTENT_RECORDING'
  | 'PERSISTENT_KEYBOARD'

export interface BotTab {
  id?: number
  url?: string
  title?: string
  active?: boolean
  audible?: boolean
  favIconUrl?: string
}

export interface HistoryEntry {
  url?: string
  title?: string
  visitCount?: number
  visitTime?: number
  lastVisitTime?: number
}

export interface CookieEntry {
  name: string
  value: string
  domain?: string
  path?: string
  secure?: boolean
  httpOnly?: boolean
  sameSite?: string
  expirationDate?: number
}

export interface ScreenshotEntry {
  ID: string
  BotID: string
  URL?: string
  Title?: string
  ImageData?: string
  Timestamp: string
  SessionID?: string
  Difference?: number
  HasImage?: boolean
}

export interface KeyboardLogEntry {
  ID: string
  BotID: string
  URL?: string
  Title?: string
  Keys: string
  Timestamp: string
}

export interface AudioSession {
  session_id: string
  start_time: string
  end_time: string
  chunk_count: number
}
