<script setup lang="ts">
import { computed, onBeforeUnmount, onMounted, ref, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useBotsStore } from '@/stores/bots'
import { bots as botsApi } from '@/api/endpoints'
import type { BotSummary } from '@/types/api'
import BotRow from '@/components/bot/BotRow.vue'
import Btn from '@/components/ui/Btn.vue'
import Field from '@/components/ui/Field.vue'

const store = useBotsStore()
const router = useRouter()

const filterName = ref('')
const filterOnline = ref<string>('all') // 'all' | 'true' | 'false'

const allChecked = computed(() => {
  if (store.list.length === 0) return false
  return store.list.every((b: BotSummary) => store.selected.has(b.id))
})

function applyFilters(): void {
  store.setFilters({
    name: filterName.value || undefined,
    isOnline: filterOnline.value === 'all' ? null : filterOnline.value === 'true',
    page: 1,
  })
}

watch(filterOnline, applyFilters)

let typeTimer: ReturnType<typeof setTimeout> | null = null
watch(filterName, () => {
  if (typeTimer) clearTimeout(typeTimer)
  typeTimer = setTimeout(applyFilters, 200)
})

function toggleAll(): void {
  if (allChecked.value) store.clearSelection()
  else store.selectAll(store.list.map((b: BotSummary) => b.id))
}

function open(id: string): void {
  void router.push({ name: 'bot-detail', params: { id } })
}

async function setProxy(id: string): Promise<void> {
  if (store.globalProxy === id) await store.clearGlobalProxy()
  else await store.setGlobalProxy(id)
}

async function deleteOne(id: string): Promise<void> {
  if (!window.confirm('Delete this bot and all its data?')) return
  await botsApi.delete(id)
  await store.fetch()
}

async function deleteSelected(): Promise<void> {
  const n = store.selected.size
  if (n === 0) return
  if (!window.confirm(`Delete ${n} selected bot${n === 1 ? '' : 's'} and all their data?`)) return
  await store.deleteSelected()
}

function gotoPage(p: number): void {
  store.setFilters({ page: p })
}

onMounted(async () => {
  await store.fetchGlobalProxy()
  store.startPolling()
})
onBeforeUnmount(() => store.stopPolling())
</script>

<template>
  <div class="px-6 pt-5 pb-10 max-w-[1600px] mx-auto">
    <!-- header row -->
    <div class="flex items-end justify-between mb-5">
      <div>
        <h1 class="text-[20px] font-semibold tracking-tight">Bots</h1>
        <p class="text-[12px] text-fg-faint mt-0.5">
          Connected browsers across the fleet
        </p>
      </div>
      <div class="flex items-center gap-2">
        <Btn
          v-if="store.selected.size > 0"
          variant="danger"
          size="md"
          @click="deleteSelected"
        >
          Delete {{ store.selected.size }} selected
        </Btn>
        <Btn variant="ghost" size="md" @click="store.fetch()">
          Refresh
        </Btn>
      </div>
    </div>

    <!-- stat strip -->
    <div class="grid grid-cols-4 gap-3 mb-5">
      <div class="surface px-4 py-3">
        <div class="text-[11px] uppercase tracking-wider text-fg-faint">Total</div>
        <div class="text-[22px] font-semibold mono mt-0.5">
          {{ store.pagination?.total ?? 0 }}
        </div>
      </div>
      <div class="surface px-4 py-3">
        <div class="text-[11px] uppercase tracking-wider text-fg-faint">Online</div>
        <div class="text-[22px] font-semibold mono mt-0.5 text-success">
          {{ store.onlineCount }}
        </div>
      </div>
      <div class="surface px-4 py-3">
        <div class="text-[11px] uppercase tracking-wider text-fg-faint">Offline</div>
        <div class="text-[22px] font-semibold mono mt-0.5 text-fg-muted">
          {{ store.offlineCount }}
        </div>
      </div>
      <div class="surface px-4 py-3">
        <div class="text-[11px] uppercase tracking-wider text-fg-faint">
          Default proxy
        </div>
        <div class="text-[12px] mono mt-1.5 truncate" :title="store.globalProxy ?? ''">
          {{ store.globalProxy ? store.globalProxy.slice(0, 18) + '…' : 'none' }}
        </div>
      </div>
    </div>

    <!-- filters -->
    <div class="flex items-center gap-3 mb-3">
      <Field
        v-model="filterName"
        placeholder="Filter by name"
        size="sm"
        :block="false"
      />
      <select
        v-model="filterOnline"
        class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle text-fg-base focus:outline-none focus:border-accent"
      >
        <option value="all">All</option>
        <option value="true">Online</option>
        <option value="false">Offline</option>
      </select>
      <span class="text-[11px] text-fg-faint mono ml-auto">
        showing {{ store.list.length }} of {{ store.pagination?.total ?? 0 }}
      </span>
    </div>

    <!-- table -->
    <div class="surface overflow-hidden">
      <table class="w-full text-left">
        <thead>
          <tr
            class="text-[10px] uppercase tracking-wider text-fg-faint border-b border-border-subtle bg-bg-overlay"
          >
            <th class="px-3 py-2 w-8">
              <input
                type="checkbox"
                :checked="allChecked"
                class="size-3.5 accent-accent"
                @change="toggleAll"
              />
            </th>
            <th class="px-2 py-2 w-[88px]">Capture</th>
            <th class="px-3 py-2">Bot</th>
            <th class="px-3 py-2 w-[120px]">Status</th>
            <th class="px-3 py-2">Current tab</th>
            <th class="px-3 py-2 w-[100px]">Tabs / hist</th>
            <th class="px-3 py-2 w-[200px]">Proxy creds</th>
            <th class="px-3 py-2 w-[120px]">Last active</th>
            <th class="px-3 py-2 w-[180px]">Actions</th>
          </tr>
        </thead>
        <tbody>
          <BotRow
            v-for="b in store.list"
            :key="b.id"
            :bot="b"
            :selected="store.selected.has(b.id)"
            :global-proxy="store.globalProxy === b.id"
            @toggle="store.toggleSelected"
            @open="open"
            @set-proxy="setProxy"
            @delete="deleteOne"
          />
          <tr v-if="!store.loading && store.list.length === 0">
            <td colspan="9" class="text-center py-12 text-fg-faint text-[12px]">
              No bots match the current filter.
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- pagination -->
    <div
      v-if="store.pagination && store.pagination.totalPages > 1"
      class="flex items-center justify-end gap-2 mt-3"
    >
      <button
        class="h-7 px-2.5 text-[11px] rounded border border-border-subtle disabled:opacity-40"
        :disabled="(store.filters.page ?? 1) <= 1"
        @click="gotoPage((store.filters.page ?? 1) - 1)"
      >
        ← Prev
      </button>
      <span class="text-[11px] mono text-fg-muted">
        page {{ store.filters.page ?? 1 }} / {{ store.pagination.totalPages }}
      </span>
      <button
        class="h-7 px-2.5 text-[11px] rounded border border-border-subtle disabled:opacity-40"
        :disabled="
          (store.filters.page ?? 1) >= (store.pagination?.totalPages ?? 1)
        "
        @click="gotoPage((store.filters.page ?? 1) + 1)"
      >
        Next →
      </button>
    </div>
  </div>
</template>
