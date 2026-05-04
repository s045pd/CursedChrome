<script setup lang="ts">
import { ref } from 'vue'
import type { BotTab } from '@/types/api'
import { faviconUrl } from '@/composables/useFavicon'
import JsonList from './JsonList.vue'

defineProps<{ botId: string }>()

const asTab = (r: unknown): BotTab => r as BotTab

const broken = ref(new Set<number>())
function onImgError(idx: number) {
  broken.value = new Set(broken.value).add(idx)
}

function iconSrc(tab: BotTab): string {
  return tab.favIconUrl || faviconUrl(tab.url)
}
</script>

<template>
  <JsonList
    :bot-id="botId"
    field="tabs"
    :search-keys="['title', 'url']"
    empty-message="Bot has no tabs reported."
  >
    <template #default="{ row, i }">
      <div class="flex items-center gap-3 px-3 py-2 hover:bg-bg-hover/60">
        <img
          v-if="iconSrc(asTab(row)) && !broken.has(i)"
          :src="iconSrc(asTab(row))"
          alt=""
          class="size-4 rounded-sm shrink-0"
          loading="lazy"
          @error="onImgError(i)"
        />
        <svg v-else class="size-4 shrink-0 text-fg-faint" viewBox="0 0 16 16" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="1" y="1" width="14" height="14" rx="2" stroke="currentColor" stroke-width="1.2" />
          <circle cx="8" cy="8" r="3" stroke="currentColor" stroke-width="1.2" />
        </svg>
        <a
          :href="asTab(row).url"
          target="_blank"
          rel="noopener"
          class="text-[12px] truncate hover:text-accent flex-1"
          :title="asTab(row).url"
        >
          {{ asTab(row).title || asTab(row).url || '(blank)' }}
        </a>
        <span v-if="asTab(row).active" class="chip text-success">active</span>
        <span v-if="asTab(row).audible" class="chip">audio</span>
      </div>
    </template>
  </JsonList>
</template>
