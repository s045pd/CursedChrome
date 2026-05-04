<script setup lang="ts">
import { onBeforeUnmount, ref, watch } from 'vue'
import WaveSurfer from 'wavesurfer.js'
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
const fallbackUrl = ref<string | null>(null)
let wave: WaveSurfer | null = null
const waveContainer = ref<HTMLElement | null>(null)
const fallbackAudio = ref<HTMLAudioElement | null>(null)

async function load(): Promise<void> {
  loading.value = true
  try {
    sessions.value = (await media.audioSessions(props.botId)) ?? []
  } finally {
    loading.value = false
  }
}
watch(() => props.botId, load, { immediate: true })

function disposeWave(): void {
  wave?.destroy()
  wave = null
  fallbackUrl.value = null
}

async function play(s: AudioSession): Promise<void> {
  if (playing.value === s.session_id) {
    if (wave) wave.playPause()
    else fallbackAudio.value?.pause()
    playing.value = null
    return
  }

  if (!waveContainer.value) return
  disposeWave()
  playing.value = s.session_id
  audioLoading.value = true
  fallbackUrl.value = null
  const url = media.audioSessionURL(s.session_id)

  wave = WaveSurfer.create({
    container: waveContainer.value,
    waveColor: 'oklch(45% 0.012 240)',
    progressColor: 'oklch(72% 0.16 220)',
    cursorColor: 'oklch(80% 0.18 220)',
    height: 64,
    barWidth: 2,
    barRadius: 1,
    url,
  })
  wave.on('ready', () => {
    audioLoading.value = false
    wave?.play()
  })
  wave.on('finish', () => {
    playing.value = null
  })
  wave.on('error', () => {
    disposeWave()
    audioLoading.value = false
    fallbackUrl.value = url
    playing.value = s.session_id
  })
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
  await remote.stopAudio(props.botId)
  recording.value = false
  await load()
}

onBeforeUnmount(disposeWave)
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
      <div ref="waveContainer" />
      <div v-if="audioLoading" class="absolute inset-0 grid place-items-center text-[11px] text-fg-faint">
        Loading audio…
      </div>
      <div v-else-if="!playing && !fallbackUrl" class="absolute inset-0 grid place-items-center text-[11px] text-fg-faint">
        Select a session to play
      </div>
      <audio
        v-if="fallbackUrl"
        ref="fallbackAudio"
        :src="fallbackUrl"
        controls
        autoplay
        class="w-full mt-1"
        @ended="playing = null; fallbackUrl = null"
      />
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
