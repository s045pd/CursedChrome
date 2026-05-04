<script setup lang="ts">
import { ref, watch, computed } from 'vue'
import { media } from '@/api/endpoints'
import type { ScreenshotEntry } from '@/types/api'
import Btn from '@/components/ui/Btn.vue'
import { formatDate } from '@/composables/useTime'

const props = defineProps<{ botId: string }>()
const list = ref<ScreenshotEntry[]>([])
const loading = ref(false)
const lightbox = ref<ScreenshotEntry | null>(null)
const limit = ref(60)

function imgURL(s: ScreenshotEntry): string {
  return s.ImageData || media.screenshotImageURL(s.ID)
}

async function load(): Promise<void> {
  loading.value = true
  try {
    list.value = (await media.screenshots(props.botId, limit.value, 0)) ?? []
  } finally {
    loading.value = false
  }
}

watch(() => props.botId, load, { immediate: true })

const grouped = computed(() => {
  const map = new Map<string, ScreenshotEntry[]>()
  for (const s of list.value) {
    const day = new Date(s.Timestamp).toLocaleDateString()
    if (!map.has(day)) map.set(day, [])
    map.get(day)!.push(s)
  }
  return Array.from(map.entries())
})
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center gap-2">
      <span class="text-[11px] text-fg-faint mono ml-auto">
        {{ list.length }} screenshots
      </span>
      <select
        v-model.number="limit"
        class="h-7 text-[11px] px-2 rounded bg-bg-overlay border border-border-subtle"
        @change="load"
      >
        <option :value="20">last 20</option>
        <option :value="60">last 60</option>
        <option :value="200">last 200</option>
        <option :value="500">last 500</option>
      </select>
      <Btn size="sm" variant="ghost" :loading="loading" @click="load">Reload</Btn>
    </div>

    <div v-if="!loading && list.length === 0" class="text-center py-12 text-fg-faint text-[12px]">
      No screenshots captured yet for this bot.
    </div>

    <section v-for="[day, items] in grouped" :key="day" class="space-y-2">
      <h4 class="text-[10px] uppercase tracking-wider text-fg-faint mono">
        {{ day }} <span class="ml-1.5">{{ items.length }}</span>
      </h4>
      <div class="grid grid-cols-4 gap-2">
        <button
          v-for="s in items"
          :key="s.ID"
          class="group relative surface overflow-hidden aspect-video focus:outline-none focus-visible:ring-2 focus-visible:ring-accent"
          @click="lightbox = s"
        >
          <img
            :src="imgURL(s)"
            :alt="s.Title || ''"
            loading="lazy"
            class="w-full h-full object-cover transition-transform group-hover:scale-[1.02]"
          />
          <div
            class="absolute inset-x-0 bottom-0 px-2 py-1.5 bg-gradient-to-t from-black/85 to-transparent text-[10px] mono text-white/90 truncate text-left"
          >
            {{ s.Title || s.URL || '—' }}
          </div>
          <div
            class="absolute top-1.5 right-1.5 chip bg-black/60 text-white/90 border-white/10"
          >
            {{ new Date(s.Timestamp).toLocaleTimeString(undefined, { hour12: false }) }}
          </div>
        </button>
      </div>
    </section>

    <!-- lightbox -->
    <Teleport to="body">
      <div
        v-if="lightbox"
        class="fixed inset-0 z-50 bg-black/85 backdrop-blur-sm grid place-items-center p-6"
        @click="lightbox = null"
      >
        <div class="max-w-[1200px] w-full" @click.stop>
          <img
            :src="imgURL(lightbox)"
            :alt="lightbox.Title || ''"
            class="w-full h-auto rounded-md border border-border-subtle"
          />
          <div class="mt-3 flex items-center justify-between">
            <div class="space-y-0.5">
              <p class="text-[13px]">{{ lightbox.Title || '(no title)' }}</p>
              <p class="text-[11px] text-fg-faint mono truncate max-w-[700px]">
                {{ lightbox.URL }}
              </p>
              <p class="text-[10px] text-fg-faint mono">
                {{ formatDate(lightbox.Timestamp) }}
                <span v-if="lightbox.SessionID" class="ml-2">session: {{ lightbox.SessionID }}</span>
              </p>
            </div>
            <Btn variant="subtle" @click.stop="lightbox = null">Close</Btn>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
