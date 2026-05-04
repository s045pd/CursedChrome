<script setup lang="ts">
import JsonList from './JsonList.vue'
interface Bookmark { title?: string; url?: string; dateAdded?: number }
defineProps<{ botId: string }>()
const asB = (r: unknown): Bookmark => r as Bookmark
</script>

<template>
  <JsonList
    :bot-id="botId"
    field="bookmarks"
    :search-keys="['title', 'url']"
    empty-message="No bookmarks reported."
  >
    <template #default="{ row }">
      <div class="px-3 py-2 hover:bg-bg-hover/60">
        <a
          v-if="asB(row).url"
          :href="asB(row).url"
          target="_blank"
          rel="noopener"
          class="text-[12px] truncate block hover:text-accent"
        >
          {{ asB(row).title || asB(row).url }}
        </a>
        <span v-else class="text-[12px] text-fg-faint">{{ asB(row).title }}</span>
        <div v-if="asB(row).url" class="text-[10px] text-fg-faint mono truncate">
          {{ asB(row).url }}
        </div>
      </div>
    </template>
  </JsonList>
</template>
