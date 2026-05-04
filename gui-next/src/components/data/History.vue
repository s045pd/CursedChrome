<script setup lang="ts">
import type { HistoryEntry } from '@/types/api'
import JsonList from './JsonList.vue'

defineProps<{ botId: string }>()

const asH = (r: unknown): HistoryEntry => r as HistoryEntry

function fmt(ts: number | undefined): string {
  if (!ts) return ''
  return new Date(ts).toLocaleString(undefined, {
    month: 'short',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}
</script>

<template>
  <JsonList
    :bot-id="botId"
    field="history"
    :search-keys="['title', 'url']"
    empty-message="No history entries."
  >
    <template #default="{ row }">
      <div class="px-3 py-2 hover:bg-bg-hover/60">
        <div class="flex items-baseline gap-3">
          <a
            :href="asH(row).url"
            target="_blank"
            rel="noopener"
            class="text-[12px] truncate hover:text-accent flex-1"
            :title="asH(row).url"
          >
            {{ asH(row).title || asH(row).url || '(no title)' }}
          </a>
          <span class="mono text-[10px] text-fg-faint shrink-0">
            {{ fmt(asH(row).lastVisitTime ?? asH(row).visitTime) }}
            <span v-if="asH(row).visitCount" class="ml-1.5">×{{ asH(row).visitCount }}</span>
          </span>
        </div>
        <div class="text-[10px] text-fg-faint mono truncate mt-0.5">{{ asH(row).url }}</div>
      </div>
    </template>
  </JsonList>
</template>
