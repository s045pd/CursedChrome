<script setup lang="ts">
import { ref } from 'vue'
import type { HistoryEntry } from '@/types/api'
import { faviconUrl } from '@/composables/useFavicon'
import JsonList from './JsonList.vue'

defineProps<{ botId: string }>()

const asH = (r: unknown): HistoryEntry => r as HistoryEntry

const broken = ref(new Set<number>())
function onImgError(idx: number) {
  broken.value = new Set(broken.value).add(idx)
}

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
    <template #default="{ row, i }">
      <div class="flex items-center gap-3 px-3 py-2 hover:bg-bg-hover/60">
        <img
          v-if="faviconUrl(asH(row).url) && !broken.has(i)"
          :src="faviconUrl(asH(row).url)"
          alt=""
          class="size-4 rounded-sm shrink-0"
          loading="lazy"
          @error="onImgError(i)"
        />
        <svg v-else class="size-4 shrink-0 text-fg-faint" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="1" y="1" width="14" height="14" rx="2" stroke="currentColor" stroke-width="1.2" />
          <circle cx="8" cy="8" r="3" stroke="currentColor" stroke-width="1.2" />
        </svg>
        <div class="flex-1 min-w-0">
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
      </div>
    </template>
  </JsonList>
</template>
