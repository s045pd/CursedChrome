<script setup lang="ts">
import JsonList from './JsonList.vue'
interface Download {
  filename?: string
  url?: string
  state?: string
  bytesReceived?: number
  totalBytes?: number
  startTime?: string
}
defineProps<{ botId: string }>()
const asD = (r: unknown): Download => r as Download

function size(b: number | undefined): string {
  if (!b) return ''
  const u = ['B', 'KB', 'MB', 'GB']
  let i = 0
  let v = b
  while (v > 1024 && i < u.length - 1) {
    v /= 1024
    i++
  }
  return `${v.toFixed(i === 0 ? 0 : 1)} ${u[i]}`
}
</script>

<template>
  <JsonList
    :bot-id="botId"
    field="downloads"
    :search-keys="['filename', 'url']"
    empty-message="No downloads recorded."
  >
    <template #default="{ row }">
      <div class="px-3 py-2 hover:bg-bg-hover/60">
        <div class="flex items-baseline gap-2">
          <span class="text-[12px] truncate flex-1">
            {{ asD(row).filename || asD(row).url }}
          </span>
          <span class="chip">{{ asD(row).state || 'unknown' }}</span>
          <span class="mono text-[10px] text-fg-faint">
            {{ size(asD(row).bytesReceived) }}
          </span>
        </div>
        <div v-if="asD(row).url" class="text-[10px] text-fg-faint mono truncate">
          {{ asD(row).url }}
        </div>
      </div>
    </template>
  </JsonList>
</template>
