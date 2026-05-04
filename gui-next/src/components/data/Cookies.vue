<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { bots as botsApi } from '@/api/endpoints'
import type { CookieEntry } from '@/types/api'
import Field from '@/components/ui/Field.vue'
import Btn from '@/components/ui/Btn.vue'
import { useClipboard } from '@/composables/useClipboard'

const props = defineProps<{ botId: string }>()
const list = ref<CookieEntry[]>([])
const filter = ref('')
const loading = ref(false)
const { copy, copied } = useClipboard()

async function load(): Promise<void> {
  loading.value = true
  try {
    list.value = (await botsApi.field<CookieEntry[]>(props.botId, 'cookies')) ?? []
  } finally {
    loading.value = false
  }
}

watch(() => props.botId, load, { immediate: true })

const filtered = computed(() => {
  const q = filter.value.trim().toLowerCase()
  if (!q) return list.value
  return list.value.filter(
    (c) =>
      c.name?.toLowerCase().includes(q) ||
      c.domain?.toLowerCase().includes(q) ||
      c.value?.toLowerCase().includes(q),
  )
})

function exportJson(): void {
  copy(JSON.stringify(list.value, null, 2), 'export')
}
function exportNetscape(): void {
  // Netscape cookie format the curl --cookie-jar tool consumes.
  const lines = list.value.map((c) => {
    const flag = c.domain?.startsWith('.') ? 'TRUE' : 'FALSE'
    const secure = c.secure ? 'TRUE' : 'FALSE'
    const exp = c.expirationDate ? Math.floor(c.expirationDate) : 0
    return [
      c.domain ?? '',
      flag,
      c.path ?? '/',
      secure,
      exp,
      c.name ?? '',
      c.value ?? '',
    ].join('\t')
  })
  copy('# Netscape HTTP Cookie File\n' + lines.join('\n'), 'netscape')
}
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center gap-2">
      <Field
        v-model="filter"
        placeholder="Filter by name, domain, value"
        size="sm"
        :block="false"
      />
      <span class="text-[11px] text-fg-faint mono ml-auto">
        {{ filtered.length }} / {{ list.length }} cookies
      </span>
      <Btn size="sm" variant="ghost" @click="exportJson">
        {{ copied === 'export' ? '✓ copied' : 'Copy JSON' }}
      </Btn>
      <Btn size="sm" variant="ghost" @click="exportNetscape">
        {{ copied === 'netscape' ? '✓ copied' : 'Copy Netscape' }}
      </Btn>
      <Btn size="sm" variant="ghost" :loading="loading" @click="load">Reload</Btn>
    </div>

    <div class="surface overflow-hidden">
      <table class="w-full text-left">
        <thead>
          <tr
            class="text-[10px] uppercase tracking-wider text-fg-faint border-b border-border-subtle bg-bg-overlay"
          >
            <th class="px-3 py-2">Name</th>
            <th class="px-3 py-2">Domain</th>
            <th class="px-3 py-2">Value</th>
            <th class="px-3 py-2 w-[110px]">Flags</th>
            <th class="px-3 py-2 w-[60px]"></th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(c, i) in filtered"
            :key="i"
            class="border-b border-border-subtle last:border-b-0 hover:bg-bg-hover/60"
          >
            <td class="px-3 py-1.5 mono text-[12px] truncate max-w-[180px]" :title="c.name">
              {{ c.name }}
            </td>
            <td class="px-3 py-1.5 text-[12px] truncate max-w-[200px]" :title="c.domain">
              {{ c.domain }}
            </td>
            <td class="px-3 py-1.5 mono text-[11px] text-fg-muted truncate max-w-[400px]" :title="c.value">
              {{ c.value }}
            </td>
            <td class="px-3 py-1.5">
              <span class="flex flex-wrap gap-1">
                <span v-if="c.secure" class="chip">secure</span>
                <span v-if="c.httpOnly" class="chip">httpOnly</span>
                <span v-if="c.sameSite" class="chip">{{ c.sameSite }}</span>
              </span>
            </td>
            <td class="px-3 py-1.5">
              <button
                class="text-fg-faint hover:text-accent text-[11px]"
                @click="copy(c.value ?? '', 'v-' + i)"
              >
                {{ copied === 'v-' + i ? '✓' : 'copy' }}
              </button>
            </td>
          </tr>
          <tr v-if="!loading && filtered.length === 0">
            <td colspan="5" class="text-center py-10 text-fg-faint text-[12px]">
              No cookies match.
            </td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
