<script setup lang="ts">
import type { BotTab } from '@/types/api'
import JsonList from './JsonList.vue'

defineProps<{ botId: string }>()

// Slot rows arrive as Record<string, unknown>; cast once per slot use.
const asTab = (r: unknown): BotTab => r as BotTab
</script>

<template>
  <JsonList
    :bot-id="botId"
    field="tabs"
    :search-keys="['title', 'url']"
    empty-message="Bot has no tabs reported."
  >
    <template #default="{ row }">
      <div class="flex items-center gap-3 px-3 py-2 hover:bg-bg-hover/60">
        <img
          v-if="asTab(row).favIconUrl"
          :src="asTab(row).favIconUrl"
          alt=""
          class="size-4 rounded-sm"
          loading="lazy"
        />
        <div v-else class="size-4" />
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
