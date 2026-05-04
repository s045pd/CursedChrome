<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { bots as botsApi } from '@/api/endpoints'
import Field from '@/components/ui/Field.vue'
import Btn from '@/components/ui/Btn.vue'

// Untyped row container — callers cast inside their slot. We don't use
// generics here because Vue 3.5 generic SFCs interact poorly with both
// vue-tsc and the template type checker for slot scoped variables, and
// the search keys are looked up by string anyway.
type Row = Record<string, unknown>

interface Props {
  botId: string
  field: string
  searchKeys: string[]
  emptyMessage?: string
}
const props = withDefaults(defineProps<Props>(), {
  emptyMessage: 'Nothing here yet.',
})

const list = ref<Row[]>([])
const filter = ref('')
const loading = ref(false)

async function load(): Promise<void> {
  loading.value = true
  try {
    list.value = ((await botsApi.field<Row[]>(props.botId, props.field)) ?? []) as Row[]
  } finally {
    loading.value = false
  }
}

watch(() => [props.botId, props.field], load, { immediate: true })

const filtered = computed(() => {
  const q = filter.value.trim().toLowerCase()
  if (!q) return list.value
  return list.value.filter((row) => {
    for (const k of props.searchKeys) {
      const v = row[k]
      if (typeof v === 'string' && v.toLowerCase().includes(q)) return true
    }
    return false
  })
})
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center gap-2">
      <Field
        v-model="filter"
        placeholder="Filter"
        size="sm"
        :block="false"
      />
      <span class="text-[11px] text-fg-faint mono ml-auto">
        {{ filtered.length }} / {{ list.length }}
      </span>
      <Btn size="sm" variant="ghost" :loading="loading" @click="load">Reload</Btn>
    </div>

    <div v-if="!loading && filtered.length === 0" class="text-center py-10 text-fg-faint text-[12px]">
      {{ emptyMessage }}
    </div>

    <div v-else class="surface divide-y divide-border-subtle">
      <slot v-for="(row, i) in filtered" :key="i" :row="row" :i="i" />
    </div>
  </div>
</template>
