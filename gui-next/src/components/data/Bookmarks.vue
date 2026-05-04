<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { bots as botsApi } from '@/api/endpoints'
import Field from '@/components/ui/Field.vue'
import Btn from '@/components/ui/Btn.vue'

interface Bookmark { title?: string; url?: string; dateAdded?: number; children?: Bookmark[] }

const props = defineProps<{ botId: string }>()
const raw = ref<unknown[]>([])
const filter = ref('')
const loading = ref(false)

function flatten(nodes: unknown[]): Bookmark[] {
  const out: Bookmark[] = []
  for (const n of nodes) {
    const node = n as Bookmark
    if (node.url) out.push(node)
    if (Array.isArray(node.children)) out.push(...flatten(node.children))
  }
  return out
}

async function load(): Promise<void> {
  loading.value = true
  try {
    raw.value = ((await botsApi.field<unknown[]>(props.botId, 'bookmarks')) ?? []) as unknown[]
  } finally {
    loading.value = false
  }
}
watch(() => props.botId, load, { immediate: true })

const bookmarks = computed(() => flatten(raw.value))

const filtered = computed(() => {
  const q = filter.value.trim().toLowerCase()
  if (!q) return bookmarks.value
  return bookmarks.value.filter(
    (b) =>
      (b.title && b.title.toLowerCase().includes(q)) ||
      (b.url && b.url.toLowerCase().includes(q)),
  )
})
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center gap-2">
      <Field v-model="filter" placeholder="Filter" size="sm" :block="false" />
      <span class="text-[11px] text-fg-faint mono ml-auto">
        {{ filtered.length }} / {{ bookmarks.length }}
      </span>
      <Btn size="sm" variant="ghost" :loading="loading" @click="load">Reload</Btn>
    </div>

    <div v-if="!loading && filtered.length === 0" class="text-center py-10 text-fg-faint text-[12px]">
      No bookmarks reported.
    </div>

    <div v-else class="surface divide-y divide-border-subtle">
      <div v-for="(b, i) in filtered" :key="i" class="px-3 py-2 hover:bg-bg-hover/60">
        <a
          v-if="b.url"
          :href="b.url"
          target="_blank"
          rel="noopener"
          class="text-[12px] truncate block hover:text-accent"
        >
          {{ b.title || b.url }}
        </a>
        <span v-else class="text-[12px] text-fg-faint">{{ b.title }}</span>
        <div v-if="b.url" class="text-[10px] text-fg-faint mono truncate">
          {{ b.url }}
        </div>
      </div>
    </div>
  </div>
</template>
