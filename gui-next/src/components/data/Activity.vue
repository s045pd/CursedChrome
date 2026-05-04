<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { bots as botsApi } from '@/api/endpoints'
import Btn from '@/components/ui/Btn.vue'

interface ActivityPoint { start: number | string; end: number | string }

const props = defineProps<{ botId: string }>()
const list = ref<ActivityPoint[]>([])
const loading = ref(false)
const days = ref(30)

async function load(): Promise<void> {
  loading.value = true
  try {
    list.value = (await botsApi.field<ActivityPoint[]>(props.botId, 'activity')) ?? []
  } finally {
    loading.value = false
  }
}
watch(() => props.botId, load, { immediate: true })

const dayMs = 24 * 60 * 60 * 1000

const buckets = computed(() => {
  const now = Date.now()
  const start = now - days.value * dayMs
  // 30-day grid bucketed per hour
  const cellMs = 60 * 60 * 1000
  const cells = Math.ceil((now - start) / cellMs)
  const arr = new Float32Array(cells)
  for (const p of list.value) {
    const s = typeof p.start === 'string' ? Date.parse(p.start) : p.start
    const e = typeof p.end === 'string' ? Date.parse(p.end) : p.end
    if (!s || !e || e < start) continue
    const a = Math.max(start, s)
    const b = Math.min(now, e)
    if (b <= a) continue
    const ai = Math.floor((a - start) / cellMs)
    const bi = Math.min(cells - 1, Math.floor((b - start) / cellMs))
    for (let i = ai; i <= bi; i++) arr[i] += 1
  }
  return { arr, start, cellMs }
})

const maxIntensity = computed(() => {
  let m = 0
  for (const v of buckets.value.arr) if (v > m) m = v
  return m
})

function color(v: number, max: number): string {
  if (v <= 0) return 'oklch(28% 0.01 240)'
  const t = max ? v / max : 0
  // green→accent gradient by intensity
  const l = 30 + 50 * t
  return `oklch(${l}% 0.18 ${145 + 80 * t})`
}

function tooltip(i: number): string {
  const t = buckets.value.start + i * buckets.value.cellMs
  return `${new Date(t).toLocaleString()} — ${buckets.value.arr[i].toFixed(0)} active`
}
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center gap-2">
      <select
        v-model.number="days"
        class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle"
      >
        <option :value="7">7d</option>
        <option :value="14">14d</option>
        <option :value="30">30d</option>
        <option :value="60">60d</option>
      </select>
      <span class="text-[11px] text-fg-faint mono ml-auto">
        {{ list.length }} sessions tracked
      </span>
      <Btn size="sm" variant="ghost" :loading="loading" @click="load">Reload</Btn>
    </div>

    <div class="surface px-4 py-4 overflow-x-auto">
      <div
        class="grid gap-[3px]"
        :style="{
          gridTemplateColumns: `repeat(${days * 24}, 6px)`,
          gridAutoRows: '14px',
        }"
      >
        <div
          v-for="(v, i) in buckets.arr"
          :key="i"
          class="rounded-sm"
          :style="{ background: color(v, maxIntensity) }"
          :title="tooltip(i)"
        />
      </div>
      <div class="flex items-center justify-end gap-1.5 mt-3 text-[10px] text-fg-faint mono">
        <span>less</span>
        <span class="size-3 rounded-sm" style="background: oklch(28% 0.01 240)" />
        <span class="size-3 rounded-sm" style="background: oklch(50% 0.12 145)" />
        <span class="size-3 rounded-sm" style="background: oklch(60% 0.16 200)" />
        <span class="size-3 rounded-sm" style="background: oklch(80% 0.18 225)" />
        <span>more</span>
      </div>
    </div>

    <div v-if="!loading && list.length === 0" class="text-center py-6 text-fg-faint text-[12px]">
      No activity recorded for this bot yet.
    </div>
  </div>
</template>
