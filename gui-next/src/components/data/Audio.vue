<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import { media, remote } from '@/api/endpoints'
import type { AudioSession } from '@/types/api'
import Btn from '@/components/ui/Btn.vue'
import { formatDate } from '@/composables/useTime'

const props = defineProps<{ botId: string }>()
const sessions = ref<AudioSession[]>([])
const loading = ref(false)
const playing = ref<string | null>(null)
const recording = ref(false)
const audioLoading = ref(false)

const audioEl = ref<HTMLAudioElement | null>(null)
let chunkQueue: string[] = []
let chunkIndex = 0

async function load(): Promise<void> {
  loading.value = true
  try {
    sessions.value = (await media.audioSessions(props.botId)) ?? []
  } finally {
    loading.value = false
  }
}
watch(() => props.botId, load, { immediate: true })

function stopPlayback(): void {
  if (audioEl.value) {
    audioEl.value.pause()
    audioEl.value.src = ''
  }
  chunkQueue = []
  chunkIndex = 0
  playing.value = null
}

function playNextChunk(): void {
  if (chunkIndex >= chunkQueue.length) {
    stopPlayback()
    return
  }
  if (!audioEl.value) return
  audioEl.value.src = media.audioChunkURL(chunkQueue[chunkIndex])
  audioEl.value.play().catch(() => {
    stopPlayback()
  })
  chunkIndex++
}

async function play(s: AudioSession): Promise<void> {
  if (playing.value === s.session_id) {
    stopPlayback()
    return
  }

  stopPlayback()
  playing.value = s.session_id
  audioLoading.value = true

  try {
    const chunks = await media.audioSessionChunks(s.session_id)
    if (!chunks || chunks.length === 0) {
      stopPlayback()
      return
    }
    chunkQueue = chunks.map(c => c.id)
    chunkIndex = 0
    audioLoading.value = false
    playNextChunk()
  } catch {
    audioLoading.value = false
    stopPlayback()
  }
}

function onAudioEnded(): void {
  playNextChunk()
}

async function startRec(): Promise<void> {
  recording.value = true
  try {
    await remote.startAudio(props.botId)
  } catch {
    recording.value = false
  }
}

async function stopRec(): Promise<void> {
  try {
    await remote.stopAudio(props.botId)
  } catch {
    // ignore — bot may have disconnected
  }
  recording.value = false
  await load()
}

onBeforeUnmount(stopPlayback)
</script>

<template>
  <div class="space-y-3">
    <div class="flex items-center gap-2">
      <Btn
        v-if="!recording"
        size="sm"
        variant="success"
        @click="startRec"
      >
        ● Start recording
      </Btn>
      <Btn
        v-else
        size="sm"
        variant="danger"
        @click="stopRec"
      >
        ■ Stop recording
      </Btn>
      <span class="text-[11px] text-fg-faint mono ml-auto">
        {{ sessions.length }} sessions
      </span>
      <Btn size="sm" variant="ghost" :loading="loading" @click="load">Reload</Btn>
    </div>

    <div class="surface px-3 py-3 min-h-[80px] relative">
      <audio
        ref="audioEl"
        class="w-full"
        controls
        :class="{ 'opacity-0 pointer-events-none': !playing }"
        @ended="onAudioEnded"
      />
      <div v-if="audioLoading" class="absolute inset-0 grid place-items-center text-[11px] text-fg-faint">
        Loading audio…
      </div>
      <div v-else-if="!playing" class="absolute inset-0 grid place-items-center text-[11px] text-fg-faint">
        Select a session to play
      </div>
      <div v-if="playing && chunkQueue.length > 1" class="text-center text-[10px] text-fg-faint mono mt-1">
        chunk {{ chunkIndex }} / {{ chunkQueue.length }}
      </div>
    </div>

    <div class="surface divide-y divide-border-subtle">
      <div
        v-for="s in sessions"
        :key="s.session_id"
        class="px-3 py-2 hover:bg-bg-hover/60 flex items-center gap-3"
      >
        <button
          class="size-8 grid place-items-center rounded bg-accent-soft text-accent border border-accent/30 hover:bg-accent/20"
          :aria-label="playing === s.session_id ? 'Playing' : 'Play'"
          @click="play(s)"
        >
          <span v-if="playing === s.session_id">▮▮</span>
          <span v-else>▶</span>
        </button>
        <div class="flex-1">
          <div class="text-[12px] mono truncate">{{ s.session_id }}</div>
          <div class="text-[10px] text-fg-faint mono">
            {{ formatDate(s.start_time) }} → {{ formatDate(s.end_time) }}
          </div>
        </div>
        <span class="chip mono">{{ s.chunk_count }} chunks</span>
        <a
          :href="media.audioSessionURL(s.session_id)"
          download
          class="text-[11px] text-fg-muted hover:text-accent"
        >
          download
        </a>
      </div>
      <div v-if="!loading && sessions.length === 0" class="text-center py-10 text-fg-faint text-[12px]">
        No audio sessions yet. Start a recording to capture.
      </div>
    </div>
  </div>
</template>
