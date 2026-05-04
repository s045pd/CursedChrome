import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { bots as botsApi, settings as settingsApi } from '@/api/endpoints'
import type { ListBotsParams } from '@/api/endpoints'
import type { BotSummary, Pagination } from '@/types/api'

const REFRESH_MS = 5000

export const useBotsStore = defineStore('bots', () => {
  const list = ref<BotSummary[]>([])
  const pagination = ref<Pagination | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const filters = ref<ListBotsParams>({ page: 1, limit: 20 })
  const selected = ref<Set<string>>(new Set())
  const globalProxy = ref<string | null>(null)

  let timer: ReturnType<typeof setTimeout> | null = null

  const onlineCount = computed(() => list.value.filter((b: BotSummary) => b.is_online).length)
  const offlineCount = computed(() => list.value.filter((b: BotSummary) => !b.is_online).length)

  async function fetch(): Promise<void> {
    loading.value = true
    error.value = null
    try {
      const r = await botsApi.list(filters.value)
      list.value = r.bots
      pagination.value = r.pagination
    } catch (e) {
      error.value = e instanceof Error ? e.message : 'failed to load bots'
    } finally {
      loading.value = false
    }
  }

  async function fetchGlobalProxy(): Promise<void> {
    try {
      globalProxy.value = await settingsApi.getGlobalProxy()
    } catch {
      globalProxy.value = null
    }
  }

  async function setGlobalProxy(botId: string): Promise<void> {
    await settingsApi.setGlobalProxy(botId)
    globalProxy.value = botId
  }

  async function clearGlobalProxy(): Promise<void> {
    await settingsApi.setGlobalProxy('')
    globalProxy.value = null
  }

  function setFilters(patch: Partial<ListBotsParams>): void {
    filters.value = { ...filters.value, ...patch, page: patch.page ?? 1 }
    void fetch()
  }

  function toggleSelected(id: string): void {
    if (selected.value.has(id)) selected.value.delete(id)
    else selected.value.add(id)
    selected.value = new Set(selected.value)
  }

  function selectAll(ids: string[]): void {
    selected.value = new Set(ids)
  }

  function clearSelection(): void {
    selected.value = new Set()
  }

  async function deleteSelected(): Promise<number> {
    if (selected.value.size === 0) return 0
    const ids = Array.from(selected.value)
    const r = await botsApi.batchDelete(ids)
    clearSelection()
    await fetch()
    return r.deletedCount
  }

  function startPolling(): void {
    stopPolling()
    const tick = async (): Promise<void> => {
      await fetch()
      timer = setTimeout(() => void tick(), REFRESH_MS)
    }
    void tick()
  }

  function stopPolling(): void {
    if (timer) {
      clearTimeout(timer)
      timer = null
    }
  }

  return {
    list,
    pagination,
    loading,
    error,
    filters,
    selected,
    globalProxy,
    onlineCount,
    offlineCount,
    fetch,
    fetchGlobalProxy,
    setGlobalProxy,
    clearGlobalProxy,
    setFilters,
    toggleSelected,
    selectAll,
    clearSelection,
    deleteSelected,
    startPolling,
    stopPolling,
  }
})
