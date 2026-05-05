<script setup lang="ts">
import { ref, watch, computed, onBeforeUnmount } from 'vue'
import { media } from '@/api/endpoints'
import type { ScreenshotEntry } from '@/types/api'
import Btn from '@/components/ui/Btn.vue'
import { formatDate } from '@/composables/useTime'

const props = defineProps<{ botId: string }>()
const list = ref<ScreenshotEntry[]>([])
const loading = ref(false)
const lightbox = ref<ScreenshotEntry | null>(null)
const limit = ref(60)

// Playback state
const playing = ref(false)
const playIndex = ref(0)
const playSpeed = ref(1)
let playTimer: ReturnType<typeof setTimeout> | null = null
const speeds = [0.5, 1, 2, 5, 10]

const chronological = computed(() => [...list.value].filter(s => s.HasImage).reverse())

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

// Playback controls
function startPlayback(fromIndex = 0) {
  if (chronological.value.length === 0) return
  playIndex.value = fromIndex
  playing.value = true
  scheduleNext()
}

function stopPlayback() {
  playing.value = false
  if (playTimer) { clearTimeout(playTimer); playTimer = null }
}

function togglePlayback() {
  if (playing.value) stopPlayback()
  else startPlayback(playIndex.value)
}

function scheduleNext() {
  if (playTimer) clearTimeout(playTimer)
  if (!playing.value) return
  const interval = 1000 / playSpeed.value
  playTimer = setTimeout(() => {
    if (!playing.value) return
    if (playIndex.value < chronological.value.length - 1) {
      playIndex.value++
      scheduleNext()
    } else {
      stopPlayback()
    }
  }, interval)
}

function setSpeed(s: number) {
  playSpeed.value = s
  if (playing.value) scheduleNext()
}

function seekTo(idx: number) {
  playIndex.value = Math.max(0, Math.min(idx, chronological.value.length - 1))
  if (playing.value) scheduleNext()
}

function stepPrev() {
  stopPlayback()
  seekTo(playIndex.value - 1)
}

function stepNext() {
  stopPlayback()
  seekTo(playIndex.value + 1)
}

const currentShot = computed(() => chronological.value[playIndex.value] ?? null)
const showPlayer = ref(false)

function openPlayer() {
  if (chronological.value.length === 0) return
  showPlayer.value = true
  playIndex.value = 0
}

function handleClosePlayer() {
  stopPlayback()
  showPlayer.value = false
  playIndex.value = 0
}

function handleKeydown(e: KeyboardEvent) {
  if (!showPlayer.value) return
  if (e.key === 'Escape') { handleClosePlayer(); return }
  if (e.key === ' ') { e.preventDefault(); togglePlayback(); return }
  if (e.key === 'ArrowLeft') { stepPrev(); return }
  if (e.key === 'ArrowRight') { stepNext(); return }
  if (e.key === 'ArrowUp') {
    const i = speeds.indexOf(playSpeed.value)
    if (i < speeds.length - 1) setSpeed(speeds[i + 1])
    return
  }
  if (e.key === 'ArrowDown') {
    const i = speeds.indexOf(playSpeed.value)
    if (i > 0) setSpeed(speeds[i - 1])
  }
}

if (typeof window !== 'undefined') {
  window.addEventListener('keydown', handleKeydown)
}
onBeforeUnmount(() => {
  stopPlayback()
  window.removeEventListener('keydown', handleKeydown)
})

const progressPct = computed(() =>
  chronological.value.length <= 1 ? 0 : (playIndex.value / (chronological.value.length - 1)) * 100
)
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center gap-2">
      <Btn
        v-if="list.length > 1"
        size="sm"
        variant="primary"
        @click="openPlayer"
      >
        Playback
      </Btn>
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

    <!-- single image lightbox -->
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

    <!-- playback player -->
    <Teleport to="body">
      <div
        v-if="showPlayer"
        class="fixed inset-0 z-50 bg-black/95 flex flex-col"
        @click.self="handleClosePlayer"
      >
        <!-- image area -->
        <div class="flex-1 flex items-center justify-center p-4 min-h-0">
          <img
            v-if="currentShot"
            :src="imgURL(currentShot)"
            :alt="currentShot.Title || ''"
            class="max-w-full max-h-full object-contain rounded"
          />
        </div>

        <!-- controls -->
        <div class="shrink-0 bg-black/80 border-t border-white/10 px-6 py-3 space-y-2">
          <!-- progress bar -->
          <div
            class="relative h-1.5 bg-white/10 rounded-full cursor-pointer group"
            @click.stop="(e: MouseEvent) => {
              const rect = (e.currentTarget as HTMLElement).getBoundingClientRect()
              const pct = Math.max(0, Math.min(1, (e.clientX - rect.left) / rect.width))
              seekTo(Math.round(pct * (chronological.length - 1)))
            }"
          >
            <div
              class="absolute inset-y-0 left-0 bg-accent rounded-full transition-[width] duration-100"
              :style="{ width: progressPct + '%' }"
            />
            <div
              class="absolute top-1/2 -translate-y-1/2 w-3 h-3 bg-accent rounded-full shadow opacity-0 group-hover:opacity-100 transition-opacity"
              :style="{ left: `calc(${progressPct}% - 6px)` }"
            />
          </div>

          <div class="flex items-center gap-4">
            <!-- transport -->
            <div class="flex items-center gap-1">
              <button
                class="w-8 h-8 flex items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 transition-colors"
                :disabled="playIndex === 0"
                :class="{ 'opacity-30 cursor-not-allowed': playIndex === 0 }"
                @click.stop="seekTo(0)"
                title="First"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M6 6h2v12H6zm3.5 6 8.5 6V6z"/></svg>
              </button>
              <button
                class="w-8 h-8 flex items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 transition-colors"
                :disabled="playIndex === 0"
                :class="{ 'opacity-30 cursor-not-allowed': playIndex === 0 }"
                @click.stop="stepPrev"
                title="Previous (Left arrow)"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M15.41 7.41 14 6l-6 6 6 6 1.41-1.41L10.83 12z"/></svg>
              </button>
              <button
                class="w-10 h-10 flex items-center justify-center rounded-full text-white bg-accent hover:bg-accent-strong transition-colors"
                @click.stop="togglePlayback"
                :title="playing ? 'Pause (Space)' : 'Play (Space)'"
              >
                <svg v-if="!playing" class="w-5 h-5 ml-0.5" fill="currentColor" viewBox="0 0 24 24"><path d="M8 5v14l11-7z"/></svg>
                <svg v-else class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M6 19h4V5H6zm8-14v14h4V5z"/></svg>
              </button>
              <button
                class="w-8 h-8 flex items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 transition-colors"
                :disabled="playIndex >= chronological.length - 1"
                :class="{ 'opacity-30 cursor-not-allowed': playIndex >= chronological.length - 1 }"
                @click.stop="stepNext"
                title="Next (Right arrow)"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M10 6 8.59 7.41 13.17 12l-4.58 4.59L10 18l6-6z"/></svg>
              </button>
              <button
                class="w-8 h-8 flex items-center justify-center rounded text-white/70 hover:text-white hover:bg-white/10 transition-colors"
                :disabled="playIndex >= chronological.length - 1"
                :class="{ 'opacity-30 cursor-not-allowed': playIndex >= chronological.length - 1 }"
                @click.stop="seekTo(chronological.length - 1)"
                title="Last"
              >
                <svg class="w-4 h-4" fill="currentColor" viewBox="0 0 24 24"><path d="M6 18l8.5-6L6 6v12zm2 0h2V6h-2v12z" transform="scale(-1,1) translate(-24,0)"/></svg>
              </button>
            </div>

            <!-- speed -->
            <div class="flex items-center gap-1 ml-2">
              <button
                v-for="s in speeds"
                :key="s"
                class="h-6 px-2 text-[10px] mono rounded transition-colors"
                :class="playSpeed === s
                  ? 'bg-accent text-bg-base'
                  : 'text-white/50 hover:text-white hover:bg-white/10'"
                @click.stop="setSpeed(s)"
              >
                {{ s }}x
              </button>
            </div>

            <!-- frame info -->
            <div class="ml-auto flex items-center gap-4 text-[11px] mono text-white/60">
              <span v-if="currentShot">
                {{ new Date(currentShot.Timestamp).toLocaleTimeString(undefined, { hour12: false }) }}
              </span>
              <span>{{ playIndex + 1 }} / {{ chronological.length }}</span>
              <span v-if="currentShot" class="max-w-[300px] truncate text-white/40">
                {{ currentShot.Title || currentShot.URL || '' }}
              </span>
            </div>

            <!-- close -->
            <button
              class="w-8 h-8 flex items-center justify-center rounded text-white/50 hover:text-white hover:bg-white/10 transition-colors ml-2"
              @click.stop="handleClosePlayer"
              title="Close (Esc)"
            >
              <svg class="w-5 h-5" fill="currentColor" viewBox="0 0 24 24"><path d="M19 6.41 17.59 5 12 10.59 6.41 5 5 6.41 10.59 12 5 17.59 6.41 19 12 13.41 17.59 19 19 17.59 13.41 12z"/></svg>
            </button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
